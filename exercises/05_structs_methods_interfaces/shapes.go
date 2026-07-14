// Package shapes models simple 2D geometric figures to practice structs,
// value and pointer receivers, interfaces, the fmt.Stringer interface, and
// iota-based enumerations.
//
// Implement every function and method below. Replace each panic("not
// implemented") with working code; do not change any signature.
package shapes

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
//
// TODO(task 1): implement Area for Rectangle.
func (r Rectangle) Area() float64 {
	panic("not implemented")
}

// Perimeter returns the rectangle's perimeter (2 * (Width + Height)).
//
// TODO(task 1): implement Perimeter for Rectangle.
func (r Rectangle) Perimeter() float64 {
	panic("not implemented")
}

// String renders the rectangle as "Rectangle(width=%.2f, height=%.2f)" so
// that Rectangle satisfies fmt.Stringer.
//
// TODO(task 2): implement String for Rectangle.
func (r Rectangle) String() string {
	panic("not implemented")
}

// Scale multiplies Width and Height by factor in place. A factor that is
// less than or equal to zero must leave the rectangle unchanged.
//
// TODO(task 3): implement Scale for Rectangle using a pointer receiver.
func (r *Rectangle) Scale(factor float64) {
	panic("not implemented")
}

// Area returns the circle's area (pi * Radius^2). Use math.Pi.
//
// TODO(task 4): implement Area for Circle.
func (c Circle) Area() float64 {
	panic("not implemented")
}

// Perimeter returns the circle's circumference (2 * pi * Radius).
//
// TODO(task 4): implement Perimeter for Circle.
func (c Circle) Perimeter() float64 {
	panic("not implemented")
}

// String renders the circle as "Circle(radius=%.2f)" so that Circle
// satisfies fmt.Stringer.
//
// TODO(task 5): implement String for Circle.
func (c Circle) String() string {
	panic("not implemented")
}

// Scale multiplies Radius by factor in place. A factor that is less than or
// equal to zero must leave the circle unchanged.
//
// TODO(task 5): implement Scale for Circle using a pointer receiver.
func (c *Circle) Scale(factor float64) {
	panic("not implemented")
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
// the declared constants must render as "Unknown", so Size satisfies
// fmt.Stringer.
//
// TODO(task 6): implement String for Size.
func (s Size) String() string {
	panic("not implemented")
}

// Classify returns Small for shapes with an area below 10, Medium for an
// area in [10, 100), and Large for an area of 100 or more.
//
// TODO(task 7): implement Classify.
func Classify(s Shape) Size {
	panic("not implemented")
}

// TotalArea returns the sum of Area() across every shape in shapes. It
// returns 0 for an empty or nil slice.
//
// TODO(task 7): implement TotalArea.
func TotalArea(shapes []Shape) float64 {
	panic("not implemented")
}
