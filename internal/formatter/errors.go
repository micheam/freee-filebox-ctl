package formatter

import "fmt"

// UnsupportedFieldError  represents an error for unsupported fields in receipt formatting.
// This will be returned when a field that is not recognized or supported is requested.
type UnsupportedFieldError struct {
	idx       uint
	fieldName string
}

func (e *UnsupportedFieldError) Error() string {
	return fmt.Sprintf("unsupported field[%d]: %s", e.idx, e.fieldName)
}

func (e *UnsupportedFieldError) FieldName() string {
	return e.fieldName
}
