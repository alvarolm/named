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
	nested *nestedStructInfo
}

type schema struct {
	fields []fieldInfo
	TagKey string
}

type nestedStructInfo struct {
	valueOffset uintptr
	valueType   reflect.Type
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

	if sch.TagKey != tagKey {
		panic("Mismatched tag key in schema cache")
	}

	// link all Field[T] name pointers and recursively link nested structs
	for i := range sch.fields {
		field := &sch.fields[i]
		fieldPtr := unsafe.Pointer(uintptr(ptr) + field.offset)
		(*fieldHeader)(fieldPtr).name = field.tagPtr
		/*
			// Handle nested structs inline
			if field.nested != nil {
				valuePtr := unsafe.Pointer(uintptr(ptr) + field.nested.valueOffset)
				linkGenericReflect(valuePtr, field.nested.valueType, tagKey)
			}
		*/
	}

}

func buildSchema(tVal reflect.Type, tagKey string) *schema {
	var fields []fieldInfo
	stringPtrType := reflect.TypeOf((*string)(nil))

	for i := 0; i < tVal.NumField(); i++ {
		field := tVal.Field(i)

		// skip unexported fields
		if !field.IsExported() {
			continue
		}

		// skip fields with json:"-"
		tagName := strings.Split(field.Tag.Get(tagKey), ",")[0]
		if tagName == "-" {
			continue
		}

		// check for matching layout (Field[T])
		if field.Type.Kind() == reflect.Struct && field.Type.NumField() > 0 {
			firstField := field.Type.Field(0)
			if firstField.Type == stringPtrType && firstField.Name == "name" {
				// Found a Field[T], create tag name
				n := strings.Split(field.Tag.Get(tagKey), ",")[0]
				if n == "" {
					n = field.Name
				}

				// Create fieldInfo
				fInfo := fieldInfo{
					tagPtr: &n,
					offset: field.Offset,
					nested: nil,
				}

				// Check if this Field[T] has a struct Value needing recursive linking
				if field.Type.NumField() >= 2 {
					secondField := field.Type.Field(1)
					if secondField.Name == "Value" && secondField.Type.Kind() == reflect.Struct {
						fInfo.nested = &nestedStructInfo{
							valueOffset: field.Offset + secondField.Offset,
							valueType:   secondField.Type,
						}
					}
				}

				fields = append(fields, fInfo)
			}
		}
	}

	return &schema{
		fields: fields,
		TagKey: tagKey,
	}
}
