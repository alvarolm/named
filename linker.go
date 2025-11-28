package named

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type Field[T any] struct {
	name  *string // goes first so it's aligned with fildHeader
	Value T
}

func (f *Field[T]) Name() string {
	if f.name == nil {
		return ""
	}
	return *f.name
}

func (f *Field[T]) NoName() bool {
	return (f.name == nil || *f.name == "")
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
	tagPtr *string
	offset uintptr
}

type schema struct {
	fields []fieldInfo
	TagKey string
}

var genericSchemaCache = sync.Map{}

// fieldHeader matches initial memory layout Field
type fieldHeader struct {
	name *string
}

func Link[T any](s *T, tagKey string) {
	ptr := unsafe.Pointer(s)

	var zero T
	tVal := reflect.TypeOf(zero)

	linkGenericReflect(ptr, tVal, tagKey)
}

// linkGenericReflect is a helper that recursively links a struct given its reflect.Type
func linkGenericReflect(ptr unsafe.Pointer, tVal reflect.Type, tagKey string) {

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

	// link all Field[T] name pointers (flat iteration, no recursion)
	for _, field := range sch.fields {
		fieldPtr := unsafe.Pointer(uintptr(ptr) + field.offset)
		(*fieldHeader)(fieldPtr).name = field.tagPtr
	}

}

func buildSchema(tVal reflect.Type, tagKey string) *schema {
	var fields []fieldInfo
	collectFields(tVal, tagKey, 0, &fields)
	return &schema{
		fields: fields,
		TagKey: tagKey,
	}
}

// collectFields recursively collects all Field[T] fields with absolute offsets
func collectFields(tVal reflect.Type, tagKey string, baseOffset uintptr, fields *[]fieldInfo) {
	stringPtrType := reflect.TypeOf((*string)(nil))

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
			if firstField.Type == stringPtrType && firstField.Name == "name" {
				// Found a Field[T]
				n := strings.Split(field.Tag.Get(tagKey), ",")[0]
				if n == "" {
					n = field.Name
				}

				// Add to flat list with absolute offset
				*fields = append(*fields, fieldInfo{
					tagPtr: &n,
					offset: baseOffset + field.Offset,
				})

				// Check if Value is a struct that might contain more Field[T] fields
				if field.Type.NumField() >= 2 {
					secondField := field.Type.Field(1)
					if secondField.Name == "Value" && secondField.Type.Kind() == reflect.Struct {
						// Recursively collect fields from nested struct
						nestedBaseOffset := baseOffset + field.Offset + secondField.Offset
						collectFields(secondField.Type, tagKey, nestedBaseOffset, fields)
					}
				}
			}
		}
	}
}
