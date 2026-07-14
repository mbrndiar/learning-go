// Package main demonstrates package organization: an exported, importable
// API (catalog) plus an internal implementation detail (internal/pricing)
// that Go prevents other modules and unrelated packages from depending on.
package main

import (
	"fmt"

	"github.com/mbrndiar/learning-go/lessons/07_packages_and_generics/01_package_organization/catalog"
	"github.com/mbrndiar/learning-go/lessons/07_packages_and_generics/01_package_organization/internal/pricing"
)

func main() {
	// Import paths are built from the module path declared in go.mod
	// ("github.com/mbrndiar/learning-go") plus the directory path to the
	// package. There is nothing magic about the URL-shaped prefix here: it
	// only needs to be unique enough to avoid clashing with other modules,
	// and by convention matches where the module's source is hosted.
	product := catalog.New("Desk Lamp", "DL-010")
	fmt.Println("product:", product)

	basePrice := 40.0
	discounted := pricing.ApplyDiscount(basePrice, 25)
	fmt.Printf("price: %.2f -> %.2f after discount\n", basePrice, discounted)

	// catalog is a normal package: any other module that imported this
	// repository could use catalog.Product and catalog.New exactly as this
	// file does.
	//
	// pricing cannot be imported that way. Its import path contains
	// "internal", so only code rooted at this lesson's own directory
	// ("01_package_organization" and anything below it) is permitted to
	// import it - not sibling lessons like "02_generic_helpers", and not an
	// external module. Try adding an internal-path import from outside this
	// tree and the build fails with "use of internal package ... not
	// allowed", not just a lint warning.
	fmt.Println("exported vs internal names: catalog.Product, catalog.New, pricing.ApplyDiscount")
}
