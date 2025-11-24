[![GoDoc](https://godoc.org/github.com/username/repo?status.svg)](https://godoc.org/github.com/alvarolm/named)

# named

Retrieve struct field names easily!

In Go, structs are commonly used to represent structured data that is directly encoded into popular formats like JSON, YAML and others, these formats have field names that are easy to set but a challenge to retrieve them.

I could not find a way to do this that was also:

- (A) ergonomic: mainly provide type safety and autocompletion, dont want to make mistakes while typing.
- (B) performant enough: there is no point if its too expensive.

## Runtime solution:
At first I tried a runtime solution to see how further I could go with it.

Decided to extend from the field value: a struct can hold the value itself and a reference to the field name.

To resolve A:
- for the field type value: a type parameter, type safe.
- to retrieve the name: a method (exposes the name but not the reference), autocompletes.

this is what I ended up using

```go
type Field[T any] struct {
	Value T
	name  *string
}

func (f *Field[T]) Name() string {
	if f.name == nil {
		return ""
	}
	return *f.name
}
```

to resolve B:

since the structs are immutable a cache was used, the first run generates the schemas with the fields names, and subsequent runs just retrieve the schema and assign the name pointer to each Field struct:

```bash
goos: linux
goarch: amd64
pkg: namedtest
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkReflection-8           11680456               110.7 ns/op            64 B/op          1 allocs/op
BenchmarkManual-8               1000000000               0.2411 ns/op          0 B/op          0 allocs/op
```

note: 
    manual is just the baseline where each name reference is manually assigned.

compared to "manual" of course its magnitudes slower, nonetheless I saw room for improvement: the only issue is that I used unsafe which is not recommended but saw it as acceptable, as its used to calculate the pointer of each field from their offset, and also cast *Field into a custom struct (named fieldHeader) pointer that matched the memory layout just to directly assign the string pointer to the name field.

considerations from unsafe use:

- the field pointer calculation is standard in go.
- I have to be careful with fieldHeader, so it matches Field layout correctly.

optimizations results:

```bash
goos: linux
goarch: amd64
pkg: namedtest
cpu: 11th Gen Intel(R) Core(TM) i5-1135G7 @ 2.40GHz
BenchmarkUnsafe-8      69204219                15.69 ns/op            0 B/op          0 allocs/op
BenchmarkManual-8       1000000000               0.2431 ns/op          0 B/op          0 allocs/op
```

I think this is an acceptable trade off, for my case this is good enough to handle small structs.

**[old-benchmarks](/old_bench_test.go.txt)**

### Usage:

1) implement the Field struct as field value in your structs:
```go
type ExampleStruct struct {
	A Field[int]     `json:"a"`		// field name: "a"
	J Field[int]     				// field name: "J", uses raw field name if no tag name is present
	L Field[int]     `json:"-"` 	// field name: none, if the name is "-", the name is skipped
	m Field[any]     				// field name: none, field is unexported (starts with lower case) so its skipped
}
```
2) call Link on the struct pointer (once)
```go
// x := &ExampleStruct{Field[int]{Value: 10}}
named.Link(&s, "json")
```
3) retrieve the field name with the Name method
```go
fmt.Println(x.A.Name())
Output: a
```
[example](/linker_test.go)

Field is compatible with:

- with json encoding, you can also write your custom encoder embedding the Field struct. see https://pkg.go.dev/github.com/alvarolm/named#Field
- the omitzero option from the json tag options (https://pkg.go.dev/encoding/json) as it implements the IsZero() bool method.

## post processing solution:
    
Generating go code.
Still needed to have a solution without the use of unsafe package and as close to native performance for comparison.

To resolve A:

Adding methods to the original struct sounded fine at first but I could not find the right naming and somehow seemed too invasive to me.
I wanted to keep things separated but intuitive. Decided to create a struct with a method for each field that returns the field name, then export an instance of this struct with the same original name but with the suffix "Named". I admit that there is a gap to err when typing this exported var but its minimal.

To resolve B:

There is not much to think, the generated code would come at the expense of a few ns at most.

### Usage:

1) install
```bash
go install github.com/alvarolm/named/cmd/generate-named@latest
```

2) add the named directives
```go
// GENERATE-NAMED=StructName:Person,TagKey:json
```
these can go anywere as long as they are in the same directory of the structs

3) add the go generate directive
```go
//go:generate generate-named .
```
just once.

4) call the generated field method
```go
// v:=&Person{}
fmt.Println(PersonNamed.Email())
Output: email
```
[example](/generate_example.go)

<details>
<summary>generate-named options</summary>

	Usage: generate-named [flags] [path...]
	
	Generates type-safe field name accessors for Go structs.
	
	Flags:
	  -clean
	    	remove all generated *_named_generated.go files
	  -v	verbose mode: show detailed processing information
	  -verbose
	    	verbose mode: show detailed processing information
	
	Arguments:
	  path    File or directory to process (default: current directory)
	
	Examples:
	  generate-named                    # Process current directory
	  generate-named -v                 # Process with verbose output
	  generate-named -clean             # Remove all generated files
	  generate-named ./pkg              # Process specific directory
	  generate-named file.go            # Process specific file
	
	For each struct with a GENERATE-NAMED directive, creates a *_named_generated.go file
	with methods to access field names based on struct tags.
</details>

## which should you use ?

it depends on your needs.

Use the runtime solution if:
- understand the limitations of the Field struct.
- want to avoid the hassle of an extra build step.
- dont mind the extra overhead (grows with number of fields)

Use the post processing solution if:
- performance is critical
- want to avoid the use of the unsafe package
- dont mind the extra build step
- must not alter existing structs
