package forja

import (
	"fmt"
	"reflect"
	"strings"
)

var typeDefinitions = make(map[string]string)

func printTypeDefs(sb *strings.Builder) {
	for _, typedef := range typeDefinitions {
		fmt.Fprintln(sb, typedef)
	}
}

func getFullTypeName(t reflect.Type) string {
	if t.Name() == "" {
		return ""
	}
	if t.PkgPath() == "" {
		return t.Name()
	}
	// Get just the last part of the package path
	pkgParts := strings.Split(t.PkgPath(), "/")
	pkg := pkgParts[len(pkgParts)-1]
	return fmt.Sprintf("%s_%s", pkg, t.Name())
}

func fillTypeDefinitions(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Struct:
		fullName := getFullTypeName(t)
		if fullName != "" {
			if _, exists := typeDefinitions[fullName]; exists {
				return fullName
			}

			// Generate the type definition
			var fields []string
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				fieldName := field.Name
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" {
					fieldName = strings.Split(jsonTag, ",")[0]
				}

				fieldType := fillTypeDefinitions(field.Type)
				fields = append(fields, fmt.Sprintf("  %s: %s", fieldName, fieldType))
			}

			typeDefinitions[fullName] = fmt.Sprintf("export type %s = {\n%s\n}", fullName, strings.Join(fields, "\n"))

			return fullName
		}

		// For anonymous structs, inline the definition
		var fields []string
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				jsonTag = strings.ToLower(field.Name)
			} else {
				jsonTag = strings.Split(jsonTag, ",")[0]
			}

			fieldType := fillTypeDefinitions(field.Type)
			fields = append(fields, fmt.Sprintf("  %s: %s", jsonTag, fieldType))
		}
		return fmt.Sprintf("{\n%s\n}", strings.Join(fields, "\n"))

	case reflect.Slice:
		return ("(" + fillTypeDefinitions(t.Elem()) + "[] | null)")

	case reflect.String:
		return "string"

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"

	case reflect.Bool:
		return "boolean"

	case reflect.Ptr:
		return fillTypeDefinitions(t.Elem())

	default:
		return "any"
	}
}
