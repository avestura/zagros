package mathx

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func noFamily(f *tools.Func) { f.Family = "" }

func positive(a tools.Args, names ...string) error {
	for _, n := range names {
		v := a.Float(n)
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return fmt.Errorf("%s must be a non-negative number", n)
		}
	}
	return nil
}

func heron(a, b, c float64) (float64, error) {
	if a <= 0 || b <= 0 || c <= 0 || a+b <= c || a+c <= b || b+c <= a {
		return 0, errors.New("these sides cannot form a triangle")
	}
	s := (a + b + c) / 2
	return math.Sqrt(s * (s - a) * (s - b) * (s - c)), nil
}

func shapeInfo(shape string, a, b, c float64) (map[string]string, error) {
	for _, v := range []float64{a, b, c} {
		if v < 0 || math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, errors.New("dimensions must be non-negative numbers")
		}
	}
	need := func(n int) error {
		vs := []float64{a, b, c}
		for i := 0; i < n; i++ {
			if vs[i] == 0 {
				return fmt.Errorf("%s needs %d positive dimension(s)", shape, n)
			}
		}
		return nil
	}
	pi := math.Pi
	switch strings.ToLower(strings.TrimSpace(shape)) {
	case "circle":
		if err := need(1); err != nil {
			return nil, err
		}
		return dict("area", pi*a*a, "perimeter", 2*pi*a, "circumference", 2*pi*a, "diameter", 2*a), nil
	case "rect", "rectangle":
		if err := need(2); err != nil {
			return nil, err
		}
		return dict("area", a*b, "perimeter", 2*(a+b), "diagonal", math.Hypot(a, b)), nil
	case "square":
		if err := need(1); err != nil {
			return nil, err
		}
		return dict("area", a*a, "perimeter", 4*a, "diagonal", a*math.Sqrt2), nil
	case "triangle":
		if err := need(3); err != nil {
			return nil, err
		}
		ar, err := heron(a, b, c)
		if err != nil {
			return nil, err
		}
		return dict("area", ar, "perimeter", a+b+c, "inradius", ar/((a+b+c)/2), "circumradius", a*b*c/(4*ar)), nil
	case "ellipse":
		if err := need(2); err != nil {
			return nil, err
		}
		h := math.Pow(a-b, 2) / math.Pow(a+b, 2)
		return dict("area", pi*a*b, "perimeter", pi*(a+b)*(1+3*h/(10+math.Sqrt(4-3*h)))), nil
	case "sphere":
		if err := need(1); err != nil {
			return nil, err
		}
		return dict("volume", 4.0/3*pi*a*a*a, "surface_area", 4*pi*a*a, "area", 4*pi*a*a), nil
	case "cube":
		if err := need(1); err != nil {
			return nil, err
		}
		return dict("volume", a*a*a, "surface_area", 6*a*a, "area", 6*a*a, "diagonal", a*math.Sqrt(3)), nil
	case "cylinder":
		if err := need(2); err != nil {
			return nil, err
		}
		return dict("volume", pi*a*a*b, "surface_area", 2*pi*a*(a+b), "area", 2*pi*a*(a+b), "lateral_area", 2*pi*a*b), nil
	case "cone":
		if err := need(2); err != nil {
			return nil, err
		}
		sl := math.Hypot(a, b)
		return dict("volume", pi*a*a*b/3, "surface_area", pi*a*(a+sl), "area", pi*a*(a+sl), "lateral_area", pi*a*sl, "slant_height", sl), nil
	}
	return nil, fmt.Errorf("unknown shape %q (circle, rectangle, square, triangle, ellipse, sphere, cube, cylinder, cone)", shape)
}

