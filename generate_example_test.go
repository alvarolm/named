package named

import "testing"

func TestPersonNamed(t *testing.T) {
	n := PersonNamed

	tests := []struct {
		method   func() string
		expected string
	}{
		{n.Name, "name"},
		{n.Age, "age"},
		{n.Email, "email"},
	}

	for _, tt := range tests {
		if got := tt.method(); got != tt.expected {
			t.Errorf("Expected %q, got %q", tt.expected, got)
		}
	}
}

func TestUserNamed(t *testing.T) {
	n := UserNamed

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"ID", n.ID, "user_id"},
		{"Username", n.Username, "username"},
		{"Active", n.Active, "is_active"},
	}

	for _, tt := range tests {
		if got := tt.method(); got != tt.expected {
			t.Errorf("%s: Expected %q, got %q", tt.name, tt.expected, got)
		}
	}
}

func TestProductNamed(t *testing.T) {
	n := ProductNamed

	tests := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"SKU", n.SKU, "sku"},
		{"Name", n.Name, "product_name"},
		{"Price", n.Price, "price"},
		{"Description", n.Description, "Description"},
	}

	for _, tt := range tests {
		if got := tt.method(); got != tt.expected {
			t.Errorf("%s: Expected %q, got %q", tt.name, tt.expected, got)
		}
	}
}

func TestGeneratedNamedWithActualStruct(t *testing.T) {
	// Test that the Named struct provides correct field name access
	n := PersonNamed

	// Verify each method returns the expected tag value
	if got := n.Name(); got != "name" {
		t.Errorf("Name(): expected %q, got %q", "name", got)
	}

	if got := n.Age(); got != "age" {
		t.Errorf("Age1(): expected %q, got %q", "age1", got)
	}

	if got := n.Email(); got != "email" {
		t.Errorf("Email(): expected %q, got %q", "email", got)
	}
}

func BenchmarkGeneratedNamed(b *testing.B) {
	n := PersonNamed
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = n.Name()
	}
}
