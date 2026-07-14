// Package catalog is the public-style API of this lesson: any code that can
// import this lesson's module path can import catalog, the same way an
// external package would depend on a library package.
package catalog

import "fmt"

// Product is exported (its name starts with an uppercase letter), so code
// outside this package can refer to catalog.Product and to its exported
// fields. An identifier starting with a lowercase letter would only be
// visible inside this package.
type Product struct {
	Name string
	SKU  string
}

// New is an exported constructor. Exporting a constructor alongside a
// struct's exported fields is optional in Go - callers may also build a
// Product directly with a struct literal - but a constructor is useful once
// creation needs validation, defaults, or is likely to gain steps later.
func New(name, sku string) Product {
	return Product{Name: name, SKU: sku}
}

// String implements fmt.Stringer so a Product prints legibly.
func (p Product) String() string {
	return fmt.Sprintf("%s (%s)", p.Name, p.SKU)
}
