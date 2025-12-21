package union

import (
	"encoding/json"
	"strings"
	"testing"
)

type (
	Circle struct {
		Radius float64 `json:"radius"`
	}
	Rectangle struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	}
	Triangle struct {
		Base   float64 `json:"base"`
		Height float64 `json:"height"`
	}
)

type Shape struct {
	Circle    *Circle    `variant:"circle"`
	Rectangle *Rectangle `variant:"rectangle"`
	Triangle  *Triangle  `variant:"triangle"`
}

type CustomFieldNamesShape struct {
	Circle    *Circle    `variant:"circle"`
	Rectangle *Rectangle `variant:"rectangle"`
}

func (s CustomFieldNamesShape) TaggedFieldNames() (variant, value string) {
	return "kind", "data"
}

type NonPointerShape struct {
	Circle    Circle    `variant:"circle"`
	Rectangle Rectangle `variant:"rectangle"`
}

type NonStructTagsShape struct {
	Circle    *Circle
	Rectangle *Rectangle
	Triangle  *Triangle
}

func TestGetValue(t *testing.T) {
	tests := []struct {
		name     string
		shape    TaggedUnion[Shape]
		expected any
	}{
		{
			name: "returns circle variant",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Circle: &Circle{Radius: 5.0},
				},
			},
			expected: Circle{Radius: 5.0},
		},
		{
			name: "returns rectangle variant",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expected: Rectangle{Width: 10, Height: 5},
		},
		{
			name: "returns triangle variant",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Triangle: &Triangle{Base: 8, Height: 4},
				},
			},
			expected: Triangle{Base: 8, Height: 4},
		},
		{
			name:     "returns nil when no variant is set",
			shape:    TaggedUnion[Shape]{},
			expected: nil,
		},
		{
			name: "returns nil when multiple variants are set",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Circle:    &Circle{Radius: 5.0},
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.shape.GetValue()
			assertValueEquals(t, value, tt.expected)
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		shape       any
		expected    string
		expectErr   bool
		expectedErr string
	}{
		{
			name: "marshals circle variant",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Circle: &Circle{Radius: 5.0},
				},
			},
			expected: `{"type":"circle","value":{"radius":5}}`,
		},
		{
			name: "marshals rectangle variant",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expected: `{"type":"rectangle","value":{"width":10,"height":5}}`,
		},
		{
			name: "marshals triangle variant",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Triangle: &Triangle{Base: 8, Height: 4},
				},
			},
			expected: `{"type":"triangle","value":{"base":8,"height":4}}`,
		},
		{
			name:        "returns error when no variant is set",
			shape:       TaggedUnion[Shape]{},
			expectErr:   true,
			expectedErr: "zero variants set",
		},
		{
			name: "returns error when multiple variants are set",
			shape: TaggedUnion[Shape]{
				Value: Shape{
					Circle:    &Circle{Radius: 5.0},
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expectErr:   true,
			expectedErr: "multiple variants set",
		},
		{
			name: "marshals with custom field names",
			shape: TaggedUnion[CustomFieldNamesShape]{
				Value: CustomFieldNamesShape{
					Circle: &Circle{Radius: 5.0},
				},
			},
			expected: `{"data":{"radius":5},"kind":"circle"}`,
		},
		{
			name: "marshals non-pointer variant",
			shape: TaggedUnion[NonPointerShape]{
				Value: NonPointerShape{
					Circle: Circle{Radius: 5.0},
				},
			},
			expected: `{"type":"circle","value":{"radius":5}}`,
		},
		{
			name: "marshals without struct tags",
			shape: TaggedUnion[NonStructTagsShape]{
				Value: NonStructTagsShape{
					Circle: &Circle{Radius: 5.0},
				},
			},
			expected: `{"type":"Circle","value":{"radius":5}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.shape)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.expectedErr != "" && !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error '%s', got '%v'", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(data))
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		shape       interface{ GetValue() any }
		jsonData    string
		expected    any
		expectErr   bool
		expectedErr string
	}{
		{
			name:     "unmarshals circle variant",
			shape:    &TaggedUnion[Shape]{},
			jsonData: `{"type":"circle","value":{"radius":5}}`,
			expected: Circle{Radius: 5.0},
		},
		{
			name:     "unmarshals rectangle variant",
			shape:    &TaggedUnion[Shape]{},
			jsonData: `{"type":"rectangle","value":{"width":10,"height":5}}`,
			expected: Rectangle{Width: 10, Height: 5},
		},
		{
			name:     "unmarshals triangle variant",
			shape:    &TaggedUnion[Shape]{},
			jsonData: `{"type":"triangle","value":{"base":8,"height":4}}`,
			expected: Triangle{Base: 8, Height: 4},
		},
		{
			name:        "returns error for unknown variant",
			shape:       &TaggedUnion[Shape]{},
			jsonData:    `{"type":"hexagon","value":{"sides":6}}`,
			expectErr:   true,
			expectedErr: "unknown variant: hexagon",
		},
		{
			name:        "returns error for missing variant field",
			shape:       &TaggedUnion[Shape]{},
			jsonData:    `{"value":{"Radius":5}}`,
			expectErr:   true,
			expectedErr: "missing variant field: type",
		},
		{
			name:        "returns error for missing value field",
			shape:       &TaggedUnion[Shape]{},
			jsonData:    `{"type":"circle"}`,
			expectErr:   true,
			expectedErr: "missing value field: value",
		},
		{
			name:      "returns error for malformed JSON",
			shape:     &TaggedUnion[Shape]{},
			jsonData:  `{invalid json}`,
			expectErr: true,
		},
		{
			name:     "unmarshals with custom field names",
			shape:    &TaggedUnion[CustomFieldNamesShape]{},
			jsonData: `{"kind":"circle","data":{"radius":5}}`,
			expected: Circle{Radius: 5.0},
		},
		{
			name:     "unmarshals non-pointer variant",
			shape:    &TaggedUnion[NonPointerShape]{},
			jsonData: `{"type":"circle","value":{"radius":5}}`,
			expected: Circle{Radius: 5.0},
		},
		{
			name:     "unmarshals without struct tags",
			shape:    &TaggedUnion[NonStructTagsShape]{},
			jsonData: `{"type":"Circle","value":{"radius":5}}`,
			expected: Circle{Radius: 5.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.jsonData), tt.shape)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.expectedErr != "" && err.Error() != tt.expectedErr {
					t.Errorf("expected error '%s', got '%v'", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			value := tt.shape.GetValue()
			assertValueEquals(t, value, tt.expected)
		})
	}
}

// assertValueEquals compares a value from GetValue() with an expected value type.
// It handles pointer dereferencing and type assertions for Circle, Rectangle, and Triangle.
func assertValueEquals(t *testing.T, value, expected any) {
	t.Helper()

	if expected == nil {
		if value != nil {
			t.Errorf("expected nil, got %v", value)
		}
		return
	}

	if value == nil {
		t.Fatal("expected non-nil value")
	}

	switch expected := expected.(type) {
	case Circle:
		circle, ok := value.(*Circle)
		if !ok {
			if val, okVal := value.(Circle); okVal {
				circle, ok = &val, true
			}
		}
		if !ok {
			t.Fatalf("expected type %T, got %T", expected, value)
		}
		if *circle != expected {
			t.Errorf("expected %+v, got %+v", expected, *circle)
		}
	case Rectangle:
		rect, ok := value.(*Rectangle)
		if !ok {
			if val, okVal := value.(Rectangle); okVal {
				rect, ok = &val, true
			}
		}
		if !ok {
			t.Fatalf("expected type %T, got %T", expected, value)
		}
		if *rect != expected {
			t.Errorf("expected %+v, got %+v", expected, *rect)
		}
	case Triangle:
		tri, ok := value.(*Triangle)
		if !ok {
			if val, okVal := value.(Triangle); okVal {
				tri, ok = &val, true
			}
		}
		if !ok {
			t.Fatalf("expected type %T, got %T", expected, value)
		}
		if *tri != expected {
			t.Errorf("expected %+v, got %+v", expected, *tri)
		}
	default:
		t.Fatalf("unexpected expected type: %T", expected)
	}
}
