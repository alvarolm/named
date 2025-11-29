package named

import (
	"encoding/json"
	"unsafe"
)

// ################################
// basics
// ################################

type fielder interface {
	Name() string
	FullName(separator string) string
	Path() []string
	PathWithoutName() []string
	NoName() bool
	NoValue() bool
	IsZero() bool
}

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

func fieldNameOp(pathPtr *[]string) string {
	if pathPtr == nil || len(*pathPtr) == 0 {
		return ""
	}
	return (*pathPtr)[len(*pathPtr)-1]
}

func stringJoinRawSize(elems []string, sep string) int {
	n := len(sep) * (len(elems) - 1)
	for i := range elems {
		n += len(elems[i])
	}
	return n
}

func fieldFullNameOp(pathPtr, parentPathPtr *[]string, separator string) string {

	if separator == "" {
		separator = DefaulyFullNameSeparator
	}

	if pathPtr == nil {
		return ""
	}

	if parentPathPtr == nil || len(*parentPathPtr) == 0 {

		if len(*pathPtr) == 1 {
			return (*pathPtr)[0]
		}

		n := stringJoinRawSize(*pathPtr, separator)

		buf := make([]byte, 0, n)
		buf = append(buf, (*pathPtr)[0]...)

		for _, elem := range (*pathPtr)[1:] {
			buf = append(buf, separator...)
			buf = append(buf, elem...)
		}

		return unsafe.String(unsafe.SliceData(buf), n)
	}

	size := stringJoinRawSize(*parentPathPtr, separator) + len(separator) + stringJoinRawSize(*pathPtr, separator)

	buf := make([]byte, 0, size)
	buf = append(buf, (*parentPathPtr)[0]...)

	for _, elem := range (*parentPathPtr)[1:] {
		buf = append(buf, separator...)
		buf = append(buf, elem...)
	}

	for _, elem := range *pathPtr {
		buf = append(buf, separator...)
		buf = append(buf, elem...)
	}

	return unsafe.String(unsafe.SliceData(buf), size)
}

func fieldPathWithoutNameOp(pathPtr *[]string) []string {
	if pathPtr == nil || len(*pathPtr) == 0 {
		return nil
	}
	return (*pathPtr)[:len(*pathPtr)-1]
}

func fieldNoNameOp(pathPtr *[]string) bool {
	return pathPtr == nil || len(*pathPtr) == 0
}

// getCombinedPath combines parentPath and path into a single path
func getCombinedPath(path, parent *[]string) []string {
	if path == nil {
		return nil
	}
	if parent == nil {
		return *path
	}
	// Combine parentPath + path
	combined := make([]string, len(*parent)+len(*path))
	copy(combined, *parent)
	copy(combined[len(*parent):], *path)
	return combined
}

// ################################
// comparable Field[T]
// ################################

type Field[T comparable] struct {
	path       *[]string // goes first so it's aligned with fieldHeader
	parentPath *[]string // second field, aligned with fieldHeader
	Value      T
}

var _ fielder = (*Field[int])(nil) // check interface compliance

// Name returns the leaf name of the field (last component of the path).
func (f *Field[T]) Name() string {
	return fieldNameOp(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *Field[T]) FullName(separator string) string {
	return fieldFullNameOp(f.path, f.parentPath, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *Field[T]) Path() []string {
	return getCombinedPath(f.path, f.parentPath)
}

func (f *Field[T]) PathWithoutName() []string {
	return fieldPathWithoutNameOp(f.path)
}

func (f *Field[T]) NoName() bool {
	return fieldNoNameOp(f.path)
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

var _ fielder = (*FieldSlice[[]int, int])(nil) // check interface compliance

// Name returns the leaf name of the field (last component of the path).
func (f *FieldSlice[T, E]) Name() string {
	return fieldNameOp(f.path)
}

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *FieldSlice[T, E]) FullName(separator string) string {
	return fieldFullNameOp(f.path, f.parentPath, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *FieldSlice[T, E]) Path() []string {
	return getCombinedPath(f.path, f.parentPath)
}

func (f *FieldSlice[T, E]) PathWithoutName() []string {
	return fieldPathWithoutNameOp(f.path)
}

func (f *FieldSlice[T, E]) NoName() bool {
	return fieldNoNameOp(f.path)
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
