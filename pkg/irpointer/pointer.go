// Package irpointer contains an algorithm to use your wiimote IR-sensors as pointer on a screen
package irpointer

// Algorithm to process Wiimote IR tracking data into a usable pointer position
// by tracking the sensor bar. Made by "marcan" at
// https://gist.github.com/marcan/c7ca900d5191610957c478bbdbb516c0. Rewritten to Go
// by Friedel Schön.
//
// Copyright (c) 2008-2011 Hector Martin "marcan" <marcan@marcan.st>
// Copyright (c) 2025 Friedel Schön <derfriedmundschoen@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
//    this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

//go:generate stringer -type IRHealth -output stringer.go

import (
	"math"

	"github.com/friedelschoen/go-xwiimote"
)

// FVec2 represents a 2D floating point vector to X and Y.
type FVec2 struct {
	X, Y float64
}

// SensorBar holds state information on the sensor bar
type SensorBar struct {
	// Angle wiimote to sensorbar in radians.
	Angle float64

	dots     [2]FVec2
	accDots  [2]FVec2
	rotDots  [2]FVec2
	offAngle float64
	score    float64
}

// IRPointer holds the current state of the pointer. The smoothed position is
// roughly in the range (-512..512) for both X and Y, where 0 is center and
// 512 is about the maximum offset. The actual returned
// values will not necessarily cover that range (e.g. don't expect more than
// -384..384 for Y if the wiimote is level). This range represents a square
// screen. The values might exceed -512 or 512 under some circumstances.
//
// Keep in mind that you want to map the screen to a subset of this space, both
// because presumably your screen doesn't have a 1:1 aspect ratio, and because
// the Wiimote won't be able to cover the entire space. The worst case scenario
// is when the Wiimote is sideways, where the X coordinate might only cover
// about -384..384, and the corresponding Y coordinate range would only be
// -216..216 for a 16:9 screen. There is a tradeoff here: using a larger range
// means not being able to reach the edges of the screen with the wiimote turned
// sideways, while using a smaller range means the cursor moves faster and is
// harder to control.
//
// For example, a conservative mapping for a 16:9 screen might be:
//
// X left = -340, X right = 340
// If sensor bar is below screen:
//
//	Y top = -290
//	Y bottom = 92
//
// If sensor bar is above screen:
//
//	Y top = -92
//	Y bottom = 290
//
// While a wider mapping might be:
// X left = -430, X right = 430
// If sensor bar is below screen:
//
//	Y top = -290
//	Y bottom = 194
//
// If sensor bar is above screen:
//
//	Y top = -194
//	Y bottom = 290
//
// Wider than the above starts having trouble at the edges.
//
// Notes on signs and ranges:
// Raw Wiimote IR data maps 0,0 to the bottom left corner of the sensor's field
// of view (this corresponds to pointing the wiimote up and to the right). This
// is the format expected in ir.dot. roll should be 0 when the wiimote is
// level and should increase as it is rotated clockwise, covering a -pi to pi
// range. Output data has a positive X when pointing to the right of the sensor
// bar, and a positive Y when pointing under the sensor bar, with 0,0
// corresponding to pointing directly at the sensor bar.
type IRPointer struct {
	SensorBar
	params *IRParams

	// Health of pointer
	Health IRHealth
	// Distance from wiimote to screen in centimeters
	Distance float64
	// Smoothed coordinate
	Position *FVec2

	errorCount  int // Error count from smoothing algorithm
	glitchCount int // Glitch count from smoothing algorithm
}

// IRHealth describes the current tracking of sensorbars
type IRHealth uint

const (
	// The pointer is dead if it never had any signal yet
	IRDead IRHealth = iota
	// The pointer is lost if it had a signal which is now lost
	IRLost
	// The pointer can track one IR signal and must guess the position, a position is available but can be varying
	IRSingle
	// The pointer can track the whole sensorbar (two IR signals) and the position is good
	IRGood
)

