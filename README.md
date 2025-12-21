# Union

### Tagged unions (discriminated unions) for Go.

Union provides a generic tagged union (discriminated union) implementation for Go with JSON marshaling and unmarshaling support.

## Installation

```sh
go get github.com/eriicafes/union
```

## Usage

### Define a union type

Create a struct where each field represents a possible variant of the union. Only one field should be set at a time.

Use struct tags to customize the variant names in JSON.

```go
package main

import "github.com/eriicafes/union"

type Shape struct {
    Circle    *Circle    `variant:"circle"`
    Rectangle *Rectangle `variant:"rectangle"`
    Triangle  *Triangle  `variant:"triangle"`
}

type Circle struct {
    Radius float64 `json:"radius"`
}

type Rectangle struct {
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

type Triangle struct {
    Base   float64 `json:"base"`
    Height float64 `json:"height"`
}
```

### Create and use a union

```go
func main() {
    var shape union.TaggedUnion[Shape]

    // Set the active variant
    shape.Value.Circle = &Circle{Radius: 5.0}

    // Get the active value
    value := shape.GetValue() // returns *Circle{Radius: 5.0}
}
```

### JSON marshaling

Unions serialize to JSON with a type field indicating which variant is active and a value field containing the variant's data.

```go
shape.Value.Circle = &Circle{Radius: 5.0}

data, _ := json.Marshal(shape)
// {"type": "circle", "value": {"radius": 5}}
```

### JSON unmarshaling

```go
jsonData := []byte(`{"type": "rectangle", "value": {"width": 10, "height": 5}}`)

var shape union.TaggedUnion[Shape]
json.Unmarshal(jsonData, &shape)

// shape.Value.Rectangle is now set to &Rectangle{Width: 10, Height: 5}
```

## Custom field names

Implement the `TaggedFieldNames` method to customize the JSON field names.

```go
func (s Shape) TaggedFieldNames() (variant, value string) {
    return "kind", "data"
}

// Marshals to: {"kind": "circle", "data": {...}}
```

## Error handling

The union enforces invariants and returns errors when:
- No variant is set (zero state)
- Multiple variants are set (invalid state)
- The variant field doesn't match any known variant
- JSON data is malformed
