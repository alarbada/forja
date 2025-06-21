package forja

import (
	"fmt"
	"reflect"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type typegen struct {
	typeDefs        *orderedmap.OrderedMap[string, string]
	processingTypes map[string]bool
}

func newTypegen() *typegen {
	return &typegen{
		typeDefs:        orderedmap.New[string, string](),
		processingTypes: make(map[string]bool),
	}
}

func (tp *typegen) printTypeDefs(sb *strings.Builder) {
	// OrderedMap maintains insertion order, so we can iterate directly
	for pair := tp.typeDefs.Oldest(); pair != nil; pair = pair.Next() {
		fmt.Fprintln(sb, pair.Value)
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

func escapeFieldName(name string) string {
	// If empty, needs quotes
	if name == "" {
		return "''"
	}

	// Check first character - must start with letter, underscore, or dollar sign
	if !strings.ContainsAny(string(name[0]), "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_$") {
		return "'" + name + "'"
	}

	// Check rest of string - can only contain letters, numbers, underscore, or dollar sign
	for _, ch := range name[1:] {
		if !strings.ContainsAny(string(ch), "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_$") {
			return "'" + name + "'"
		}
	}

	return name
}

func (tp *typegen) FillTypeDefinitions(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Struct:
		fullName := getFullTypeName(t)
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return "string"
		}

		if strings.Contains(fullName, "forja_Option") {
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				if field.Name == "Value" {
					return tp.FillTypeDefinitions(field.Type)
				}
			}
		}

		if fullName != "" {
			// Check if we're already processing this type (circular reference)
			if tp.processingTypes[fullName] {
				return fullName // Just return the type name for circular references
			}

			if _, exists := tp.typeDefs.Get(fullName); exists {
				return fullName
			}

			// Mark this type as being processed
			tp.processingTypes[fullName] = true

			var fields []string
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				fieldName := field.Name
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" {
					fieldName = strings.Split(jsonTag, ",")[0]
				}
				fieldName = escapeFieldName(fieldName)

				fieldType := tp.FillTypeDefinitions(field.Type)
				if field.Type.Kind() == reflect.Ptr {
					fields = append(fields, fmt.Sprintf("  %s?: %s", fieldName, fieldType))
				} else if strings.Contains(getFullTypeName(field.Type), "forja_Option") {
					fields = append(fields, fmt.Sprintf("  %s?: %s", fieldName, fieldType))
				} else {
					fields = append(fields, fmt.Sprintf("  %s: %s", fieldName, fieldType))
				}

			}

			// Remove from processing map after we're done
			delete(tp.processingTypes, fullName)

			tp.typeDefs.Set(fullName, fmt.Sprintf("export type %s = {\n%s\n}", fullName, strings.Join(fields, "\n")))
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
			jsonTag = escapeFieldName(jsonTag)
			fieldType := tp.FillTypeDefinitions(field.Type)
			optional := ""
			if field.Type.Kind() == reflect.Ptr {
				fieldType = tp.FillTypeDefinitions(field.Type.Elem())
				optional = "?"
			}
			fields = append(fields, fmt.Sprintf("  %s%s: %s", jsonTag, optional, fieldType))
		}
		return fmt.Sprintf("{\n%s\n}", strings.Join(fields, "\n"))

	case reflect.Slice:
		return fmt.Sprintf("(%s[] | null)", tp.FillTypeDefinitions(t.Elem()))
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Ptr:
		return tp.FillTypeDefinitions(t.Elem())
	default:
		return "any"
	}
}

func (tp *typegen) generateTypeDefinition(t reflect.Type) string {
	typename := tp.FillTypeDefinitions(t)

	var sb strings.Builder

	if typename == "" {
		panic("Anonymous type not supported")
	}

	typedef, _ := tp.typeDefs.Get(typename)
	fmt.Fprintln(&sb, typedef)
	return sb.String()
}