type IRParams struct {
	// Height is half-height of the IR sensor if half-width is 1
	Height float64

	// MaxSbSlope is maximum sensor bar slope
	MaxSbSlope float64

	// MinSbWidth is minimum sensor bar width in view, relative to half of the IR sensor area
	MinSbWidth float64

	// SbWidth is physical dimensions center to center of emitters
	SbWidth float64

	// SbDotWidth is half-width of emitters
	SbDotWidth float64

	// SbDotHeight is half-height of emitters (with some tolerance)
	SbDotHeight float64

	// dots further out than these coords are allowed to not be picked up
	// otherwise assume something's wrong
	SbOffScreenX float64
	SbOffScreenY float64

	// SbSingleNoGuessDistance is a point and if it's closer to one of the previous SB points
	// when it reappears, consider it the same instead of trying to guess
	// which one of the two it is
	SbSingleNoGuessDistance float64

	// WiimoteFOVCoefficient is distance from the center of the FOV to the left or right edge,
	// when the wiimote is at one meter
	WiimoteFOVCoefficient float64

	SmootherRadius float64
	SmootherSpeed  float64

	// SmootherDeadzone is distance between old and new value where nothing should change
	SmootherDeadzone float64

	// ErrorMaxCount is max number of errors before cooked data drops out
	ErrorMaxCount int

	// GlitchMaxCount is max number of glitches before cooked data updates
	GlitchMaxCount int

	// GlitchDistance is squared delta over which we consider something a glitch
	GlitchDistance float64
}

var DefaultIRParams = IRParams{
	Height: 384.0 / 512.0,

	MaxSbSlope: 0.7, // : tan(35 degrees)

	MinSbWidth: 0.1,

	SbWidth: 19.5, // cm

	SbDotWidth: 2.25, // cm

	SbDotHeight: 1.0, // cm

	SbOffScreenX: 0.0,
	SbOffScreenY: 0.0,

	// SbOffScreenX :  0.8f,
	// SbOffScreenY :  (0.8f * Height),
	// disable, may be doing more harm than good due to sensor pickup glitches

	SbSingleNoGuessDistance: (100.0 * 100.0),

	WiimoteFOVCoefficient: 0.39, // meter

	SmootherRadius: 8.0,
	SmootherSpeed:  0.25,

	SmootherDeadzone: 2.5, // pixels

	ErrorMaxCount: 8,

	GlitchMaxCount: 5,

	GlitchDistance: (150.0 * 150.0),
}

// rotates dots in `in` into `out`, `out` is expected to be at least as big as `in`
func rotateDots(out []FVec2, in []FVec2, theta float64) {
	// theta=0 doesn't do anything
	if theta == 0 {
		copy(out, in)
		return
	}

	s := math.Sin(theta)
	c := math.Cos(theta)

	for i, dot := range in {
		out[i].X = (c * dot.X) + (-s * dot.Y)
		out[i].Y = (s * dot.X) + (c * dot.Y)
	}
}

func square(f float64) float64 {
	return f * f
}

// NewIRPointer allocates a new IRPointer. params can alter some functionality
// but can also be nil'ed to use default parameters
func NewIRPointer(params *IRParams) *IRPointer {
	ir := &IRPointer{}
	ir.params = &DefaultIRParams
	if params != nil {
		ir.params = params
	}
	ir.errorCount = ir.params.ErrorMaxCount
	return ir
}

func findDots(slots [4]xwiimote.IRSlot) (dots []FVec2) {
	// count visible dots and populate dots structure
	// dots[] is in -1..1 units for width
	for _, slot := range slots {
		if slot.Valid() {
			var dot FVec2
			dot.X = -(float64(slot.X) - 512.0) / 512.0
			dot.Y = (float64(slot.Y) - 384.0) / 512.0
			dots = append(dots, dot)
		}
	}
	return
}

