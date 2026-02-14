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

func (r FRect) Translate(p FVec2, dst FRect, clamp bool) FVec2 {
	if r.Empty() || dst.Empty() {
		return FVec2{}
	}

	if clamp {
		if p.X < r.Min.X {
			p.X = r.Min.X
		} else if p.X > r.Max.X {
			p.X = r.Max.X
		}
		if p.Y < r.Min.Y {
			p.Y = r.Min.Y
		} else if p.Y > r.Max.Y {
			p.Y = r.Max.Y
		}
	}
	if r == dst {
		return p
	}

	// Normaliseer naar 0..1 binnen source
	w := r.Width()
	h := r.Height()

	nx := (p.X - r.Min.X) / w
	ny := (p.Y - r.Min.Y) / h

	// Projecteer naar destination
	dx := dst.Min.X + nx*dst.Width()
	dy := dst.Min.Y + ny*dst.Height()

	return FVec2{X: dx, Y: dy}
}
