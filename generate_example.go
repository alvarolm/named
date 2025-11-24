package named

// Struct definitions

// GENERATE-NAMED=StructName:TestStruct,TagKey:json
// GENERATE-NAMED=StructName:Person,TagKey:json
// GENERATE-NAMED=StructName:User,TagKey:db
// GENERATE-NAMED=StructName:Product,TagKey:json
// These directives can be in any file.

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

type Person struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type User struct {
	ID       int    `db:"user_id"`
	Username string `db:"username"`
	Password string `db:"-"` // should be skipped
	Active   bool   `db:"is_active"`
	internal string `db:"internal"` // should be skipped (unexported)
}

type Product struct {
	SKU         string  `json:"sku"`
	Name        string  `json:"product_name"`
	Price       float64 `json:"price"`
	Description string  // no tag, should use field name
}
