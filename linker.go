package named

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type FieldBasics interface {
	Name() string
	Parent() string
}

type Field[T any] struct {
	name        *string     // goes first so it's aligned with fildHeader
	parentField FieldBasics // optional, for nested fields
	Value       T
}

func (f *Field[T]) Parent() FieldBasics {
	return f.parentField
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

type schema struct {
	tagPtrs       []*string
	offsets       []uintptr
	nestedStructs []nestedStructInfo
	TagKey        string
}

type nestedStructInfo struct {
	valueOffset uintptr
	valueType   reflect.Type
}

var genericSchemaCache = sync.Map{}

// fieldHeader matches initial memory layout Field
type fieldHeader struct {
	name        *string
	parentField FieldBasics
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

	// link all Field[T] name pointers
	offsets := sch.offsets
	tagPtrs := sch.tagPtrs
	for i := range offsets {
		fieldPtr := unsafe.Pointer(uintptr(ptr) + offsets[i])
		(*fieldHeader)(fieldPtr).name = tagPtrs[i]
	}

	// recursively link nested structs
	for _, nested := range sch.nestedStructs {
		valuePtr := unsafe.Pointer(uintptr(ptr) + nested.valueOffset)
		linkGenericReflect(valuePtr, nested.valueType, tagKey)
	}

}

func buildSchema(tVal reflect.Type, tagKey string) *schema {
	var matchingIndices []int
	var nestedStructs []nestedStructInfo
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

		// check for matching layout
		if field.Type.Kind() == reflect.Struct && field.Type.NumField() > 0 {
			firstField := field.Type.Field(0)
			if firstField.Type == stringPtrType && firstField.Name == "name" {
				matchingIndices = append(matchingIndices, i)

				// check if this Field[T] has a struct Value that needs recursive linking
				if field.Type.NumField() >= 2 {
					secondField := field.Type.Field(1)
					if secondField.Name == "Value" && secondField.Type.Kind() == reflect.Struct {
						nestedStructs = append(nestedStructs, nestedStructInfo{
							valueOffset: field.Offset + secondField.Offset,
							valueType:   secondField.Type,
						})
					}
				}
			}
		}
	}

	// build schema for matching fields
	sch := &schema{
		tagPtrs:       make([]*string, len(matchingIndices)),
		offsets:       make([]uintptr, len(matchingIndices)),
		nestedStructs: nestedStructs,
		TagKey:        tagKey,
	}

	for idx, i := range matchingIndices {
		field := tVal.Field(i)
		n := strings.Split(field.Tag.Get(tagKey), ",")[0]
		if n == "" {
			n = field.Name
		}
		sch.tagPtrs[idx] = &n
		sch.offsets[idx] = field.Offset
	}

	return sch
}
