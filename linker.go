package named

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type Field[T any] struct {
	path  *[]string // goes first so it's aligned with fieldHeader
	Value T
}

// Name returns the leaf name of the field (last component of the path).
func (f *Field[T]) Name() string {
	if f.path == nil || len(*f.path) == 0 {
		return ""
	}
	return (*f.path)[len(*f.path)-1]
}

const DefaulyFullNameSeparator = "."

// FullName returns the full hierarchical path as a separated string.
// If separator is empty, defaults to ".".
// This provides backward compatibility for users who need the old Name() behavior.
func (f *Field[T]) FullName(separator string) string {
	if f.path == nil || len(*f.path) == 0 {
		return ""
	}
	if separator == "" {
		separator = DefaulyFullNameSeparator
	}
	return strings.Join(*f.path, separator)
}

// Path returns the complete hierarchical path as a slice.
// Returns nil if the field has no path information.
func (f *Field[T]) Path() []string {
	if f.path == nil {
		return nil
	}
	return *f.path
}

func (f *Field[T]) NoName() bool {
	return f.path == nil || len(*f.path) == 0
}

func (f *Field[T]) NoValue() bool {
	v := any(f.Value)
	switch val := v.(type) {
	case int:
		return val == 0
	case int8:
		return val == 0
	case int16:
		return val == 0
	case int32:
		return val == 0
	case int64:
		return val == 0
	case uint:
		return val == 0
	case uint8:
		return val == 0
	case uint16:
		return val == 0
	case uint32:
		return val == 0
	case uint64:
		return val == 0
	case float32:
		return val == 0
	case float64:
		return val == 0
	case string:
		return val == ""
	case bool:
		return !val
	case *int, *int8, *int16, *int32, *int64,
		*uint, *uint8, *uint16, *uint32, *uint64,
		*float32, *float64, *string, *bool:
		return val == nil
	default:
		return reflect.ValueOf(f.Value).IsZero()
	}
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

type fieldInfo struct {
	pathPtr *[]string // Full hierarchical path: ["parent", "child"]
	offset  uintptr
}

type schema struct {
	fields []fieldInfo
	TagKey string
}

var genericSchemaCache = sync.Map{}

// fieldHeader matches initial memory layout Field
type fieldHeader struct {
	path *[]string
}

func Link[T any](s *T, tagKey string) {
	ptr := unsafe.Pointer(s)

	var zero T
	tVal := reflect.TypeOf(zero)

	// get cached schema or build it
	var sch *schema
	if cached, ok := genericSchemaCache.Load(tVal); ok {
		sch = cached.(*schema)
	} else {

		if tVal.Kind() != reflect.Struct {
			return
		}

		sch = buildSchema(tVal, tagKey)
		genericSchemaCache.Store(tVal, sch)

	}

	// Note:
	// breaking change: no longer tagkey is checked, assumes the schema is built with the correct tagkey

	// link all Field[T] path pointers
	for _, field := range sch.fields {
		(*fieldHeader)(unsafe.Pointer(uintptr(ptr) + field.offset)).path = field.pathPtr
	}

}

func buildSchema(tVal reflect.Type, tagKey string) *schema {
	var fields []fieldInfo
	collectFields(tVal, tagKey, 0, nil, &fields)
	return &schema{
		fields: fields,
		TagKey: tagKey,
	}
}

// collectFields recursively collects all Field[T] fields with absolute offsets
func collectFields(tVal reflect.Type, tagKey string, baseOffset uintptr, parentPath []string, fields *[]fieldInfo) {
	sliceStringPtrType := reflect.TypeOf((*[]string)(nil))

	for i := 0; i < tVal.NumField(); i++ {
		field := tVal.Field(i)

		// skip unexported fields
		if !field.IsExported() {
			continue
		}

		// skip fields with tag "-"
		tagName := strings.Split(field.Tag.Get(tagKey), ",")[0]
		if tagName == "-" {
			continue
		}

		// check for Field[T] pattern
		if field.Type.Kind() == reflect.Struct && field.Type.NumField() > 0 {
			firstField := field.Type.Field(0)
			if firstField.Type == sliceStringPtrType && firstField.Name == "path" {
				// Found a Field[T]
				n := strings.Split(field.Tag.Get(tagKey), ",")[0]
				if n == "" {
					n = field.Name
				}

				// Build hierarchical path as slice
				var currentPath []string
				if len(parentPath) > 0 {
					currentPath = make([]string, len(parentPath)+1)
					copy(currentPath, parentPath)
					currentPath[len(parentPath)] = n
				} else {
					currentPath = []string{n}
				}

				// Allocate path slice on heap to ensure it persists
				pathPtr := new([]string)
				*pathPtr = currentPath

				// Add to flat list with absolute offset
				*fields = append(*fields, fieldInfo{
					pathPtr: pathPtr,
					offset:  baseOffset + field.Offset,
				})

				// Check if Value is a struct that might contain more Field[T] fields
				if field.Type.NumField() >= 2 {
					valueField := field.Type.Field(1) // Value is at index 1 (path=0, Value=1)
					if valueField.Name == "Value" && valueField.Type.Kind() == reflect.Struct {
						// Recursively collect fields from nested struct, passing current path
						nestedBaseOffset := baseOffset + field.Offset + valueField.Offset
						collectFields(valueField.Type, tagKey, nestedBaseOffset, currentPath, fields)
					}
				}
			}
		}
	}
}
