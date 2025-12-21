# Union

### Generic union types for Go with JSON support.

Union provides generic union type implementations for Go with JSON marshaling and unmarshaling support:

- **TaggedUnion**: A discriminated union with explicit variant/value wrapper
- **Union**: An untagged union that marshals data directly

## Installation

```sh
go get github.com/eriicafes/union
```

## TaggedUnion

TaggedUnion represents a discriminated union where the JSON representation includes a variant field indicating which type is active and a value field containing the data.

### Define a tagged union type

Create a struct where each field represents a possible variant. Use `variant` struct tags to customize variant names in JSON.

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

### Create and use a tagged union

```go
func main() {
    var shape union.TaggedUnion[Shape]

    // Set the active variant
    shape.Value.Circle = &Circle{Radius: 5.0}

    // Get the active value
    value := shape.GetValue() // returns *Circle{Radius: 5.0}
}
```

### JSON marshaling (TaggedUnion)

TaggedUnion serializes to JSON with a type field indicating which variant is active and a value field containing the variant's data.

```go
shape.Value.Circle = &Circle{Radius: 5.0}

data, _ := json.Marshal(shape)
// {"type": "circle", "value": {"radius": 5}}
```

### JSON unmarshaling (TaggedUnion)

```go
jsonData := []byte(`{"type": "rectangle", "value": {"width": 10, "height": 5}}`)

var shape union.TaggedUnion[Shape]
json.Unmarshal(jsonData, &shape)

// shape.Value.Rectangle is now set to &Rectangle{Width: 10, Height: 5}
```

### Custom field names

Implement the `TaggedFieldNames` method to customize the JSON field names.

```go
func (s Shape) TaggedFieldNames() (variant, value string) {
    return "kind", "data"
}

// Marshals to: {"kind": "circle", "data": {...}}
```

## Union

Union represents an untagged union where the JSON representation is the data itself, without any wrapper. When unmarshaling, each field is tried in order until one successfully deserializes to a non-zero value.

### Define an untagged union type

Create a struct where each field represents a possible variant. No struct tags are needed.

```go
type Shape struct {
    Circle    *Circle
    Rectangle *Rectangle
    Triangle  *Triangle
}
```

### Create and use an untagged union

```go
func main() {
    var shape union.Union[Shape]

    // Set the active variant
    shape.Value.Circle = &Circle{Radius: 5.0}

    // Get the active value
    value := shape.GetValue() // returns *Circle{Radius: 5.0}
}
```

### JSON marshaling (Union)

Union serializes data directly without a wrapper.

```go
shape.Value.Circle = &Circle{Radius: 5.0}

data, _ := json.Marshal(shape)
// {"radius": 5}
```

### JSON unmarshaling (Union)

Union tries each field in order until one successfully deserializes to a non-zero value.

```go
jsonData := []byte(`{"width": 10, "height": 5}`)

var shape union.Union[Shape]
json.Unmarshal(jsonData, &shape)

// shape.Value.Rectangle is now set to &Rectangle{Width: 10, Height: 5}
```

## Error handling

Both union types enforce invariants and return errors when:
- No variant is set (zero state)
- Multiple variants are set (invalid state)
- JSON data is malformed

**TaggedUnion** additionally returns errors when:
- The variant field doesn't match any known variant
- The variant or value fields are missing

**Union** additionally returns errors when:
- No field successfully unmarshals to a non-zero value
