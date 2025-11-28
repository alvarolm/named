package named

import "testing"

type Sample5Fields struct {
	A Field[int]     `json:"a"`
	B Field[string]  `json:"b"`
	C Field[float64] `json:"c"`
	D Field[any]     `json:"d"`
	E Field[uint64]  `json:"e"`
}

func BenchmarkLikner_5Fields(b *testing.B) {
	// Warm up cache
	x := Sample5Fields{}
	Link(&x, "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := Sample5Fields{}
		Link(&s, "json")
	}
}

func BenchmarkLikner_Simple(b *testing.B) {
	// Warm up cache
	x := SampleSimple{}
	Link(&x, "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := SampleSimple{}
		Link(&s, "json")
	}
}

func BenchmarkLikner_Embedded(b *testing.B) {
	// Warm up cache
	x := SampleEmbedStruct{}
	Link(&x, "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := SampleEmbedStruct{}
		Link(&s, "json")
	}
}

func BenchmarkLinkerNoValue_Int(b *testing.B) {
	f := Field[int]{Value: 0}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}

func BenchmarkLinkerNoValue_IntNonZero(b *testing.B) {
	f := Field[int]{Value: 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}

func BenchmarkLinkerNoValue_String(b *testing.B) {
	f := Field[string]{Value: ""}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}

func BenchmarkLinkerNoValue_StringNonZero(b *testing.B) {
	f := Field[string]{Value: "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}

func BenchmarkLinkerNoValue_Pointer(b *testing.B) {
	f := Field[*string]{Value: nil}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}

func BenchmarkLinkerNoValue_PointerNonZero(b *testing.B) {
	s := "test"
	f := Field[*string]{Value: &s}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}

func BenchmarkLinkerNoValue_Struct(b *testing.B) {
	type MyStruct struct {
		A int
		B string
	}
	f := Field[MyStruct]{Value: MyStruct{}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}

func BenchmarkLinkerNoValue_StructNonZero(b *testing.B) {
	type MyStruct struct {
		A int
		B string
	}
	f := Field[MyStruct]{Value: MyStruct{A: 1, B: "test"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
	}
}
