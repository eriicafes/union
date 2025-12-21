package union

import (
	"encoding/json"
	"strings"
	"testing"
)

type UnionShape struct {
	Circle    *Circle
	Rectangle *Rectangle
	Triangle  *Triangle
}

type UnionNonPointerShape struct {
	Circle    Circle
	Rectangle Rectangle
}

type UnionEmptyShape struct {
}

type UnionNonStructType int

func TestUnionGetValue(t *testing.T) {
	tests := []struct {
		name     string
		shape    interface{ GetValue() any }
		expected any
	}{
		{
			name: "returns circle variant",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Circle: &Circle{Radius: 5.0},
				},
			},
			expected: Circle{Radius: 5.0},
		},
		{
			name: "returns rectangle variant",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expected: Rectangle{Width: 10, Height: 5},
		},
		{
			name: "returns triangle variant",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Triangle: &Triangle{Base: 8, Height: 4},
				},
			},
			expected: Triangle{Base: 8, Height: 4},
		},
		{
			name:     "returns nil when no variant is set",
			shape:    Union[UnionShape]{},
			expected: nil,
		},
		{
			name: "returns nil when multiple variants are set",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Circle:    &Circle{Radius: 5.0},
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expected: nil,
		},
		{
			name:     "returns nil for non-struct type",
			shape:    Union[NonStructType]{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.shape.GetValue()
			assertUnionValueEquals(t, value, tt.expected)
		})
	}
}

func TestUnionMarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		shape       any
		expected    string
		expectErr   bool
		expectedErr string
	}{
		{
			name: "marshals circle variant",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Circle: &Circle{Radius: 5.0},
				},
			},
			expected: `{"radius":5}`,
		},
		{
			name: "marshals rectangle variant",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expected: `{"width":10,"height":5}`,
		},
		{
			name: "marshals triangle variant",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Triangle: &Triangle{Base: 8, Height: 4},
				},
			},
			expected: `{"base":8,"height":4}`,
		},
		{
			name:        "returns error when no variant is set",
			shape:       Union[UnionShape]{},
			expectErr:   true,
			expectedErr: "zero variants set",
		},
		{
			name: "returns error when multiple variants are set",
			shape: Union[UnionShape]{
				Value: UnionShape{
					Circle:    &Circle{Radius: 5.0},
					Rectangle: &Rectangle{Width: 10, Height: 5},
				},
			},
			expectErr:   true,
			expectedErr: "multiple variants set",
		},
		{
			name: "marshals non-pointer variant",
			shape: Union[UnionNonPointerShape]{
				Value: UnionNonPointerShape{
					Circle: Circle{Radius: 5.0},
				},
			},
			expected: `{"radius":5}`,
		},
		{
			name:        "returns error for struct with no variants",
			shape:       Union[UnionEmptyShape]{},
			expectErr:   true,
			expectedErr: "zero variants set",
		},
		{
			name:        "returns error for non-struct type",
			shape:       Union[UnionNonStructType]{},
			expectErr:   true,
			expectedErr: "spec must be a struct",
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

func TestUnionUnmarshalJSON(t *testing.T) {
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
			shape:    &Union[UnionShape]{},
			jsonData: `{"radius":5}`,
			expected: Circle{Radius: 5.0},
		},
		{
			name:     "unmarshals rectangle variant",
			shape:    &Union[UnionShape]{},
			jsonData: `{"width":10,"height":5}`,
			expected: Rectangle{Width: 10, Height: 5},
		},
		{
			name:     "unmarshals triangle variant",
			shape:    &Union[UnionShape]{},
			jsonData: `{"base":8,"height":4}`,
			expected: Triangle{Base: 8, Height: 4},
		},
		{
			name:        "returns error when no field matches",
			shape:       &Union[UnionShape]{},
			jsonData:    `{"sides":6}`,
			expectErr:   true,
			expectedErr: "no field matched",
		},
		{
			name:      "returns error for malformed JSON",
			shape:     &Union[UnionShape]{},
			jsonData:  `{invalid json}`,
			expectErr: true,
		},
		{
			name:     "unmarshals non-pointer variant",
			shape:    &Union[UnionNonPointerShape]{},
			jsonData: `{"radius":5}`,
			expected: Circle{Radius: 5.0},
		},
		{
			name:        "returns error for struct with no variants",
			shape:       &Union[UnionEmptyShape]{},
			jsonData:    `{"radius":5}`,
			expectErr:   true,
			expectedErr: "no field matched",
		},
		{
			name:        "returns error for non-struct type",
			shape:       &Union[UnionNonStructType]{},
			jsonData:    `42`,
			expectErr:   true,
			expectedErr: "spec must be a struct",
		},
		{
			name:     "matches first non-zero field in order",
			shape:    &Union[UnionShape]{},
			jsonData: `{"height":10}`,
			expected: Rectangle{Width: 0, Height: 10},
		},
		{
			name:      "returns error when JSON is array",
			shape:     &Union[UnionShape]{},
			jsonData:  `[1, 2, 3]`,
			expectErr: true,
		},
		{
			name:      "returns error when JSON is string",
			shape:     &Union[UnionShape]{},
			jsonData:  `"hello"`,
			expectErr: true,
		},
		{
			name:      "returns error when JSON is number",
			shape:     &Union[UnionShape]{},
			jsonData:  `123`,
			expectErr: true,
		},
		{
			name:        "returns error when value cannot be unmarshaled",
			shape:       &Union[UnionShape]{},
			jsonData:    `{"radius":"not a number"}`,
			expectErr:   true,
			expectedErr: "no field matched",
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
			assertUnionValueEquals(t, value, tt.expected)
		})
	}
}

// assertUnionValueEquals compares a value from GetValue() with an expected value type.
// It handles pointer dereferencing and type assertions for Circle, Rectangle, and Triangle.
func assertUnionValueEquals(t *testing.T, value, expected any) {
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
