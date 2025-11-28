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
		if tt.field.name == nil || *tt.field.name != tt.expected {
			t.Errorf("%s%s: Expected name to be '%s', got %v", prefix, tt.name, tt.expected, *tt.field.name)
		}
	}

	// Test other types
	if s.F.name == nil || *s.F.name != "f" {
		t.Errorf("%sF: Expected name to be 'f', got %v", prefix, s.F.name)
	}

	if s.G.name == nil || *s.G.name != "g" {
		t.Errorf("%sG: Expected name to be 'g', got %v", prefix, s.G.name)
	}

	if s.H.name == nil || *s.H.name != "h" {
		t.Errorf("%sH: Expected name to be 'h', got %v", prefix, s.H.name)
	}

	if s.I.name == nil || *s.I.name != "i" {
		t.Errorf("%sI: Expected name to be 'i', got %v", prefix, s.I.name)
	}

	if s.K.name == nil || *s.K.name != "k" {
		t.Errorf("%sK: Expected name to be 'k', got %v", prefix, s.K.name)
	}

	if s.J.name == nil || *s.J.name != "J" {
		t.Errorf("%sJ: Expected name to be 'J', got %v", prefix, s.J.name)
	}

	// L should be skipped due to json:"-"
	if s.L.name != nil {
		t.Errorf("%sL: Expected name to be nil (skipped), got %v", prefix, *s.L.name)
	}

	// m should be skipped because it's unexported
	if s.m.name != nil {
		t.Errorf("%sm: Expected name to be nil (unexported field), got %v", prefix, *s.m.name)
	}

}

func TestLink_Simple(t *testing.T) {
	s := SampleSimple{}
	Link(&s, "json")
	testSampleSimple(t, &s, "")
}

type SampleEmbedStruct struct {
	A Field[string]       `json:"x"`
	B Field[SampleSimple] `json:"y"`
}

func TestLink_Embedded(t *testing.T) {
	s := SampleEmbedStruct{}
	Link(&s, "json")

	if s.A.name == nil || *s.A.name != "x" {
		t.Errorf("Expected s.A.name to be 'x', got %v", s.A.name)
	}

	if s.B.name == nil || *s.B.name != "y" {
		t.Errorf("Expected s.B.name to be 'y', got %v", s.B.name)
	}

	// Test nested fields have parent prefix prepended
	tests := []struct {
		name     string
		field    *Field[int]
		expected string
	}{
		{"A", (*Field[int])(unsafe.Pointer(&s.B.Value.A)), "y.a"},
		{"C", &s.B.Value.C, "y.c"},
		{"E", &s.B.Value.E, "y.e"},
		{"J (no tag)", &s.B.Value.J, "y.J"},
	}

	for _, tt := range tests {
		if tt.field.name == nil || *tt.field.name != tt.expected {
			t.Errorf("B.Value.%s: Expected name to be '%s', got %v", tt.name, tt.expected, *tt.field.name)
		}
	}

	// Test other types with parent prefix
	if s.B.Value.F.name == nil || *s.B.Value.F.name != "y.f" {
		t.Errorf("B.Value.F: Expected name to be 'y.f', got %v", *s.B.Value.F.name)
	}

	if s.B.Value.G.name == nil || *s.B.Value.G.name != "y.g" {
		t.Errorf("B.Value.G: Expected name to be 'y.g', got %v", *s.B.Value.G.name)
	}

	if s.B.Value.H.name == nil || *s.B.Value.H.name != "y.h" {
		t.Errorf("B.Value.H: Expected name to be 'y.h', got %v", *s.B.Value.H.name)
	}

	if s.B.Value.I.name == nil || *s.B.Value.I.name != "y.i" {
		t.Errorf("B.Value.I: Expected name to be 'y.i', got %v", *s.B.Value.I.name)
	}

	if s.B.Value.K.name == nil || *s.B.Value.K.name != "y.k" {
		t.Errorf("B.Value.K: Expected name to be 'y.k', got %v", *s.B.Value.K.name)
	}

	// L should be skipped due to json:"-"
	if s.B.Value.L.name != nil {
		t.Errorf("B.Value.L: Expected name to be nil (skipped), got %v", *s.B.Value.L.name)
	}

	// m should be skipped because it's unexported
	if s.B.Value.m.name != nil {
		t.Errorf("B.Value.m: Expected name to be nil (unexported field), got %v", *s.B.Value.m.name)
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

/*

func TestParentField_Simple(t *testing.T) {
	type Inner struct {
		A Field[int]    `json:"a"`
		B Field[string] `json:"b"`
	}

	type Outer struct {
		X Field[int]   `json:"x"`
		Y Field[Inner] `json:"y"`
	}

	s := Outer{}
	Link(&s, "json")

	// Root fields should have nil parent
	if s.X.Parent() != nil {
		t.Errorf("Root field X should have nil parent")
	}
	if s.Y.Parent() != nil {
		t.Errorf("Root field Y should have nil parent")
	}

	// Nested fields should have parent pointing to Y
	if s.Y.Value.A.Parent() == nil {
		t.Errorf("Nested field A should have parent")
	}
	if s.Y.Value.A.Parent().Name() != "y" {
		t.Errorf("Expected parent name 'y', got '%s'", s.Y.Value.A.Parent().Name())
	}
	if s.Y.Value.B.Parent() == nil {
		t.Errorf("Nested field B should have parent")
	}
	if s.Y.Value.B.Parent().Name() != "y" {
		t.Errorf("Expected parent name 'y', got '%s'", s.Y.Value.B.Parent().Name())
	}
}

func TestParentField_MultipleNesting(t *testing.T) {
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
	Link(&s, "json")

	// Root level should have nil parent
	if s.Top.Parent() != nil {
		t.Errorf("Root field Top should have nil parent")
	}

	// Check level 1 -> level 2 parent
	if s.Top.Value.Mid.Parent() == nil {
		t.Errorf("Mid field should have parent")
	}
	if s.Top.Value.Mid.Parent().Name() != "top" {
		t.Errorf("Expected parent name 'top', got '%s'", s.Top.Value.Mid.Parent().Name())
	}

	// Check level 2 -> level 3 parent
	if s.Top.Value.Mid.Value.Deep.Parent() == nil {
		t.Errorf("Deep field should have parent")
	}
	if s.Top.Value.Mid.Value.Deep.Parent().Name() != "mid" {
		t.Errorf("Expected parent name 'mid', got '%s'", s.Top.Value.Mid.Value.Deep.Parent().Name())
	}
}

func TestParentField_ChainTraversal(t *testing.T) {
	type Inner struct {
		Value Field[int] `json:"value"`
	}
	type Outer struct {
		Container Field[Inner] `json:"container"`
	}

	s := Outer{}
	Link(&s, "json")

	// Traverse parent chain
	field := &s.Container.Value.Value
	chain := []string{}

	for current := field.Parent(); current != nil; current = current.Parent() {
		chain = append(chain, current.Name())
	}

	expected := []string{"container"}
	if len(chain) != len(expected) {
		t.Errorf("Expected chain length %d, got %d", len(expected), len(chain))
	}
	for i, name := range expected {
		if chain[i] != name {
			t.Errorf("Expected chain[%d] = '%s', got '%s'", i, name, chain[i])
		}
	}
}

func TestParentField_EmptyNested(t *testing.T) {
	type Empty struct {
		// No fields
	}
	type Container struct {
		E Field[Empty] `json:"e"`
	}

	s := Container{}
	Link(&s, "json")

	// Should not panic
	if s.E.Parent() != nil {
		t.Errorf("Root field should have nil parent")
	}
}
*/
