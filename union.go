package union

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
)

// Union represents an untagged union type that can hold one of several
// variant types defined in the Spec struct. The Spec type should be a struct where
// each field represents a possible variant of the union.
//
// Only one field in the Spec struct should be non-zero at any time. When marshaling
// to JSON, the union's data is marshaled directly without a wrapper. When unmarshaling,
// each field is tried in order until one successfully deserializes to a non-zero value.
type Union[Spec any] struct{ Value Spec }

// GetValue returns the value of the active variant in the union.
// It iterates through all fields in the Spec struct and returns the value
// of the non-zero field. If no fields are set or multiple fields are set,
// it returns nil (indicating an invalid state).
func (u Union[Spec]) GetValue() any {
	v := reflect.ValueOf(u.Value)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil
	}

	var value any
	for i := 0; i < t.NumField(); i++ {
		vf := v.Field(i)

		if vf.IsZero() {
			continue
		}
		if value != nil {
			// invariant violation: multiple variants set
			return nil
		}
		value = vf.Interface()
	}

	return value
}

// MarshalJSON implements the json.Marshaler interface.
// It serializes the union's active variant data directly to JSON.
//
// Returns an error if:
//   - The Spec type is not a struct
//   - No fields are set (zero state)
//   - Multiple fields are set (invalid state)
func (u Union[Spec]) MarshalJSON() ([]byte, error) {
	v := reflect.ValueOf(u.Value)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil, errors.New("spec must be a struct")
	}

	var value any
	for i := 0; i < t.NumField(); i++ {
		vf := v.Field(i)

		if vf.IsZero() {
			continue
		}
		if value != nil {
			// invariant violation: multiple variants set
			return nil, errors.New("multiple variants set")
		}
		value = vf.Interface()
	}
	if value == nil {
		return nil, errors.New("zero variants set")
	}

	return json.Marshal(value)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It deserializes JSON data into the union by trying each field in order
// until one successfully unmarshals to a non-zero value.
// Uses strict matching to ensure all JSON fields map to struct fields.
//
// Returns an error if:
//   - The JSON data is malformed
//   - The Spec type is not a struct
//   - No field successfully unmarshals to a non-zero value
func (u *Union[Spec]) UnmarshalJSON(data []byte) error {
	v := reflect.ValueOf(&u.Value).Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return errors.New("spec must be a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		vf := v.Field(i)
		tf := t.Field(i)

		target := reflect.New(tf.Type)

		// Use decoder with DisallowUnknownFields for strict matching
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(target.Interface()); err != nil {
			continue
		}

		// Check if the unmarshaled value is non-zero
		if !target.Elem().IsZero() {
			vf.Set(target.Elem())
			return nil
		}
	}

	return errors.New("no field matched")
}