func init() {
	reg("geometry", "Shape geometry", "Area, perimeter, volume and other measures of a shape as a dict. circle(a=radius), rectangle(a=width, b=height), square(a), triangle(a, b, c = three sides), ellipse(a, b = semi-axes), sphere(a=radius), cube(a), cylinder(a=radius, b=height), cone(a=radius, b=height). Unused dimensions stay 0; non-positive needed ones are an error.",
		"dict[str, str]", `math.geometry("cylinder", 2, 5)`, 92, "geometry shape area volume perimeter circumference surface circle rectangle triangle sphere cylinder cone cube",
		[]P{rs("shape", "Shape name."), of("a", "First dimension.", 0), of("b", "Second dimension.", 0), of("c", "Third dimension.", 0)},
		func(a tools.Args) (any, error) {
			return shapeInfo(a.Str("shape"), a.Float("a"), a.Float("b"), a.Float("c"))
		}, noFamily)

	one := func(name, title, doc, ex string, order int, params []P, f func(a tools.Args) (float64, error)) {
		var names []string
		for _, p := range params {
			names = append(names, p.Name)
		}
		reg(name, title, doc, "float", ex, order, "geometry area volume perimeter surface shape size "+strings.ReplaceAll(name, "_", " "), params,
			func(a tools.Args) (any, error) {
				if err := positive(a, names...); err != nil {
					return nil, err
				}
				r, err := f(a)
				if err != nil {
					return nil, err
				}
				return finite(name, r)
			})
	}
	r := []P{rf("r", "Radius.")}
	rh := []P{rf("r", "Radius."), rf("h", "Height.")}
	one("circle_area", "Circle area", "Area of a circle: pi*r^2.", `math.circle_area(3)`, 93, r, func(a tools.Args) (float64, error) { return math.Pi * a.Float("r") * a.Float("r"), nil })
	one("circle_circumference", "Circle circumference", "Circumference of a circle: 2*pi*r.", `math.circle_circumference(3)`, 94, r, func(a tools.Args) (float64, error) { return 2 * math.Pi * a.Float("r"), nil })
	one("rect_area", "Rectangle area", "Area of a rectangle: w*h.", `math.rect_area(3, 4)`, 95, []P{rf("w", "Width."), rf("h", "Height.")}, func(a tools.Args) (float64, error) { return a.Float("w") * a.Float("h"), nil })
	one("rect_perimeter", "Rectangle perimeter", "Perimeter of a rectangle: 2*(w+h).", `math.rect_perimeter(3, 4)`, 96, []P{rf("w", "Width."), rf("h", "Height.")}, func(a tools.Args) (float64, error) { return 2 * (a.Float("w") + a.Float("h")), nil })
	one("triangle_area", "Triangle area", "Area of a triangle from its three side lengths (Heron's formula; 3, 4, 5 gives 6). Sides that violate the triangle inequality are an error.", `math.triangle_area(3, 4, 5)`, 97,
		[]P{rf("a", "Side a."), rf("b", "Side b."), rf("c", "Side c.")}, func(x tools.Args) (float64, error) { return heron(x.Float("a"), x.Float("b"), x.Float("c")) })
	one("triangle_perimeter", "Triangle perimeter", "Perimeter of a triangle: a+b+c (sides must satisfy the triangle inequality).", `math.triangle_perimeter(3, 4, 5)`, 98,
		[]P{rf("a", "Side a."), rf("b", "Side b."), rf("c", "Side c.")}, func(x tools.Args) (float64, error) {
			if _, err := heron(x.Float("a"), x.Float("b"), x.Float("c")); err != nil {
				return 0, err
			}
			return x.Float("a") + x.Float("b") + x.Float("c"), nil
		})
	one("sphere_volume", "Sphere volume", "Volume of a sphere: 4/3*pi*r^3.", `math.sphere_volume(2)`, 99, r, func(a tools.Args) (float64, error) { return 4.0 / 3 * math.Pi * math.Pow(a.Float("r"), 3), nil })
	one("sphere_area", "Sphere surface area", "Surface area of a sphere: 4*pi*r^2.", `math.sphere_area(2)`, 100, r, func(a tools.Args) (float64, error) { return 4 * math.Pi * a.Float("r") * a.Float("r"), nil })
	one("cylinder_volume", "Cylinder volume", "Volume of a cylinder: pi*r^2*h.", `math.cylinder_volume(2, 5)`, 101, rh, func(a tools.Args) (float64, error) { return math.Pi * a.Float("r") * a.Float("r") * a.Float("h"), nil })
	one("cylinder_area", "Cylinder surface area", "Total surface area of a cylinder: 2*pi*r*(r+h).", `math.cylinder_area(2, 5)`, 102, rh, func(a tools.Args) (float64, error) {
		return 2 * math.Pi * a.Float("r") * (a.Float("r") + a.Float("h")), nil
	})
	one("cone_volume", "Cone volume", "Volume of a cone: pi*r^2*h/3.", `math.cone_volume(2, 5)`, 103, rh, func(a tools.Args) (float64, error) {
		return math.Pi * a.Float("r") * a.Float("r") * a.Float("h") / 3, nil
	})
	one("cone_area", "Cone surface area", "Total surface area of a cone: pi*r*(r+slant) with slant = sqrt(r^2+h^2).", `math.cone_area(2, 5)`, 104, rh, func(a tools.Args) (float64, error) {
		return math.Pi * a.Float("r") * (a.Float("r") + math.Hypot(a.Float("r"), a.Float("h"))), nil
	})

	reg("pythagoras", "Pythagorean theorem", "Solves a right triangle with legs a, b and hypotenuse c: set exactly one of them to 0 (the unknown) and the other two to their values. pythagoras(3, 4, 0) = 5; pythagoras(0, 12, 13) = 5. The hypotenuse must be longer than a leg.",
		"float", `math.pythagoras(3, 4, 0)`, 105, "pythagoras pythagorean theorem right triangle hypotenuse leg a2+b2=c2",
		[]P{of("a", "Leg a (0 if unknown).", 0), of("b", "Leg b (0 if unknown).", 0), of("c", "Hypotenuse c (0 if unknown).", 0)},
		func(x tools.Args) (any, error) {
			a, b, c := x.Float("a"), x.Float("b"), x.Float("c")
			if a < 0 || b < 0 || c < 0 {
				return nil, errors.New("sides must not be negative")
			}
			zeros := 0
			for _, v := range []float64{a, b, c} {
				if v == 0 {
					zeros++
				}
			}
			if zeros != 1 {
				return nil, errors.New("set exactly one side to 0 (the unknown)")
			}
			switch {
			case c == 0:
				return finite("pythagoras", math.Hypot(a, b))
			case a == 0:
				if c <= b {
					return nil, errors.New("the hypotenuse must be longer than a leg")
				}
				return math.Sqrt((c - b) * (c + b)), nil
			}
			if c <= a {
				return nil, errors.New("the hypotenuse must be longer than a leg")
			}
			return math.Sqrt((c - a) * (c + a)), nil
		})
}
