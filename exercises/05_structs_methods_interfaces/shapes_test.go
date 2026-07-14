package shapes

import (
	"fmt"
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	const epsilon = 1e-9
	return math.Abs(a-b) < epsilon
}

func TestRectangleAreaAndPerimeter(t *testing.T) {
	tests := []struct {
		name          string
		width, height float64
		wantArea      float64
		wantPerimeter float64
	}{
		{"square", 4, 4, 16, 16},
		{"wide", 10, 2, 20, 24},
		{"zero height", 5, 0, 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rectangle{Width: tt.width, Height: tt.height}
			if got := r.Area(); !almostEqual(got, tt.wantArea) {
				t.Errorf("Area() = %v, want %v", got, tt.wantArea)
			}
			if got := r.Perimeter(); !almostEqual(got, tt.wantPerimeter) {
				t.Errorf("Perimeter() = %v, want %v", got, tt.wantPerimeter)
			}
		})
	}
}

func TestCircleAreaAndPerimeter(t *testing.T) {
	tests := []struct {
		name   string
		radius float64
	}{
		{"unit circle", 1},
		{"radius two", 2},
		{"radius five", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Circle{Radius: tt.radius}
			wantArea := math.Pi * tt.radius * tt.radius
			wantPerimeter := 2 * math.Pi * tt.radius
			if got := c.Area(); !almostEqual(got, wantArea) {
				t.Errorf("Area() = %v, want %v", got, wantArea)
			}
			if got := c.Perimeter(); !almostEqual(got, wantPerimeter) {
				t.Errorf("Perimeter() = %v, want %v", got, wantPerimeter)
			}
		})
	}
}

func TestStringers(t *testing.T) {
	tests := []struct {
		name  string
		shape fmt.Stringer
		want  string
	}{
		{"rectangle", Rectangle{Width: 3, Height: 4.5}, "Rectangle(width=3.00, height=4.50)"},
		{"circle", Circle{Radius: 2}, "Circle(radius=2.00)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.shape.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRectangleScale(t *testing.T) {
	tests := []struct {
		name       string
		factor     float64
		wantWidth  float64
		wantHeight float64
	}{
		{"double", 2, 6, 8},
		{"half", 0.5, 1.5, 2},
		{"zero factor is a no-op", 0, 3, 4},
		{"negative factor is a no-op", -1, 3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Rectangle{Width: 3, Height: 4}
			r.Scale(tt.factor)
			if !almostEqual(r.Width, tt.wantWidth) || !almostEqual(r.Height, tt.wantHeight) {
				t.Errorf("after Scale(%v) = %+v, want width=%v height=%v", tt.factor, r, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestCircleScale(t *testing.T) {
	tests := []struct {
		name       string
		factor     float64
		wantRadius float64
	}{
		{"double", 2, 6},
		{"half", 0.5, 1.5},
		{"zero factor is a no-op", 0, 3},
		{"negative factor is a no-op", -2, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Circle{Radius: 3}
			c.Scale(tt.factor)
			if !almostEqual(c.Radius, tt.wantRadius) {
				t.Errorf("after Scale(%v) radius = %v, want %v", tt.factor, c.Radius, tt.wantRadius)
			}
		})
	}
}

func TestSizeString(t *testing.T) {
	tests := []struct {
		name string
		size Size
		want string
	}{
		{"small", Small, "Small"},
		{"medium", Medium, "Medium"},
		{"large", Large, "Large"},
		{"out of range", Size(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.size.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		s    Shape
		want Size
	}{
		{"small rectangle", Rectangle{Width: 2, Height: 3}, Small},
		{"medium rectangle", Rectangle{Width: 10, Height: 5}, Medium},
		{"large rectangle", Rectangle{Width: 20, Height: 10}, Large},
		{"boundary at ten is medium", Rectangle{Width: 2, Height: 5}, Medium},
		{"boundary at one hundred is large", Rectangle{Width: 10, Height: 10}, Large},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.s); got != tt.want {
				t.Errorf("Classify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTotalArea(t *testing.T) {
	tests := []struct {
		name   string
		shapes []Shape
		want   float64
	}{
		{"empty", nil, 0},
		{"single rectangle", []Shape{Rectangle{Width: 2, Height: 3}}, 6},
		{
			"mixed shapes",
			[]Shape{Rectangle{Width: 2, Height: 3}, Circle{Radius: 1}},
			6 + math.Pi,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalArea(tt.shapes); !almostEqual(got, tt.want) {
				t.Errorf("TotalArea() = %v, want %v", got, tt.want)
			}
		})
	}
}
