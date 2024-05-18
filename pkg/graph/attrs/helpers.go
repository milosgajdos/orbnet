package attrs

import (
	"errors"
	"fmt"
	"image/color"
	"reflect"
	"strings"
	"time"
	"unicode"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

// CopyFrom creates a copy of a and returns it.
func CopyFrom(a map[string]interface{}) map[string]interface{} {
	attrs := make(map[string]interface{})

	for k, v := range a {
		attrs[k] = v
	}

	return attrs
}

// isStringly checks if a is either a string
// or if it implements fmt.Stringer or fmt.GoStringer.
// It returns the bool flag indicating the result and
// the string representation of a.
// If a is not stringly, it returns false and empty string.
func isStringly(a interface{}) (bool, string) {
	switch v := a.(type) {
	case string:
		return true, v
	case fmt.Stringer:
		return true, v.String()
	case fmt.GoStringer:
		return true, v.GoString()
	default:
		return false, ""
	}
}

// ToString attempts to convert well known attributes to string.
// The following attributes are considered as well known:
//   - color
//   - date
//   - weight
//   - name
//   - relation
//
// At the moment the following attribute conversions are implemented:
//   - color to color.RGBA hex codes of RGB channels
//   - date to string representation as per time.RFC3339
//   - weight string representation
//
// If an unknown attribute key is supplied an empty string is returned.
func ToString(k string, v interface{}) string {
	switch k {
	case "color":
		if c, ok := v.(color.RGBA); ok {
			return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
		}
	case "date":
		if d, ok := v.(time.Time); ok {
			return d.Format(time.RFC3339)
		}
	case "weight":
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("%f", f)
		}
	case "name", "relation":
		if val, ok := v.(string); ok {
			return val
		}
	default:
		return ""
	}

	return ""
}

// NOTE(milosgajdos): we should turn map[string]interface{} into proper type.

// ToStringMap attempts to convert a to a map of strings.
// It first checks if the stored attribute value is stringly i.e. either of string,
// fmt.Stringer or fmt.GoStringer. If it is it returns its stringe representation.
// If the attribute value is not stringly we attempt to convert well known attributes to strings.
// If the attribute is neither stringly nor is it known how to convert it to a string
// the attribute is omitted from the returned map.
func ToStringMap(a map[string]interface{}) map[string]string {
	m := make(map[string]string)

	for k, v := range a {
		ok, val := isStringly(v)
		if ok {
			m[k] = val
		}

		val = ToString(k, v)
		if val != "" {
			m[k] = val
		}
	}

	return m
}

func toSnakeCase(s string) string {
	var words []string
	var currentWord strings.Builder

	for _, r := range s {
		if unicode.IsUpper(r) && currentWord.Len() > 0 {
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
		currentWord.WriteRune(unicode.ToLower(r))
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return strings.Join(words, "_")
}

func flattenValue(value reflect.Value) interface{} {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	if value.Kind() == reflect.Struct && value.Type().NumField() == 1 && value.Type().Field(0).Anonymous {
		return flattenValue(value.Field(0))
	}

	return value.Interface()
}

func parseJSONTag(tag, fieldName string) string {
	if tag == "" || tag == "-" {
		return toSnakeCase(fieldName)
	}
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx]
	}
	return tag
}

// Encode encodes arbitrary structs into attributes map.
func Encode(input interface{}) (map[string]interface{}, error) {
	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil, fmt.Errorf("empty value: %w", ErrInvalidInput)
	}

	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("must be struct: %v", ErrInvalidInput)
	}

	result := make(map[string]interface{})
	repoType := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := repoType.Field(i)
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldValue := value.Field(i)
		// Skip zero values
		if fieldValue.IsZero() {
			continue
		}

		jsonTag := parseJSONTag(field.Tag.Get("json"), field.Name)

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			// Check if the struct is a wrapper type
			if fieldValue.Type().NumField() == 1 && fieldValue.Type().Field(0).Anonymous {
				// Flatten the wrapper type e.g. type Timestamp{time.Time}
				if val := flattenValue(fieldValue.Field(0)); val != nil {
					result[jsonTag] = val
				}
				continue
			}
			var err error
			result[jsonTag], err = Encode(fieldValue)
			if err != nil {
				return nil, err
			}
		default:
			result[jsonTag] = fieldValue.Interface()
		}
	}

	return result, nil
}
