package forja

import (
	"fmt"
	"reflect"
	"strings"
)

var typeDefinitions = make(map[string]string)
var processingTypes = make(map[string]bool)

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
	pkgParts := strings.Split(t.PkgPath(), "/")
	pkg := pkgParts[len(pkgParts)-1]
	name := fmt.Sprintf("%s_%s", pkg, t.Name())
	return name
}

func fillTypeDefinitions(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Struct:
		fullName := getFullTypeName(t)
		if strings.Contains(fullName, "forja_Option") {
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if field.Name == "Value" {
					return fillTypeDefinitions(field.Type)
				}
			}
		}

		if fullName != "" {
			// Check if we're already processing this type (circular reference)
			if processingTypes[fullName] {
				return fullName // Just return the type name for circular references
			}

			if _, exists := typeDefinitions[fullName]; exists {
				return fullName
			}

			// Mark this type as being processed
			processingTypes[fullName] = true

			var fields []string
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				fieldName := field.Name
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" {
					fieldName = strings.Split(jsonTag, ",")[0]
				}

				fieldType := fillTypeDefinitions(field.Type)
				if field.Type.Kind() == reflect.Ptr {
					fmt.Println(fieldName, fieldType)
					fields = append(fields, fmt.Sprintf("  %s?: %s", fieldName, fieldType))
				} else if strings.Contains(getFullTypeName(field.Type), "forja_Option") {
					fields = append(fields, fmt.Sprintf("  %s?: %s", fieldName, fieldType))
				} else {
					fields = append(fields, fmt.Sprintf("  %s: %s", fieldName, fieldType))
				}

			}

			// Remove from processing map after we're done
			delete(processingTypes, fullName)

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
			optional := ""
			if field.Type.Kind() == reflect.Ptr {
				fieldType = fillTypeDefinitions(field.Type.Elem())
				optional = "?"
			}
			fields = append(fields, fmt.Sprintf("  %s%s: %s", jsonTag, optional, fieldType))
		}
		return fmt.Sprintf("{\n%s\n}", strings.Join(fields, "\n"))

	case reflect.Slice:
		return fmt.Sprintf("(%s[] | null)", fillTypeDefinitions(t.Elem()))
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
