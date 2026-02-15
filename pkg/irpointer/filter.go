package irpointer

import (
	"math"
	"time"
)

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

/*
OneEuroSmoothingFilter
- minCutoff: basis cutoff (Hz). Lager = gladder maar meer lag.
- beta: hoeveel cutoff stijgt bij snelheid. Hoger = minder lag bij snelle beweging.
- dCutoff: cutoff voor de afgeleide (Hz). Meestal 1..2 Hz.

Aanrader startwaarden voor IR pointer:

	MinCutoff = 1.2
	Beta      = 0.02 .. 0.08 (afhankelijk van units en snelheid)
	DCutoff   = 1.0
*/
type OneEuroSmoothingFilter struct {
	MinCutoff float64
	Beta      float64
	DCutoff   float64

	ClampMaxDt time.Duration // optioneel: cap dt bij lange pauses (bv 100ms). 0 = uit.

	alive bool
	lastT time.Time

	x oneEuroAxis
	y oneEuroAxis
}

type oneEuroAxis struct {
	xf lowPass
	dx lowPass
}

type lowPass struct {
	initialized bool
	y           float64
}

func (lp *lowPass) apply(x, alpha float64) float64 {
	if !lp.initialized {
		lp.initialized = true
		lp.y = x
		return lp.y
	}
	lp.y = alpha*x + (1.0-alpha)*lp.y
	return lp.y
}

func alpha(dt, cutoffHz float64) float64 {
	// dt in seconds, cutoff in Hz
	// alpha = 1 / (1 + tau/dt), tau = 1/(2*pi*cutoff)
	if cutoffHz <= 0 || dt <= 0 {
		return 1.0
	}
	tau := 1.0 / (2.0 * math.Pi * cutoffHz)
	return 1.0 / (1.0 + tau/dt)
}

func (f *OneEuroSmoothingFilter) Apply(frame Frame) Frame {
	if !frame.Valid {
		return frame
	}

	now := time.Now()

	if !f.alive {
		f.alive = true
		f.lastT = now
		f.x.xf.apply(frame.Position.X, 1.0)
		f.y.xf.apply(frame.Position.Y, 1.0)
		// dx filters init op 0 snelheid
		f.x.dx.apply(0, 1.0)
		f.y.dx.apply(0, 1.0)
		return frame
	}

	dtDur := now.Sub(f.lastT)
	if dtDur <= 0 {
		return frame
	}
	if f.ClampMaxDt > 0 && dtDur > f.ClampMaxDt {
		dtDur = f.ClampMaxDt
	}
	f.lastT = now

	dt := dtDur.Seconds()

	// Afgeleide (ruw) en low-pass daarvan
	// (v = (x - x_prev)/dt)
	// Let op: x_prev = huidige output van xf
	vx := (frame.Position.X - f.x.xf.y) / dt
	vy := (frame.Position.Y - f.y.xf.y) / dt

	ad := alpha(dt, f.DCutoff)
	vxHat := f.x.dx.apply(vx, ad)
	vyHat := f.y.dx.apply(vy, ad)

	// Dynamische cutoff: minCutoff + beta*|v|
	minC := f.MinCutoff
	beta := f.Beta

	cutX := minC + beta*math.Abs(vxHat)
	cutY := minC + beta*math.Abs(vyHat)

	ax := alpha(dt, cutX)
	ay := alpha(dt, cutY)

	xHat := f.x.xf.apply(frame.Position.X, ax)
	yHat := f.y.xf.apply(frame.Position.Y, ay)

	frame.Position = FVec2{X: xHat, Y: yHat}
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
