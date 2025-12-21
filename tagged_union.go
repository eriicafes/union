// Package union provides a generic tagged union (discriminated union) implementation
// for Go with JSON marshaling and unmarshaling support.
//
// A tagged union allows you to represent a value that can be one of several types,
// with a variant field that indicates which type is currently active.
//
// Example usage:
//
//	type Shape struct {
//	    Circle    *Circle    `variant:"circle"`
//	    Rectangle *Rectangle `variant:"rectangle"`
//	    Triangle  *Triangle  `variant:"triangle"`
//	}
//
//	var shape union.TaggedUnion[Shape]
//	shape.Value.Circle = &Circle{Radius: 5.0}
//
//	// Marshals to: {"type": "circle", "value": {"Radius": 5.0}}
//	data, _ := json.Marshal(shape)
//
// The `variant` struct tag specifies the variant name in JSON. If no tag is provided,
// the field name is used (e.g., "Circle" instead of "circle").
//
// Custom field names can be specified by implementing TaggedFieldNames():
//
//	func (s Shape) TaggedFieldNames() (variant, value string) {
//	    return "kind", "data"
//	}
//
//	// Marshals to: {"kind": "circle", "data": {...}}
package union

import (
	"cmp"
	"encoding/json"
	"errors"
	"reflect"
)

// TaggedUnion represents a discriminated union type that can hold one of several
// variant types defined in the Spec struct. The Spec type should be a struct where
// each field represents a possible variant of the union.
//
// Only one field in the Spec struct should be non-zero at any time. When marshaling
// to JSON, the union is represented as an object with a variant field (indicating which
// variant is active) and a value field (containing the variant's data).
type TaggedUnion[Spec any] struct{ Value Spec }

// fieldNames returns the names of the variant and value fields to use in JSON marshaling.
// It checks if the Spec type implements a TaggedFieldNames() method and uses those names,
// otherwise defaults to "type" and "value".
func (u *TaggedUnion[Spec]) fieldNames() (variant, value string) {
	if tu, ok := any(u.Value).(interface {
		TaggedFieldNames() (variant, value string)
	}); ok {
		return tu.TaggedFieldNames()
	}
	return "type", "value"
}

// GetValue returns the value of the active variant in the union.
// It iterates through all fields in the Spec struct and returns the value
// of the non-zero field. If no fields are set or multiple fields are set,
// it returns nil (indicating an invalid state).
func (u TaggedUnion[Spec]) GetValue() any {
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
// It serializes the union to JSON as an object with two fields:
//   - A variant field (default "type") containing the variant name
//   - A value field (default "value") containing the variant's data
//
// The variant name is determined by the struct field's `variant` struct tag,
// or the field name if no variant is specified.
//
// Returns an error if:
//   - The Spec type is not a struct
//   - No fields are set (zero state)
//   - Multiple fields are set (invalid state)
func (u TaggedUnion[Spec]) MarshalJSON() ([]byte, error) {
	v := reflect.ValueOf(u.Value)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return nil, errors.New("spec must be a struct")
	}

	var value any
	var variant string
	for i := 0; i < t.NumField(); i++ {
		vf := v.Field(i)
		tf := t.Field(i)

		if vf.IsZero() {
			continue
		}
		if value != nil {
			// invariant violation: multiple variants set
			return nil, errors.New("multiple variants set")
		}
		value = vf.Interface()
		variant = cmp.Or(tf.Tag.Get("variant"), tf.Name)
	}
	if value == nil {
		return nil, errors.New("zero variants set")
	}

	variantField, valueField := u.fieldNames()
	out := map[string]any{
		variantField: variant,
		valueField:   value,
	}

	return json.Marshal(out)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It deserializes JSON data into the union by:
//  1. Reading the variant field to determine which variant is active
//  2. Unmarshaling the value field into the corresponding struct field
//
// The method handles both pointer and non-pointer fields correctly.
//
// Returns an error if:
//   - The JSON data is malformed
//   - The Spec type is not a struct
//   - The variant or value fields are missing
//   - The variant field doesn't match any known variant
//   - Multiple struct fields match the same variant (invalid Spec definition)
//   - The value cannot be unmarshaled into the target field type
func (u *TaggedUnion[Spec]) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	v := reflect.ValueOf(&u.Value).Elem()
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return errors.New("spec must be a struct")
	}

	variantField, valueField := u.fieldNames()
	rawType, ok := raw[variantField]
	if !ok {
		return errors.New("missing variant field: " + variantField)
	}
	rawValue, ok := raw[valueField]
	if !ok {
		return errors.New("missing value field: " + valueField)
	}

	var variant string
	if err := json.Unmarshal(rawType, &variant); err != nil {
		return err
	}

	var matched bool
	for i := 0; i < t.NumField(); i++ {
		vf := v.Field(i)
		tf := t.Field(i)

		if cmp.Or(tf.Tag.Get("variant"), tf.Name) != variant {
			continue
		}
		if matched {
			return errors.New("multiple fields matched")
		}

		target := reflect.New(tf.Type)
		if err := json.Unmarshal(rawValue, target.Interface()); err != nil {
			return err
		}

		vf.Set(target.Elem())
		matched = true
	}
	if !matched {
		return errors.New("unknown variant: " + variant)
	}

	return nil
}
