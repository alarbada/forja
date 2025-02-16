package forja

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
)

type Handler[P any, R any] func(c echo.Context, params P) (R, error)

type Router interface {
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// Forja is the main struct that handles type information for every given handler.
// To add handlers to it use
type Forja struct {
	router   Router
	handlers map[string]reflect.Type // "package.handler" -> Handler type

	config Config
}

func NewForja(router Router) *Forja {
	return NewForjaWithConfig(router, Config{Path: "/"})
}

type Config struct {
	// OnErr, if not nil will be called if the handler responds with an error
	OnErr func(error)

	// Path is where the handlers will be mounted to. By default the handlers are
	// mounted at the root path '/'.
	// For example, with empty path:
	// /main_myHandler
	// With "/api" as the given path:
	// /api/main_myHandler
	// No further parsing is done to the path. Make sure that the path has a "/" path prefixed.
	Path string
}

func NewForjaWithConfig(router Router, config Config) *Forja {
	if config.Path == "" {
		config.Path = "/"
	}

	th := &Forja{
		router:   router,
		config:   config,
		handlers: make(map[string]reflect.Type),
	}
	return th
}

func cleanHandlerName(handlerName string) string {
	handlerName = strings.ReplaceAll(handlerName, "(", "")
	handlerName = strings.ReplaceAll(handlerName, ")", "")
	handlerName = strings.ReplaceAll(handlerName, "*", "")

	return handlerName
}

func AddHandler[P any, R any](th *Forja, handler Handler[P, R]) {
	handlerFunc := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
	fullName := handlerFunc.Name()

	// Explanation for this "-fm" thingy
	// https://github.com/golang/go/issues/52809#issuecomment-1122696583
	fullName = strings.TrimSuffix(fullName, "-fm")

	parts := strings.Split(fullName, "/")
	parts = strings.Split(parts[len(parts)-1], ".")

	packageName := parts[0]
	handlerName := strings.Join(parts[1:], "_")

	// if handler is a method of a struct pointer, we need to clean it, as it
	// will come in the form of:
	// (*Mystruct)_methodName
	handlerName = cleanHandlerName(handlerName)

	path := fmt.Sprintf("/%s.%s", packageName, handlerName)
	fullPath := fmt.Sprintf("%s.%s", packageName, handlerName)

	th.handlers[fullPath] = reflect.TypeOf(handler)

	th.router.POST(path, func(c echo.Context) error {
		var params P
		if err := c.Bind(&params); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		result, err := handler(c, params)
		if err != nil {
			if th.config.OnErr != nil {
				th.config.OnErr(err)
			}

			return c.JSON(400, map[string]string{
				"message": err.Error(),
			})
		}

		return c.JSON(200, result)
	})
}

func (th *Forja) WriteTsClient(path string) error {
	generated := th.GenerateTypescriptClient()
	return os.WriteFile(path, []byte(generated), 0644)
}

func (th *Forja) GenerateTypescriptClient() string {
	output := new(strings.Builder)

	// Generate ApiError type and ApiResponse type
	output.WriteString(`
export interface ApiError {
  message: string
  statusCode?: number
}
export type ApiResponse<T> =
  | { data: T; error: null }
  | { data: null; error: ApiError }

`)

	type PackageName = string
	type HandlerName = string
	type Handler struct {
		isInputEmpty bool
		handlerType  reflect.Type
	}

	// Handler is a pointer so that we can update isInputEmpty later.
	type Packages = map[PackageName]map[HandlerName]*Handler

	packages := make(Packages)
	for fullPath, handlerType := range th.handlers {
		parts := strings.Split(fullPath, ".")
		if len(parts) != 2 {
			continue
		}
		packageName, handlerName := parts[0], parts[1]
		packageParts := strings.Split(packageName, "/")
		simplifiedPackageName := packageParts[len(packageParts)-1]
		if packages[simplifiedPackageName] == nil {
			packages[simplifiedPackageName] = make(map[HandlerName]*Handler)
		}
		packages[simplifiedPackageName][handlerName] = &Handler{
			isInputEmpty: false,
			handlerType:  handlerType,
		}
	}

	type HandlerTsType = string
	apiClientTsDefinitions := map[PackageName]map[HandlerName]HandlerTsType{}

	for packageName, handlers := range packages {
		for handlerName, handler := range handlers {
			_, _, _ = packageName, handlerName, handler.handlerType
			inputType := handler.handlerType.In(1)
			outputType := handler.handlerType.Out(0).Elem()
			inputTypeName := fillTypeDefinitions(inputType)
			outputTypeName := fillTypeDefinitions(outputType)

			handlerTsName := camelcaseNames(packageName, handlerName, "Handler")

			isInputEmptyStruct := inputType.Kind() == reflect.Struct && inputType.NumField() == 0
			if isInputEmptyStruct {
				packages[packageName][handlerName].isInputEmpty = true
				fmt.Fprintf(output,
					"type %s = () => Promise<ApiResponse<%s>>\n",
					handlerTsName, outputTypeName)
			} else {
				fmt.Fprintf(output,
					"type %s = (params: %s) => Promise<ApiResponse<%s>>\n",
					handlerTsName, inputTypeName, outputTypeName)
			}

			if apiClientTsDefinitions[packageName] == nil {
				apiClientTsDefinitions[packageName] = map[string]string{}
			}

			apiClientTsDefinitions[packageName][handlerName] = handlerTsName
		}
	}

	fmt.Fprintln(output, "type ApiClient = {")
	for packageName, packageTypeDef := range apiClientTsDefinitions {
		fmt.Fprintln(output, "  ", packageName, ": {")
		for handlerName, handlerTypeName := range packageTypeDef {
			fmt.Fprintln(output, "    ", handlerName, ": ", handlerTypeName, ",")
		}
		fmt.Fprintln(output, "  ", "},")
	}
	fmt.Fprintln(output, "}")

	printTypeDefs(output)

	// Generate createApiClient function
	output.WriteString(`
type ApiClientConfig = {
  beforeRequest?: (config: RequestInit) => void | Promise<void>
}

export function createApiClient(
  baseUrl: string,
  config?: ApiClientConfig
): ApiClient {
  async function doFetch(path: string, params?: unknown) {
    try {
	  if (params === undefined) {
	  	params = {}
	  }

      const requestConfig: RequestInit = {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(params ?? {}),
      }

      if (config?.beforeRequest) {
        await config.beforeRequest(requestConfig)
      }

      const response = await fetch(` + "`${baseUrl}/${path}`" + `, requestConfig)
      if (!response.ok) {
        const data = await response.json()
        const message = data.message

        return {
          data: null,
          error: { message, statusCode: response.status },
        }
      }
      const data = await response.json()
      return { data, error: null }
    } catch (error) {
      return {
        data: null,
        error: {
          message:
            error instanceof Error ? error.message : "Unknown error occurred",
        },
      }
    }
  }
  const client: ApiClient = {
`)

	// Generate client methods
	for packageName, handlers := range packages {
		output.WriteString(fmt.Sprintf("    %s: {\n", packageName))
		for handlerName, handler := range handlers {
			var callback string
			if handler.isInputEmpty {
				callback = fmt.Sprintf(
					"      %s: () => doFetch(\"%s.%s\"),\n",
					handlerName, packageName, handlerName)
			} else {
				callback = fmt.Sprintf(
					"      %s: (params) => doFetch(\"%s.%s\", params),\n",
					handlerName, packageName, handlerName)
			}

			output.WriteString(callback)
		}
		output.WriteString("    },\n")
	}

	output.WriteString(`  }
  return client
}
`)

	return output.String()
}

func WriteToFile(th *Forja, filename string) error {
	generated := []byte(th.GenerateTypescriptClient())
	return os.WriteFile(filename, generated, 0644)
}

func camelcaseNames(names ...string) string {
	for i, name := range names {
		if len(name) > 0 {
			names[i] = strings.ToUpper(name[:1]) + name[1:]
		}
	}
	return strings.Join(names, "")
}

// Option is a special type that makes it easy to encode optional values and
// enums (see examples). We do not support generics (yet), so this type is handled
// as a separate entity when "compiling" types to typescript.
type Option[T any] struct {
	IsValid bool
	Value   T
}

func (x *Option[T]) Valid() bool {
	if x == nil {
		return false
	}

	return x.IsValid
}

func (x *Option[T]) MarshalJSON() ([]byte, error) {
	if x.IsValid {
		return json.Marshal(x.Value)
	}
	return json.Marshal(nil)
}

func (x *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		x.IsValid = false
		return nil
	}
	x.IsValid = true
	return json.Unmarshal(data, &x.Value)
}
