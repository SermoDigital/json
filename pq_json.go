package json

import (
	"database/sql/driver"

	"github.com/sermodigital/errors"
)

// JSON is a generic JSON blob.
type JSON map[string]interface{}

// Value implements driver.Valuer.
func (j JSON) Value() (driver.Value, error) {
	return Marshal(j)
}

// Scan implements sql.Scanner.
func (j *JSON) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.Errorf("invalid type: %T (wanted []byte)", value)
	}
	return Unmarshal(data, j)
}
