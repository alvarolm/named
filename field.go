package named

import (
	"encoding/json"
	"strings"
)

// ################################
// basics
// ################################

const DefaulyFullNameSeparator = "."

func field_name(pathPtr *[]string) string {
	if pathPtr == nil || len(*pathPtr) == 0 {
		return ""
	}
	return (*pathPtr)[len(*pathPtr)-1]
}

func field_fullName(pathPtr *[]string, separator string) string {

	if separator == "" {
		separator = DefaulyFullNameSeparator
	}
	return strings.Join(*pathPtr, separator)
}

func field_path(pathPtr *[]string) []string {
	if pathPtr == nil {
		return nil
	}
	return *pathPtr
}

func field_noName(pathPtr *[]string) bool {
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
	return field_name(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *Field[T]) FullName(separator string) string {
	if field_noName(f.path) {
		return ""
	}
	return field_fullName(f.path, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *Field[T]) Path() []string {
	return field_path(f.path)
}

func (f *Field[T]) NoName() bool {
	return field_noName(f.path)
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
	return field_name(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *FieldSlice[T, E]) FullName(separator string) string {
	if field_noName(f.path) {
		return ""
	}
	return field_fullName(f.path, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *FieldSlice[T, E]) Path() []string {
	return field_path(f.path)
}

func (f *FieldSlice[T, E]) NoName() bool {
	return field_noName(f.path)
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
