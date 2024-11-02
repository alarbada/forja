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

	parts := strings.Split(fullName, ".")

	// Ensure we have at least two parts (package and function name)
	if len(parts) < 2 {
		panic("Invalid function name format")
	}

	packageName := parts[len(parts)-2]
	{
		parts := strings.Split(packageName, "/")
		if len(parts) > 0 {
			packageName = parts[len(parts)-1]
		}
	}

	handlerName := parts[len(parts)-1]

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

func (th *TypedHandlers) GenerateTypescriptClient() string {
	var sb strings.Builder

	// Generate ApiError type and ApiResponse type
	sb.WriteString(`
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

	// Generate callbacks

	for packageName, handlers := range packages {
		for handlerName, handler := range handlers {
			if handler.handlerType.NumIn() < 2 || handler.handlerType.NumOut() < 1 {
				fmt.Printf("Warning: unexpected handler signature for %s.%s\n", packageName, handlerName)
				continue
			}

			// input
			inputType := handler.handlerType.In(1)
			var inputTypeName string
			if inputType.Kind() == reflect.Struct && inputType.NumField() != 0 {
				inputTypeName = camelcaseNames(packageName, handlerName, "Input")
				fmt.Fprintf(&sb,
					"export type %s = %s\n\n",
					inputTypeName, generateTypescriptType(inputType),
				)
			} else {
				handler.isInputEmpty = true
			}

			// output
			outputType := handler.handlerType.Out(0)
			outputTypeName := camelcaseNames(packageName, handlerName, "Output")
			fmt.Fprintf(&sb,
				"export type %s = %s\n\n",
				outputTypeName, generateTypescriptType(outputType),
			)

			// handler
			if inputTypeName != "" {
				handlerName := camelcaseNames(packageName, handlerName, "Handler")
				fmt.Fprintf(&sb,
					"type %s = (params: %s) => Promise<ApiResponse<%s>>\n\n",
					handlerName, inputTypeName, outputTypeName,
				)
			} else {
				handlerName := camelcaseNames(packageName, handlerName, "Handler")
				fmt.Fprintf(&sb,
					"type %s = () => Promise<ApiResponse<%s>>\n\n",
					handlerName, outputTypeName,
				)
			}
		}
	}

	// Generate ApiClient interface
	sb.WriteString("export interface ApiClient {\n")
	for packageName, handlers := range packages {
		sb.WriteString(fmt.Sprintf("  %s: {\n", packageName))
		for handlerName, handler := range handlers {
			handlerType := handler.handlerType
			if handlerType.NumIn() < 2 || handlerType.NumOut() < 1 {
				fmt.Printf("Warning: unexpected handler signature for %s.%s\n", packageName, handlerName)
				continue
			}

			fmt.Fprintf(&sb,
				"  %s: %s\n",
				handlerName, camelcaseNames(packageName, handlerName, "Handler"),
			)
		}
		sb.WriteString("  }\n")
	}
	sb.WriteString("}\n")

	// Generate createApiClient function
	sb.WriteString(`
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
		sb.WriteString(fmt.Sprintf("    %s: {\n", packageName))
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

			sb.WriteString(callback)
		}
		sb.WriteString("    },\n")
	}

	sb.WriteString(`  }
  return client
}
`)

	return sb.String()
}

func generateTypescriptType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Struct:
		var sb strings.Builder
		sb.WriteString("{\n")
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldName, optional := parseJSONTag(field.Tag.Get("json"))
			if fieldName == "-" {
				continue // Skip fields with json:"-"
			}
			if fieldName == "" {
				fieldName = field.Name
			}
			fieldType := generateTypescriptType(field.Type)
			if optional {
				sb.WriteString(fmt.Sprintf("    %s?: %s\n", fieldName, fieldType))
			} else {
				sb.WriteString(fmt.Sprintf("    %s: %s\n", fieldName, fieldType))
			}
		}
		sb.WriteString("  }")
		return sb.String()
	case reflect.Slice, reflect.Array:
		return generateTypescriptType(t.Elem()) + "[]"
	case reflect.Ptr:
		return generateTypescriptType(t.Elem())
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Interface:
		return "any"
	default:
		return "unknown"
	}
}

func parseJSONTag(tag string) (string, bool) {
	if tag == "" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	name := parts[0]
	if name == "" {
		return "", false
	}
	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			return name, true
		}
	}
	return name, false
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
