package named

import (
	"encoding/json"
	"strings"
)

// ################################
// basics
// ################################

var TextMarshaler = func(v any) ([]byte, error) {
	return json.Marshal(v)
}
var TextUnmarshaler = func(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

const DefaulyFullNameSeparator = "."

func FieldNameOp(pathPtr *[]string) string {
	if pathPtr == nil || len(*pathPtr) == 0 {
		return ""
	}
	return (*pathPtr)[len(*pathPtr)-1]
}

func FieldFullNameOp(pathPtr *[]string, separator string) string {

	if separator == "" {
		separator = DefaulyFullNameSeparator
	}
	return strings.Join(*pathPtr, separator)
}

func FieldPathOp(pathPtr *[]string) []string {
	if pathPtr == nil {
		return nil
	}
	return *pathPtr
}

func FieldNoNameOp(pathPtr *[]string) bool {
	return pathPtr == nil || len(*pathPtr) == 0
}

// ################################
// comparable Field[T]
// ################################

type Field[T comparable] struct {
	path  *[]string // goes first so it's aligned with fieldHeader
	Value T
}

// Name returns the leaf name of the field (last component of the path).
func (f *Field[T]) Name() string {
	return FieldNameOp(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *Field[T]) FullName(separator string) string {
	if FieldNoNameOp(f.path) {
		return ""
	}
	return FieldFullNameOp(f.path, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *Field[T]) Path() []string {
	return FieldPathOp(f.path)
}

func (f *Field[T]) NoName() bool {
	return FieldNoNameOp(f.path)
}

func (f *Field[T]) NoValue() bool {
	var zero T
	return f.Value == zero
}

// IsZero reports whether the Field's value is the zero value for its type.
// This method is used by encoding/json to support the omitempty tag.
func (f *Field[T]) IsZero() bool {
	return f.NoValue()
}

func (f Field[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

func (f *Field[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &f.Value)
}

func (f *Field[T]) MarshalText() (text []byte, err error) {
	return TextMarshaler(f.Value)
}

func (f *Field[T]) UnmarshalText(text []byte) error {
	return TextUnmarshaler(text, &f.Value)
}

// ################################
// slice Field[T]
// ################################

type Slice[T any] interface {
	~[]T
}

type FieldSlice[T Slice[E], E any] struct {
	path  *[]string // goes first so it's aligned with fieldHeader
	Value T
}

// Name returns the leaf name of the field (last component of the path).
func (f *FieldSlice[T, E]) Name() string {
	return FieldNameOp(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *FieldSlice[T, E]) FullName(separator string) string {
	if FieldNoNameOp(f.path) {
		return ""
	}
	return FieldFullNameOp(f.path, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *FieldSlice[T, E]) Path() []string {
	return FieldPathOp(f.path)
}

func (f *FieldSlice[T, E]) NoName() bool {
	return FieldNoNameOp(f.path)
}

func (f *FieldSlice[T, E]) NoValue() bool {
	return len(f.Value) == 0
}

// IsZero reports whether the Field's value is the zero value for its type.
// This method is used by encoding/json to support the omitempty tag.
func (f *FieldSlice[T, E]) IsZero() bool {
	return f.NoValue()
}

func (f FieldSlice[T, E]) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

func (f *FieldSlice[T, E]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &f.Value)
}

func (f *FieldSlice[T, E]) MarshalText() (text []byte, err error) {
	return TextMarshaler(f.Value)
}

func (f *FieldSlice[T, E]) UnmarshalText(text []byte) error {
	return TextUnmarshaler(text, &f.Value)
}
