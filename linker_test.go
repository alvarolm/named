package named

import (
	"encoding/json"
	"testing"
	"unsafe"
)

type SampleSimple struct {
	A Field[int]     `json:"a"`
	B string         `json:"b"` // non-Field, should be skipped
	C Field[int]     `json:"c"`
	D float64        `json:"d"` // non-Field, should be skipped
	E Field[int]     `json:"e"`
	F Field[string]  `json:"f"`
	G Field[float64] `json:"g"`
	H Field[any]     `json:"h"`
	I Field[uint64]  `json:"i"`
	K Field[any]     `json:"k"`
	L Field[int]     `json:"-"` // should be skipped
	J Field[int]     // should use field name if no tag is present
	m Field[any]     // unexported, should be skipped
}

func testSampleSimple(t *testing.T, s *SampleSimple, prefix string) {
	tests := []struct {
		name     string
		field    *Field[int]
		expected string
	}{
		{"A", (*Field[int])(unsafe.Pointer(&s.A)), "a"},
		{"C", &s.C, "c"},
		{"E", &s.E, "e"},
		{"J (no tag)", &s.J, "J"},
	}

	for _, tt := range tests {
		if tt.field.Name() != tt.expected {
			t.Errorf("%s%s: Expected name to be '%s', got '%s'", prefix, tt.name, tt.expected, tt.field.Name())
		}
	}

	// Test other types
	if s.F.Name() != "f" {
		t.Errorf("%sF: Expected name to be 'f', got '%s'", prefix, s.F.Name())
	}

	if s.G.Name() != "g" {
		t.Errorf("%sG: Expected name to be 'g', got '%s'", prefix, s.G.Name())
	}

	if s.H.Name() != "h" {
		t.Errorf("%sH: Expected name to be 'h', got '%s'", prefix, s.H.Name())
	}

	if s.I.Name() != "i" {
		t.Errorf("%sI: Expected name to be 'i', got '%s'", prefix, s.I.Name())
	}

	if s.K.Name() != "k" {
		t.Errorf("%sK: Expected name to be 'k', got '%s'", prefix, s.K.Name())
	}

	if s.J.Name() != "J" {
		t.Errorf("%sJ: Expected name to be 'J', got '%s'", prefix, s.J.Name())
	}

	// L should be skipped due to json:"-"
	if s.L.path != nil {
		t.Errorf("%sL: Expected path to be nil (skipped), got %v", prefix, s.L.path)
	}

	// m should be skipped because it's unexported
	if s.m.path != nil {
		t.Errorf("%sm: Expected path to be nil (unexported field), got %v", prefix, s.m.path)
	}

}

func TestLink_Simple(t *testing.T) {
	s := SampleSimple{}
	Link(&s)
	testSampleSimple(t, &s, "")
}

type SampleEmbedStruct struct {
	A Field[string]       `json:"x"`
	B Field[SampleSimple] `json:"y"`
}

func init() {
	// Preload schemas for benchmarks
	LoadLink[Sample5Fields]("json")
	LoadLink[SampleSimple]("json")
	LoadLink[SampleEmbedStruct]("json")
}

func TestLink_Embedded(t *testing.T) {
	s := SampleEmbedStruct{}
	Link(&s)

	// Test root-level fields
	if s.A.Name() != "x" {
		t.Errorf("Expected s.A.Name() to be 'x', got '%s'", s.A.Name())
	}

	if s.B.Name() != "y" {
		t.Errorf("Expected s.B.Name() to be 'y', got '%s'", s.B.Name())
	}

	// Test nested fields - Name() now returns leaf only
	tests := []struct {
		name         string
		field        *Field[int]
		expectedLeaf string
		expectedFull string
	}{
		{"A", (*Field[int])(unsafe.Pointer(&s.B.Value.A)), "a", "y.a"},
		{"C", &s.B.Value.C, "c", "y.c"},
		{"E", &s.B.Value.E, "e", "y.e"},
		{"J (no tag)", &s.B.Value.J, "J", "y.J"},
	}

	for _, tt := range tests {
		if tt.field.Name() != tt.expectedLeaf {
			t.Errorf("B.Value.%s: Expected Name() to be '%s', got '%s'", tt.name, tt.expectedLeaf, tt.field.Name())
		}
		if tt.field.FullName("") != tt.expectedFull {
			t.Errorf("B.Value.%s: Expected FullName() to be '%s', got '%s'", tt.name, tt.expectedFull, tt.field.FullName(""))
		}
	}

	// Test other types with nested paths
	if s.B.Value.F.Name() != "f" {
		t.Errorf("B.Value.F: Expected Name() to be 'f', got '%s'", s.B.Value.F.Name())
	}
	if s.B.Value.F.FullName("") != "y.f" {
		t.Errorf("B.Value.F: Expected FullName() to be 'y.f', got '%s'", s.B.Value.F.FullName(""))
	}

	if s.B.Value.G.Name() != "g" {
		t.Errorf("B.Value.G: Expected Name() to be 'g', got '%s'", s.B.Value.G.Name())
	}
	if s.B.Value.G.FullName("") != "y.g" {
		t.Errorf("B.Value.G: Expected FullName() to be 'y.g', got '%s'", s.B.Value.G.FullName(""))
	}

	if s.B.Value.H.Name() != "h" {
		t.Errorf("B.Value.H: Expected Name() to be 'h', got '%s'", s.B.Value.H.Name())
	}
	if s.B.Value.H.FullName("") != "y.h" {
		t.Errorf("B.Value.H: Expected FullName() to be 'y.h', got '%s'", s.B.Value.H.FullName(""))
	}

	if s.B.Value.I.Name() != "i" {
		t.Errorf("B.Value.I: Expected Name() to be 'i', got '%s'", s.B.Value.I.Name())
	}
	if s.B.Value.I.FullName("") != "y.i" {
		t.Errorf("B.Value.I: Expected FullName() to be 'y.i', got '%s'", s.B.Value.I.FullName(""))
	}

	if s.B.Value.K.Name() != "k" {
		t.Errorf("B.Value.K: Expected Name() to be 'k', got '%s'", s.B.Value.K.Name())
	}
	if s.B.Value.K.FullName("") != "y.k" {
		t.Errorf("B.Value.K: Expected FullName() to be 'y.k', got '%s'", s.B.Value.K.FullName(""))
	}

	// L should be skipped due to json:"-"
	if s.B.Value.L.path != nil {
		t.Errorf("B.Value.L: Expected path to be nil (skipped), got %v", s.B.Value.L.path)
	}

	// m should be skipped because it's unexported
	if s.B.Value.m.path != nil {
		t.Errorf("B.Value.m: Expected path to be nil (unexported field), got %v", s.B.Value.m.path)
	}
}

