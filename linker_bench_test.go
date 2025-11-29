package named

import (
	"testing"
)

type Sample5Fields struct {
	A Field[int]     `json:"a"`
	B Field[string]  `json:"b"`
	C Field[float64] `json:"c"`
	D Field[any]     `json:"d"`
	E Field[uint64]  `json:"e"`
}

func init() {
	LoadLink[Sample5Fields]("json")
}

func BenchmarkLinker_5Fields(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := Sample5Fields{}
		Link(&s)
	}
}

func BenchmarkLinker_Simple(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := SampleSimple{}
		Link(&s)
	}
}

func BenchmarkLinker_Embedded(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := SampleEmbedStruct{}
		Link(&s)
	}
}

func BenchmarkLinkerWithPath_5Fields_2Levels(b *testing.B) {
	path := []string{"level1", "level2"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := Sample5Fields{}
		LinkWithPath(&s, &path)
	}
}

func BenchmarkLinkerBasic_NameCall(b *testing.B) {
	type MyStruct struct {
		A Field[int]
		B Field[string]
	}
	s := MyStruct{}
	LoadLink[MyStruct]("json")
	Link(&s)
	//b.Logf("Name(): %v", s.A.Name())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.A.Name()
	}
}

func BenchmarkLinkerBasic_FullNameCall(b *testing.B) {
	type Inner struct {
		X Field[int]
	}
	type MyStruct struct {
		A Field[int]
		B Field[Inner]
	}
	s := MyStruct{}
	LoadLink[MyStruct]("json")
	Link(&s)
	//b.Logf("FullName(): %v", s.B.FullName("."))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.B.Value.X.FullName(".")
	}
}

func BenchmarkLinkerBasic_PathCall(b *testing.B) {
	type MyStruct struct {
		A Field[int]
		B Field[string]
	}
	s := MyStruct{}
	LoadLink[MyStruct]("json")
	Link(&s)
	//b.Logf("Path(): %v", s.B.Path())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.B.Path()
	}
}

func BenchmarkLinkerWithPathBasic_FullNameCall(b *testing.B) {
	type MyStruct struct {
		A Field[int]
		B Field[string]
	}
	s := MyStruct{}
	LoadLink[MyStruct]("json")
	pp := []string{"level1"}
	LinkWithPath(&s, &pp)
	//b.Logf("FullName(): %v", s.B.FullName("."))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.B.FullName(".")
	}
}

func BenchmarkLinkerWithPathBasic_PathCall(b *testing.B) {
	type MyStruct struct {
		A Field[int]
		B Field[string]
	}
	s := MyStruct{}
	pp := []string{"level1"}
	LoadLink[MyStruct]("json")
	LinkWithPath(&s, &pp)
	//b.Logf("Path(): %v", s.B.Path())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.B.Path()
	}
}

func BenchmarkLinkerNoValue_SliceOfStructs(b *testing.B) {
	type MyStruct struct {
		A int
		B string
	}
	f := FieldSlice[[]MyStruct, MyStruct]{Value: []MyStruct{}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = f.NoValue()
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
