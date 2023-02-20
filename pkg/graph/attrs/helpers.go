package attrs

import (
	"fmt"
	"image/color"
	"time"
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
