package memory

import (
	"reflect"
)

// applyOffset returns a slice of items skipped by offset.
// If offset is negative, it returns the original slice.
// If offset is bigger than the number of items it returns empty slice.
func applyOffset(items reflect.Value, offset int) interface{} {
	if offset > 0 {
		switch {
		case items.Len() >= offset:
			return items.Slice(offset, items.Len()).Interface()
		default:
			return reflect.Zero(reflect.TypeOf(items.Interface())).Interface()
		}
	}
	return items.Interface()
}

// applyLimit returns limit number of items.
// If limit is either negative or bigger than the number of itmes it returns all items.
func applyLimit(items reflect.Value, limit int) interface{} {
	if limit > 0 {
		switch {
		case items.Len() >= limit:
			return items.Slice(0, limit).Interface()
		default:
			return items.Interface()
		}
	}
	return items.Interface()
}

// applyOffsetLimit applies offset and limit to items and returns the result.
func applyOffsetLimit(items interface{}, offset int, limit int) interface{} {
	val := reflect.ValueOf(items)
	o := applyOffset(val, offset)
	if reflect.ValueOf(o).Len() == 0 {
		return reflect.Zero(reflect.TypeOf(val.Interface())).Interface()
	}
	return applyLimit(reflect.ValueOf(o), limit)
}
