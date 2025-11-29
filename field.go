package named

import (
	"encoding/json"
	"strings"
)

// ################################
// basics
// ################################

// fieldHeader must match with the initial layout of Field[T] and FieldSlice[T,E]
type fieldHeader struct {
	path       *[]string
	parentPath *[]string
}

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
	path       *[]string // goes first so it's aligned with fieldHeader
	parentPath *[]string // second field, aligned with fieldHeader
	Value      T
}

// Name returns the leaf name of the field (last component of the path).
func (f *Field[T]) Name() string {
	return FieldNameOp(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *Field[T]) FullName(separator string) string {
	combinedPath := f.getCombinedPath()
	if combinedPath == nil || len(*combinedPath) == 0 {
		return ""
	}
	return FieldFullNameOp(combinedPath, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *Field[T]) Path() []string {
	combinedPath := f.getCombinedPath()
	return FieldPathOp(combinedPath)
}

// getCombinedPath combines parentPath and path into a single path
func (f *Field[T]) getCombinedPath() *[]string {
	if f.parentPath == nil || len(*f.parentPath) == 0 {
		return f.path
	}
	if f.path == nil || len(*f.path) == 0 {
		return f.parentPath
	}
	// Combine parentPath + path
	combined := make([]string, len(*f.parentPath)+len(*f.path))
	copy(combined, *f.parentPath)
	copy(combined[len(*f.parentPath):], *f.path)
	return &combined
}

// ParentPath returns the runtime parent path if set.
// Returns nil if no parent path was assigned during linking.
func (f *Field[T]) ParentPath() []string {
	if f.parentPath == nil {
		return nil
	}
	return *f.parentPath
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
	path       *[]string // goes first so it's aligned with fieldHeader
	parentPath *[]string // second field, aligned with fieldHeader
	Value      T
}

// Name returns the leaf name of the field (last component of the path).
func (f *FieldSlice[T, E]) Name() string {
	return FieldNameOp(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *FieldSlice[T, E]) FullName(separator string) string {
	combinedPath := f.getCombinedPath()
	if combinedPath == nil || len(*combinedPath) == 0 {
		return ""
	}
	return FieldFullNameOp(combinedPath, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *FieldSlice[T, E]) Path() []string {
	combinedPath := f.getCombinedPath()
	return FieldPathOp(combinedPath)
}

// getCombinedPath combines parentPath and path into a single path
func (f *FieldSlice[T, E]) getCombinedPath() *[]string {
	if f.parentPath == nil || len(*f.parentPath) == 0 {
		return f.path
	}
	if f.path == nil || len(*f.path) == 0 {
		return f.parentPath
	}
	// Combine parentPath + path
	combined := make([]string, len(*f.parentPath)+len(*f.path))
	copy(combined, *f.parentPath)
	copy(combined[len(*f.parentPath):], *f.path)
	return &combined
}

// ParentPath returns the runtime parent path if set.
// Returns nil if no parent path was assigned during linking.
func (f *FieldSlice[T, E]) ParentPath() []string {
	if f.parentPath == nil {
		return nil
	}
	return *f.parentPath
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
