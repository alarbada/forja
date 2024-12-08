package forja

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
)

type Handler[P any, R any] func(c echo.Context, params P) (R, error)

type TypedHandlers struct {
	e        *echo.Echo
	handlers map[string]reflect.Type // "package.handler" -> Handler type

	OnErr func(error)
}

func NewTypedHandlersWithErrorHandler(e *echo.Echo, onErr func(error)) *TypedHandlers {
	th := &TypedHandlers{
		e:        e,
		OnErr:    onErr,
		handlers: make(map[string]reflect.Type),
	}
	return th
}

func NewTypedHandlers(e *echo.Echo) *TypedHandlers {
	return NewTypedHandlersWithErrorHandler(e, nil)
}

func AddHandler[P any, R any](th *TypedHandlers, handler Handler[P, R]) {
	handlerFunc := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
	fullName := handlerFunc.Name()

	// Explanation for this "-fm" thingy
	// https://github.com/golang/go/issues/52809#issuecomment-1122696583
	fullName = strings.TrimSuffix(fullName, "-fm")

	parts := strings.Split(fullName, "/")
	parts = strings.Split(parts[len(parts)-1], ".")

	packageName := parts[0]
	handlerName := strings.Join(parts[1:], "_")

	path := fmt.Sprintf("/%s.%s", packageName, handlerName)
	fullPath := fmt.Sprintf("%s.%s", packageName, handlerName)

	th.handlers[fullPath] = reflect.TypeOf(handler)

	th.e.POST(path, func(c echo.Context) error {
		var params P
		if err := c.Bind(&params); err != nil {
			return echo.NewHTTPError(400, err.Error())
		}

		result, err := handler(c, params)
		if err != nil {
			if th.OnErr != nil {
				th.OnErr(err)
			}

			return c.JSON(400, map[string]string{
				"message": err.Error(),
			})
		}

		return c.JSON(200, result)
	})
}

func (th *TypedHandlers) WriteTsClient(path string) error {
	generated := th.GenerateTypescriptClient()
	return os.WriteFile(path, []byte(generated), 0644)
}

func (th *TypedHandlers) GenerateTypescriptClient() string {
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

func WriteToFile(th *TypedHandlers, filename string) error {
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
