package sqlite

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func parseDSN(dsn string) (string, error) {
	scheme, path, ok := strings.Cut(dsn, "://")
	if !ok {
		return "", fmt.Errorf("invalid dsn")
	}

	if scheme != Scheme {
		return "", fmt.Errorf("invalid dsn scheme")
	}

	if path == "" {
		return "", fmt.Errorf("invalid path")
	}

	return path, nil
}

// NullTime represents a helper wrapper for time.Time. It automatically converts
// time fields to/from RFC 3339 format. Also supports NULL for zero time.
type NullTime time.Time

// Scan reads a time value from the database.
func (n *NullTime) Scan(value interface{}) error {
	if value == nil {
		*(*time.Time)(n) = time.Time{}
		return nil
	} else if value, ok := value.(string); ok {
		*(*time.Time)(n), _ = time.Parse(time.RFC3339, value)
		return nil
	}
	return fmt.Errorf("NullTime: cannot scan to time.Time: %T", value)
}

// Value formats a time value for the database.
func (n *NullTime) Value() (driver.Value, error) {
	if n == nil || (*time.Time)(n).IsZero() {
		return nil, nil
	}
	return (*time.Time)(n).UTC().Format(time.RFC3339), nil
}

func AttrsFromString(attrString string) (map[string]interface{}, error) {
	attrs := map[string]interface{}{}
	if len(attrString) > 0 {
		decoder := json.NewDecoder(strings.NewReader(attrString))
		decoder.UseNumber()
		err := decoder.Decode(&attrs)
		if err != nil {
			return nil, err
		}
		// Convert json.Number to int or float64 as needed
		// NOTE: thanks golang
		for key, value := range attrs {
			if numStr, ok := value.(json.Number); ok {
				if num, err := numStr.Int64(); err == nil {
					attrs[key] = num
				} else if num, err := numStr.Float64(); err == nil {
					attrs[key] = num
				}
			}
		}
	}
	return attrs, nil
}