func TestLink_OmitZero(t *testing.T) {

	type sample struct {
		F1 Field[int]    `json:"f1,omitzero"`
		F2 Field[string] `json:"f2"`
	}

	data, err := json.Marshal(sample{
		F1: Field[int]{Value: 0},
		F2: Field[string]{Value: "test"},
	})

	if err != nil {
		t.Fatalf("Unexpected error during marshaling: %v", err)
	}

	expected := `{"f2":"test"}`
	if string(data) != expected {
		t.Errorf("Expected JSON: %s, got: %s", expected, string(data))
	}
}

func TestField_NameReturnsLeafOnly(t *testing.T) {
	type Inner struct {
		A Field[int] `json:"a"`
	}
	type Outer struct {
		Y Field[Inner] `json:"y"`
	}

	LoadLink[Outer]("json")

	s := Outer{}
	Link(&s)

	// Nested field returns leaf name only (not "y.a")
	if s.Y.Value.A.Name() != "a" {
		t.Errorf("Expected Name() to return 'a', got '%s'", s.Y.Value.A.Name())
	}

	// Use FullName() for old behavior
	if s.Y.Value.A.FullName("") != "y.a" {
		t.Errorf("Expected FullName() to return 'y.a', got '%s'", s.Y.Value.A.FullName(""))
	}
}

func TestField_Path(t *testing.T) {
	type Level3 struct {
		Deep Field[int] `json:"deep"`
	}
	type Level2 struct {
		Mid Field[Level3] `json:"mid"`
	}
	type Level1 struct {
		Top Field[Level2] `json:"top"`
	}

	s := Level1{}
	LoadLink[Level1]("json")

	Link(&s)

	// Root level
	path := s.Top.Path()
	if len(path) != 1 || path[0] != "top" {
		t.Errorf("Expected FullPath() ['top'], got %v", path)
	}

	// Second level
	path = s.Top.Value.Mid.Path()
	if len(path) != 2 || path[0] != "top" || path[1] != "mid" {
		t.Errorf("Expected FullPath() ['top', 'mid'], got %v", path)
	}

	// Third level
	path = s.Top.Value.Mid.Value.Deep.Path()
	if len(path) != 3 || path[0] != "top" || path[1] != "mid" || path[2] != "deep" {
		t.Errorf("Expected FullPath() ['top', 'mid', 'deep'], got %v", path)
	}

}

func TestField_FullName_BackwardCompatibility(t *testing.T) {
	type Inner struct {
		A Field[int] `json:"a"`
		B Field[int] `json:"b"`
	}
	type Outer struct {
		X Field[int]   `json:"x"`
		Y Field[Inner] `json:"y"`
	}

	s := Outer{}
	LoadLink[Outer]("json")

	Link(&s)

	// Root field with default separator
	if s.X.FullName("") != "x" {
		t.Errorf("Expected FullName() 'x', got '%s'", s.X.FullName(""))
	}

	// Nested fields with default separator
	if s.Y.Value.A.FullName("") != "y.a" {
		t.Errorf("Expected FullName() 'y.a', got '%s'", s.Y.Value.A.FullName(""))
	}

	if s.Y.Value.B.FullName("") != "y.b" {
		t.Errorf("Expected FullName() 'y.b', got '%s'", s.Y.Value.B.FullName(""))
	}

	// Test custom separator
	if s.Y.Value.A.FullName("/") != "y/a" {
		t.Errorf("Expected FullName('/') 'y/a', got '%s'", s.Y.Value.A.FullName("/"))
	}

	if s.Y.Value.B.FullName("->") != "y->b" {
		t.Errorf("Expected FullName('->') 'y->b', got '%s'", s.Y.Value.B.FullName("->"))
	}
}

func TestFieldMemoryLayout(t *testing.T) {
	f := Field[int]{}

	// Verify first field is at offset 0
	pathOffset := unsafe.Offsetof(f.path)
	if pathOffset != 0 {
		t.Errorf("path field should be at offset 0, got %d", pathOffset)
	}

	// Verify Value field is at offset 8 (after one pointer)
	valueOffset := unsafe.Offsetof(f.Value)
	if valueOffset != 8 {
		t.Errorf("Value field should be at offset 8, got %d", valueOffset)
	}

	// Verify fieldHeader matches Field[T] layout (should be 8 bytes - one pointer)
	if unsafe.Sizeof(fieldHeader{}) != 8 {
		t.Errorf("fieldHeader should be 8 bytes, got %d", unsafe.Sizeof(fieldHeader{}))
	}
}
