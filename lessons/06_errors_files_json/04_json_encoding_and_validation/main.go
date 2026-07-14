// Package main covers JSON struct tags, encoding, decoding, and validating
// decoded data at the boundary of a program.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Product's struct tags control how encoding/json maps Go fields to JSON
// keys. `json:"sku"` renames the field; `,omitempty` drops the field from
// the output entirely when it holds its zero value.
type Product struct {
	Name  string   `json:"name"`
	SKU   string   `json:"sku"`
	Price float64  `json:"price"`
	Tags  []string `json:"tags,omitempty"`
}

// Validate enforces the boundary rules for a decoded Product. Decoding only
// checks that the JSON *shape* matches the struct; it says nothing about
// whether the values make sense. Always validate data that crosses a
// program boundary (a file, a request body, a config) before trusting it.
func (p Product) Validate() error {
	switch {
	case p.Name == "":
		return fmt.Errorf("product: %w: name is required", errInvalidProduct)
	case p.SKU == "":
		return fmt.Errorf("product: %w: sku is required", errInvalidProduct)
	case p.Price <= 0:
		return fmt.Errorf("product: %w: price must be positive, got %.2f", errInvalidProduct, p.Price)
	default:
		return nil
	}
}

var errInvalidProduct = fmt.Errorf("invalid product")

func main() {
	demoEncode()
	demoDecode()
	demoValidationBoundary()
}

// demoEncode shows json.Marshal and json.MarshalIndent, and the effect of
// omitempty on a zero-value slice field.
func demoEncode() {
	fmt.Println("--- encoding ---")

	withTags := Product{Name: "Keyboard", SKU: "KB-100", Price: 49.99, Tags: []string{"input", "usb"}}
	compact, err := json.Marshal(withTags)
	if err != nil {
		fmt.Println("marshal:", err)
		return
	}
	fmt.Println("compact:", string(compact))

	indented, err := json.MarshalIndent(withTags, "", "  ")
	if err != nil {
		fmt.Println("marshal indent:", err)
		return
	}
	fmt.Println("indented:")
	fmt.Println(string(indented))

	withoutTags := Product{Name: "Monitor", SKU: "MN-200", Price: 199.0}
	compactNoTags, err := json.Marshal(withoutTags)
	if err != nil {
		fmt.Println("marshal:", err)
		return
	}
	// Tags is nil (its zero value), so omitempty removes "tags" entirely
	// instead of encoding it as null or [].
	fmt.Println("omitempty drops tags:", string(compactNoTags))
}

// demoDecode shows json.Unmarshal for a simple case, and json.Decoder with
// DisallowUnknownFields to reject JSON that does not match the struct's
// known fields - a common way to catch typos and unexpected input early.
func demoDecode() {
	fmt.Println("--- decoding ---")

	validJSON := []byte(`{"name":"Headset","sku":"HS-300","price":89.5,"tags":["audio"]}`)
	var product Product
	if err := json.Unmarshal(validJSON, &product); err != nil {
		fmt.Println("unmarshal:", err)
		return
	}
	fmt.Printf("decoded: %+v\n", product)

	unknownFieldJSON := []byte(`{"name":"Headset","sku":"HS-300","price":89.5,"discount":10}`)

	var lenient Product
	if err := json.Unmarshal(unknownFieldJSON, &lenient); err != nil {
		fmt.Println("lenient unmarshal:", err)
	} else {
		// json.Unmarshal silently ignores fields it does not recognize.
		fmt.Printf("lenient decode ignored 'discount': %+v\n", lenient)
	}

	var strict Product
	decoder := json.NewDecoder(bytes.NewReader(unknownFieldJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&strict); err != nil {
		// DisallowUnknownFields makes the same input fail loudly instead of
		// silently dropping data the caller might have expected to matter.
		fmt.Println("strict decode rejected 'discount':", err)
	}
}

// demoValidationBoundary decodes two inputs that are both structurally
// valid JSON matching Product's shape, but only one passes business
// validation.
func demoValidationBoundary() {
	fmt.Println("--- boundary validation ---")

	inputs := []string{
		`{"name":"Mouse","sku":"MS-400","price":25.0}`,
		`{"name":"","sku":"MS-401","price":0}`,
	}

	for _, raw := range inputs {
		var product Product
		if err := json.Unmarshal([]byte(raw), &product); err != nil {
			fmt.Println("unmarshal:", err)
			continue
		}
		if err := product.Validate(); err != nil {
			fmt.Println("rejected:", err)
			continue
		}
		fmt.Printf("accepted: %+v\n", product)
	}
}
