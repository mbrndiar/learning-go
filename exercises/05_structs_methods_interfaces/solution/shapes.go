// Package shapes is the reference implementation for the structs, methods,
// and interfaces exercise. See ../shapes.go for the task descriptions.
package shapes

import (
	"fmt"
	"math"
)

// Shape is satisfied by any figure that can report its area and perimeter.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rectangle is a four-sided shape defined by width and height.
type Rectangle struct {
	Width  float64
	Height float64
}

// Circle is a round shape defined by its radius.
type Circle struct {
	Radius float64
}

// Area returns the rectangle's area (Width * Height).
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter returns the rectangle's perimeter (2 * (Width + Height)).
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// String renders the rectangle as "Rectangle(width=%.2f, height=%.2f)" so
// that Rectangle satisfies fmt.Stringer.
func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(width=%.2f, height=%.2f)", r.Width, r.Height)
}

// Scale multiplies Width and Height by factor in place. A factor that is
// less than or equal to zero leaves the rectangle unchanged.
func (r *Rectangle) Scale(factor float64) {
	if factor <= 0 {
		return
	}
	r.Width *= factor
	r.Height *= factor
}

// Area returns the circle's area (pi * Radius^2).
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Perimeter returns the circle's circumference (2 * pi * Radius).
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// String renders the circle as "Circle(radius=%.2f)" so that Circle
// satisfies fmt.Stringer.
func (c Circle) String() string {
	return fmt.Sprintf("Circle(radius=%.2f)", c.Radius)
}

// Scale multiplies Radius by factor in place. A factor that is less than or
// equal to zero leaves the circle unchanged.
func (c *Circle) Scale(factor float64) {
	if factor <= 0 {
		return
	}
	c.Radius *= factor
}

// Size classifies a shape by its area using an iota-based enumeration.
type Size int

// Declared Size values, in increasing order of area.
const (
	Small Size = iota
	Medium
	Large
)

// String renders Size as "Small", "Medium", or "Large". Any value outside
// the declared constants renders as "Unknown", so Size satisfies
// fmt.Stringer.
func (s Size) String() string {
	switch s {
	case Small:
		return "Small"
	case Medium:
		return "Medium"
	case Large:
		return "Large"
	default:
		return "Unknown"
	}
}

// Classify returns Small for shapes with an area below 10, Medium for an
// area in [10, 100), and Large for an area of 100 or more.
func Classify(s Shape) Size {
	area := s.Area()
	switch {
	case area < 10:
		return Small
	case area < 100:
		return Medium
	default:
		return Large
	}
}

// TotalArea returns the sum of Area() across every shape in shapes. It
// returns 0 for an empty or nil slice.
func TotalArea(shapes []Shape) float64 {
	var total float64
	for _, s := range shapes {
		total += s.Area()
	}
	return total
}
