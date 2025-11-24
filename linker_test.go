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

	// Reuse all tests from TestLink_Simple for the embedded struct
	testSampleSimple(t, &s.B.Value, "B.Value.")
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