func (ir *IRPointer) findCanditates(dots, accDots []FVec2, roll float64) (candidates []SensorBar) {
	if len(dots) < 2 {
		return nil
	}

	// iterate through all dot pairs
	for first := 0; first < (len(dots) - 1); first++ {
		for second := (first + 1); second < len(dots); second++ {
			// order the dots leftmost first into cand
			// storing both the raw dots and the accel-rotated dots
			var cand SensorBar
			if accDots[first].X > accDots[second].X {
				cand.dots[0] = dots[second]
				cand.dots[1] = dots[first]
				cand.accDots[0] = accDots[second]
				cand.accDots[1] = accDots[first]
			} else {
				cand.dots[0] = dots[first]
				cand.dots[1] = dots[second]
				cand.accDots[0] = accDots[first]
				cand.accDots[1] = accDots[second]
			}
			difference := FVec2{
				X: cand.accDots[1].X - cand.accDots[0].X,
				Y: cand.accDots[1].Y - cand.accDots[0].Y,
			}

			// check angle
			if math.Abs(difference.Y/difference.X) > ir.params.MaxSbSlope {
				continue
			}
			// rotate to the true sensor bar angle
			cand.offAngle = -math.Atan2(difference.Y, difference.X)
			cand.Angle = cand.offAngle + roll
			rotateDots(cand.rotDots[:], cand.dots[:], cand.Angle)
			// recalculate x distance - y should be zero now, so ignore it
			difference.X = cand.rotDots[1].X - cand.rotDots[0].X

			// check distance
			if difference.X < ir.params.MinSbWidth {
				continue
			}

			// middle dot check. If there's another source somewhere in the
			// middle of this candidate, then this can't be a sensor bar
			isBar := true
			for i := range dots {
				var wadj, hadj float64
				if i == first || i == second {
					continue
				}
				hadj = ir.params.SbDotHeight / ir.params.SbWidth * difference.X
				wadj = ir.params.SbDotWidth / ir.params.SbWidth * difference.X
				var tdot [1]FVec2
				rotateDots(tdot[:], dots[i:i+1], cand.Angle)
				if ((cand.rotDots[0].X + wadj) < tdot[0].X) &&
					((cand.rotDots[1].X - wadj) > tdot[0].X) &&
					((cand.rotDots[0].Y + hadj) > tdot[0].Y) &&
					((cand.rotDots[0].Y - hadj) < tdot[0].Y) {
					isBar = false
					break
				}
			}
			// failed middle dot check
			if !isBar {
				continue
			}
			cand.score = 1 / (cand.rotDots[1].X - cand.rotDots[0].X)

			// we have a candidate, store it
			candidates = append(candidates, cand)
		}
	}
	return
}

func (ir *IRPointer) guessSingle(dots, accDots []FVec2, roll float64) (sb SensorBar, ok bool) {
	closest := -1
	closestTo := 0
	best := 999.0
	var d float64
	var dx [2]float64
	var sbx [2]SensorBar
	// no sensor bar candidates, try to work with a lone dot
	switch ir.Health {
	case IRDead:
		// we've never seen a sensor bar before, so we're screwed
		return sb, false
	case IRGood, IRSingle, IRLost:
		// try to find the dot closest to the previous sensor bar
		// position
		for i, accDot := range accDots {
			for j, saccDot := range ir.accDots {
				d = square(accDot.X - saccDot.X)
				d += square(accDot.Y - saccDot.Y)
				if d < best {
					best = d
					closestTo = j
					closest = i
				}
			}
		}
		if ir.Health != IRLost ||
			best < ir.params.SbSingleNoGuessDistance {
			// now work out where the other dot would be, in the acc
			// frame
			sb.accDots[closestTo] = accDots[closest]
			sb.accDots[closestTo^1].X = (ir.accDots[closestTo^1].X -
				ir.accDots[closestTo].X +
				accDots[closest].X)
			sb.accDots[closestTo^1].Y = (ir.accDots[closestTo^1].Y -
				ir.accDots[closestTo].Y +
				accDots[closest].Y)
			// get the raw frame
			rotateDots(sb.dots[:], sb.accDots[:], -roll)
			if (math.Abs(sb.dots[closestTo^1].X) < ir.params.SbOffScreenX) &&
				(math.Abs(sb.dots[closestTo^1].Y) < ir.params.SbOffScreenY) {
				// this dot should be visible but isn't, since the
				// candidate section failed. fall through and try to
				// pick out the sensor bar without previous information
			} else {
				// calculate the rotated dots frame
				// angle tends to drift, so recalculate
				sb.offAngle = -math.Atan2(sb.accDots[1].Y-sb.accDots[0].Y,
					sb.accDots[1].X-sb.accDots[0].X)
				sb.Angle = ir.offAngle + roll
				rotateDots(sb.rotDots[:], sb.accDots[:], ir.offAngle)
				break
			}
		}
		// try to find the dot closest to the sensor edge
		for i, dot := range dots {
			d = min(1.0-math.Abs(dot.X), ir.params.Height-math.Abs(dot.Y))
			if d < best {
				best = d
				closest = i
			}
		}
		// now try it as both places in the sensor bar
		// and pick the one that places the other dot furthest off-screen
		for i := range 2 {
			sbx[i].accDots[i] = accDots[closest]
			sbx[i].accDots[i^1].X = ir.accDots[i^1].X -
				ir.accDots[i].X +
				accDots[closest].X
			sbx[i].accDots[i^1].Y = ir.accDots[i^1].Y -
				ir.accDots[i].Y +
				accDots[closest].Y
			rotateDots(sbx[i].dots[:], sbx[i].accDots[:], -roll)
			dx[i] = max(math.Abs(sbx[i].dots[i^1].X), math.Abs(sbx[i].dots[i^1].Y/ir.params.Height))
		}
		if dx[0] > dx[1] {
			sb = sbx[0]
		} else {
			sb = sbx[1]
		}
		// angle tends to drift, so recalculate
		sb.offAngle = -(math.Atan2(sb.accDots[1].Y-sb.accDots[0].Y,
			sb.accDots[1].X-sb.accDots[0].X))
		sb.Angle = ir.offAngle + roll
		rotateDots(sb.rotDots[:], sb.accDots[:], ir.offAngle)
	}
	sb.score = 0
	return sb, true
}

