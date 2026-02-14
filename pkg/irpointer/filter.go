package irpointer

import "math"

type Filter interface {
	Apply(Frame) Frame
}

type FilterChain []Filter

func (chain FilterChain) Apply(frame Frame) Frame {
	for _, f := range chain {
		frame = f.Apply(frame)
	}
	return frame
}

type ErrorFilter struct {
	// ErrorMaxCount is max number of errors before cooked data drops out
	ErrorMaxCount int

	count    int
	position FVec2
	alive    bool
}

func (f *ErrorFilter) Apply(frame Frame) Frame {
	if frame.Valid {
		f.count = 0
		f.position = frame.Position
		f.alive = true
	} else if f.alive && f.count < f.ErrorMaxCount {
		f.count++
		frame.Position = f.position
		frame.Valid = true
	}
	return frame
}

type GlitchFilter struct {
	// GlitchMaxCount is max number of glitches before cooked data updates
	GlitchMaxCount int

	// GlitchDistance is squared delta over which we consider something a glitch
	GlitchDistance float64

	count    int
	position FVec2
	alive    bool
}

func (f *GlitchFilter) Apply(frame Frame) Frame {
	if !frame.Valid {
		// Belangrijk: geen "glitch-straf" laten doorlekken na dropouts.
		f.count = 0
		return frame
	}

	if !f.alive {
		f.alive = true
		f.position = frame.Position
		f.count = 0
		return frame
	}

	d := square(frame.Position.X-f.position.X) + square(frame.Position.Y-f.position.Y)
	if d > f.GlitchDistance && f.count < f.GlitchMaxCount {
		f.count++
		frame.Position = f.position
		return frame
	}

	f.position = frame.Position
	f.count = 0
	return frame
}

type MarcanSmoothingFilter struct {
	SmootherRadius float64
	SmootherSpeed  float64

	// SmootherDeadzone is distance between old and new value where nothing should change
	SmootherDeadzone float64

	position FVec2
	alive    bool
}

func (f *MarcanSmoothingFilter) Apply(frame Frame) Frame {
	if !frame.Valid {
		return frame
	}

	raw := frame.Position

	if !f.alive {
		f.alive = true
		f.position = raw
		return frame
	}

	pos := f.position
	dx := raw.X - pos.X
	dy := raw.Y - pos.Y

	d := math.Sqrt(square(dx) + square(dy))
	if d <= f.SmootherDeadzone {
		frame.Position = pos
		return frame
	}

	if d < f.SmootherRadius {
		pos.X += dx * f.SmootherSpeed
		pos.Y += dy * f.SmootherSpeed
	} else {
		theta := math.Atan2(dy, dx)
		pos.X = raw.X - math.Cos(theta)*f.SmootherRadius
		pos.Y = raw.Y - math.Sin(theta)*f.SmootherRadius
	}

	f.position = pos
	frame.Position = pos
	return frame
}

type TranslateFilter struct {
	// input viewpoint
	Source FRect
	// output viewpoint
	Destination FRect
	Clamp       bool
}

func (f *TranslateFilter) Apply(frame Frame) Frame {
	if !frame.Valid {
		return frame
	}

	frame.Valid = f.Clamp || f.Source.Contains(frame.Position)
	frame.Position = f.Source.Translate(frame.Position, f.Destination, f.Clamp)
	return frame
}
