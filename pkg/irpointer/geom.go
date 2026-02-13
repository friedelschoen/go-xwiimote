package irpointer

// FVec2 represents a 2D floating point vector to X and Y.
type FVec2 struct {
	X, Y float64
}

// FRect represents a 2D floating point rectangle, streched over an Min and Max point.
type FRect struct {
	Min, Max FVec2
}

func (r FRect) Contains(p FVec2) bool {
	return p.X >= r.Min.X && p.X < r.Max.X && p.Y >= r.Min.Y && p.Y < r.Max.Y
}

func (r FRect) Width() float64 {
	return r.Max.X - r.Min.X
}

func (r FRect) Height() float64 {
	return r.Max.Y - r.Min.Y
}

func (r FRect) Empty() bool {
	return r.Max.X <= r.Min.X || r.Max.Y <= r.Min.Y
}

func (r FRect) Normal(p FVec2) FVec2 {
	if r.Empty() {
		return FVec2{}
	}
	if p.X < r.Min.X {
		p.X = r.Min.X
	} else if p.X >= r.Max.X-1 {
		p.X = r.Max.X
	}
	if p.Y < r.Min.Y {
		p.Y = r.Min.Y
	} else if p.Y >= r.Max.Y {
		p.Y = r.Max.Y - 1
	}

	w := r.Width()
	h := r.Height()

	x := (p.X - r.Min.X) / w
	y := (p.Y - r.Min.Y) / h

	return FVec2{X: x*2 - 1, Y: y*2 - 1}
}
