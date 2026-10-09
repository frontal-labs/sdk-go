package frontal

import (
	"encoding/json"
)

// Field represents an optional request value. Its zero value is unset; use F
// to send a value, including a zero value, or Null to send JSON null.
type Field[T any] interface {
	json.Marshaler
	IsSet() bool
	IsNull() bool
	frontalField()
}

type fieldValue[T any] struct {
	value T
	null  bool
}

// F creates a set request field containing value.
func F[T any](value T) Field[T] {
	return fieldValue[T]{value: value}
}

// Null creates a set request field that serializes as JSON null.
func Null[T any]() Field[T] {
	return fieldValue[T]{null: true}
}

func (fieldValue[T]) frontalField() {}

func (field fieldValue[T]) IsSet() bool { return true }

func (field fieldValue[T]) IsNull() bool { return field.null }

func (field fieldValue[T]) MarshalJSON() ([]byte, error) {
	if field.null {
		return []byte("null"), nil
	}
	return json.Marshal(field.value)
}