// raw      *FVec2  // Raw coordinate (-512..512, 0 is center)
// distance float64 // Pixel width of the sensor bar
func (ir *IRPointer) updateSensorbar(slots [4]xwiimote.IRSlot, roll float64) (raw FVec2, ok bool) {
	dots := findDots(slots)

	// nothing to track
	if len(dots) == 0 {
		if ir.Health != IRDead {
			ir.Health = IRLost
		}
		ok = false
		return
	}

	// first rotate according to accelerometer orientation
	accDots := make([]FVec2, len(dots))
	rotateDots(accDots, dots, roll)

	candidates := ir.findCanditates(dots, accDots, roll)
	switch len(candidates) {
	case 0:
		var sb SensorBar
		sb, ok = ir.guessSingle(dots, accDots, roll)
		if !ok {
			return
		}
		ir.SensorBar = sb
		ir.Health = IRSingle
	case 1:
		ir.SensorBar = candidates[0]
		ir.Health = IRGood
	case 2:
		ir.SensorBar = candidates[0]
		/* search for best candidate */
		for i := 1; i < len(candidates); i++ {
			if candidates[i].score > ir.score {
				ir.SensorBar = candidates[i]
			}
		}
		ir.Health = IRGood
	}
	ir.Distance = 50 / (ir.rotDots[1].X - ir.rotDots[0].X)

	raw = FVec2{
		X: (ir.rotDots[0].X + ir.rotDots[1].X) / 2 * 512.0,
		Y: (ir.rotDots[0].Y + ir.rotDots[1].Y) / 2 * 512.0,
	}
	ok = true
	return
}

func (ir *IRPointer) applySmoothing(raw FVec2) {
	dx := raw.X - ir.Position.X
	dy := raw.Y - ir.Position.Y
	d := math.Sqrt(square(dx) + square(dy))
	if d <= ir.params.SmootherDeadzone {
		return
	}
	if d < ir.params.SmootherRadius {
		ir.Position.X += dx * ir.params.SmootherSpeed
		ir.Position.Y += dy * ir.params.SmootherSpeed
	} else {
		theta := math.Atan2(dy, dx)
		ir.Position.X = raw.X - math.Cos(theta)*ir.params.SmootherRadius
		ir.Position.Y = raw.Y - math.Sin(theta)*ir.params.SmootherRadius
	}
}

// Update processes new dots and acceleration values
//
// If acceleration data is unreliable (wiimote is significantly
// accelerating) then you should supply the last known good value.
func (ir *IRPointer) Update(slots [4]xwiimote.IRSlot, accel xwiimote.Vec3) {
	roll := math.Atan2(float64(accel.X), float64(accel.Y))
	ir.UpdateRoll(slots, roll)
}

// UpdateRoll processes new dots and roll values
//
// You can calculate the roll from the accel as roll=atan2(x, z). If roll
// data is unreliable (wiimote is significantly accelerating) then you should
// supply the last known good value.
func (ir *IRPointer) UpdateRoll(slots [4]xwiimote.IRSlot, roll float64) {
	raw, ok := ir.updateSensorbar(slots, roll)

	if !ok {
		if ir.errorCount >= ir.params.ErrorMaxCount {
			ir.Position = nil
		} else {
			ir.errorCount++
		}
		return
	}

	if ir.errorCount >= ir.params.ErrorMaxCount {
		ir.Position = &raw
		ir.glitchCount = 0
	} else {
		d := square(raw.X-ir.Position.X) + square(raw.Y-ir.Position.Y)
		if d > ir.params.GlitchDistance {
			if ir.glitchCount > ir.params.GlitchMaxCount {
				ir.applySmoothing(raw)
				ir.glitchCount = 0
			} else {
				ir.glitchCount++
			}
		} else {
			ir.applySmoothing(raw)
			ir.glitchCount = 0
		}
	}
	ir.errorCount = 0
}
