package named

import (
	"errors"
	"reflect"
	"strings"
	"unsafe"
)

type fieldInfo struct {
	pathPtr *[]string // Full hierarchical path: ["parent", "child"]
	offset  uintptr
}

type schema struct {
	fields []fieldInfo
	TagKey string
}

var cachedSchemaMap = make(map[uintptr]*schema)

// emptyInterface mimics the internal memory layout of a Go empty interface (any).
// In the standard Go runtime, an interface is a pair of pointers: {type, data}.
//
// By casting a pointer to an interface variable to (*emptyInterface), we can
// directly access the underlying type pointer to use as a unique hash key.
type emptyInterface struct {
	typ unsafe.Pointer
	ptr unsafe.Pointer
}

// LoadLink generates and loads the schema for type T using the specified tagKey.
// The generated schema is cached for future Link calls. T must be a struct type.
// not async safe, should be called before any Link calls.
func LoadLink[T any](tagKey string) error {
	var zero T
	tVal := reflect.TypeOf(zero)

	if tVal.Kind() != reflect.Struct {
		return errors.New("CacheSchema: T must be a struct type")
	}

	// Get type ID for fast lookup
	var gen any = (*T)(nil)
	typeID := uintptr((*emptyInterface)(unsafe.Pointer(&gen)).typ)

	// Build schema
	var sch *schema
	{
		var fields []fieldInfo
		collectFields(tVal, tagKey, 0, nil, &fields)
		sch = &schema{
			fields: fields,
			TagKey: tagKey,
		}
	}

	// Cache schema
	cachedSchemaMap[typeID] = sch

	return nil
}

// Link populates all Field[T] fields in the struct pointed to by s with their path information.
// T must be a struct type previously registered with LoadLink.
// returns true if linking was successful, false otherwise.
func Link[T any](s *T) bool {

	ptr := unsafe.Pointer(s)

	var gen any = (*T)(nil)
	typeID := uintptr((*emptyInterface)(unsafe.Pointer(&gen)).typ)

	// load from cache
	sch, ok := cachedSchemaMap[typeID]
	if !ok {
		return false
	}

	// Note:
	// breaking change: no longer tagkey is checked, assumes the schema is built with the correct tagkey

	// link all Field[T] path pointers
	for _, field := range sch.fields {
		/*
			fp := (*fieldHeader)(unsafe.Pointer(uintptr(ptr) + field.offset))
			fp.path = field.pathPtr
			fp.parentPath = nil
		*/
		(*fieldHeader)(unsafe.Pointer(uintptr(ptr) + field.offset)).path = field.pathPtr
	}

	return true
}

type fieldRefs struct {
	paths       *[]string
	parentPaths *[]string
}

//func (fr *fieldRefs) GetPaths() {

func LinkWithPath[T any](s *T, path []string) bool {
	ptr := unsafe.Pointer(s)

	var gen any = (*T)(nil)
	typeID := uintptr((*emptyInterface)(unsafe.Pointer(&gen)).typ)

	// load from cache
	sch, ok := cachedSchemaMap[typeID]
	if !ok {
		return false
	}

	// Note:
	// breaking change: no longer tagkey is checked, assumes the schema is built with the correct tagkey

	// link all Field[T] path pointers
	for _, field := range sch.fields {
		fp := (*fieldHeader)(unsafe.Pointer(uintptr(ptr) + field.offset))
		fp.path = field.pathPtr
		//fp.parentPath = &path	}

	return true
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
				if field.Type.NumField() >= 3 {
					valueField := field.Type.Field(2) // Value is at index 2 (path=0, parentPath=1, Value=2)
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
