package xwiimote

import (
	"log"
	"math"
	"slices"
)

/*
 * Algorithm to process Wiimote IR tracking data into a usable pointer position
 * by tracking the sensor bar. Made by "marcan" at
 * https://gist.github.com/marcan/c7ca900d5191610957c478bbdbb516c0. Rewritten to Go
 * by Friedel Schön.
 *
 * Copyright (c) 2008-2011 Hector Martin "marcan" <marcan@marcan.st>
 * Copyright (c) 2025 Friedel Schön <derfriedmundschoen@gmail.com>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

// FVec2 represents a 2D floating point vector to X and Y.
type FVec2 struct {
	X, Y float64
}

// Holds state information on the sensor bar
type SensorBar struct {
	Angle    float64
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

	// Internal Health
	Health   IRHealth
	Distance float64 // Pixel width of the sensor bar
	Smooth   *FVec2  // Smoothed coordinate

	errorCount  float64 // Error count from smoothing algorithm
	glitchCount float64 // Glitch count from smoothing algorithm
}

type IRHealth uint

const (
	IRDead IRHealth = iota
	IRLost
	IRSingle
	IRGood
)

// Height is half-height of the IR sensor if half-width is 1
const Height = 384.0 / 512.0

// MaxSbSlope is maximum sensor bar slope
const MaxSbSlope = 0.7 // = tan(35 degrees)

// MinSbWidth is minimum sensor bar width in view, relative to half of the IR sensor area
const MinSbWidth = 0.1

// SbWidth is physical dimensions center to center of emitters
const SbWidth = 19.5 // cm

// SbDotWidth is half-width of emitters
const SbDotWidth = 2.25 // cm

// SbDotHeight is half-height of emitters (with some tolerance)
const SbDotHeight = 1.0 // cm

const SbDotWidthRatio = SbDotWidth / SbWidth
const SbDotHeightRatio = SbDotHeight / SbWidth

// dots further out than these coords are allowed to not be picked up
// otherwise assume something's wrong
//const SB_OFF_SCREEN_X =  0.8f
//const SB_OFF_SCREEN_Y =  (0.8f * HEIGHT)
// disable, may be doing more harm than good due to sensor pickup glitches

const SbOffScreenX = 0.0
const SbOffScreenY = 0.0

// SbSingleNoGuessDistance is a point and if it's closer to one of the previous SB points
// when it reappears, consider it the same instead of trying to guess
// which one of the two it is
const SbSingleNoGuessDistance = (100.0 * 100.0)

// SbZCoefficient is width of the sensor bar at one meter from the Wiimote
const SbZCoefficient = 256.0 // pixels

// WiimoteFOVCoefficient is distance from the center of the FOV to the left or right edge,
// when the wiimote is at one meter
const WiimoteFOVCoefficient = 0.39 // meter

const SmootherRadius = 8.0
const SmootherSpeed = 0.25

// SmootherDeadzone is distance between old and new value where nothing should change
const SmootherDeadzone = 2.5 // pixels

// ErrorMaxCount is max number of errors before cooked data drops out
const ErrorMaxCount = 8

// GlitchMaxCount is max number of glitches before cooked data updates
const GlitchMaxCount = 5

// GlitchDistance is squared delta over which we consider something a glitch
const GlitchDistance = (150.0 * 150.0)

// Debug sets if debug messages should be printed
const Debug = false

// rotates dots in `in` into `out`, `out` is expected to be at least as big as `in`
func rotateDots(out []FVec2, in []FVec2, theta float64) {
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

func printDebug(format string, values ...any) {
	if Debug {
		log.Printf(format, values...)
	}
}

func NewIRPointer() *IRPointer {
	ir := &IRPointer{}
	ir.errorCount = ErrorMaxCount
	return ir
}

func findDots(slots [4]IRSlot) (dots []FVec2) {
	// count visible dots and populate dots structure
	// dots[] is in -1..1 units for width
	for _, slot := range slots {
		if slot.Valid() {
			var dot FVec2
			dot.X = -(float64(slot.X) - 512.0) / 512.0
			dot.Y = (float64(slot.Y) - 384.0) / 512.0
			printDebug("IR: dot %d at (%d,%d) (%.03f,%.03f)\n",
				len(dots), slot.X, slot.Y,
				dot.X, dot.Y)
			dots = append(dots, dot)
		}
	}

	printDebug("IR: found %d dots\n", len(dots))
	return
}

func findCanditates(dots, accDots []FVec2, roll float64) (candidates []SensorBar) {
	if len(dots) < 2 {
		return nil
	}

	printDebug("IR: locating sensor bar candidates\n")

	// iterate through all dot pairs
	for first := 0; first < (len(dots) - 1); first++ {
		for second := (first + 1); second < len(dots); second++ {
			printDebug("IR: trying dots %d and %d\n", first, second)
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
			if math.Abs(difference.Y/difference.X) > MaxSbSlope {
				continue
			}
			printDebug("IR: passed angle check\n")
			// rotate to the true sensor bar angle
			cand.offAngle = -math.Atan2(difference.Y, difference.X)
			cand.Angle = cand.offAngle + roll
			rotateDots(cand.rotDots[:], cand.dots[:], cand.Angle)
			printDebug("IR: off_angle: %.02f, angle: %.02f\n",
				cand.offAngle, cand.Angle)
			// recalculate x distance - y should be zero now, so ignore it
			difference.X = cand.rotDots[1].X - cand.rotDots[0].X

			// check distance
			if difference.X < MinSbWidth {
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
				hadj = SbDotHeightRatio * difference.X
				wadj = SbDotWidthRatio * difference.X
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
			printDebug("IR: passed middle dot check\n")

			cand.score = 1 / (cand.rotDots[1].X - cand.rotDots[0].X)

			// we have a candidate, store it
			printDebug("IR: new candidate %d\n", len(candidates))
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
	printDebug("IR: no candidates\n")
	switch ir.Health {
	case IRDead:
		printDebug("IR: we're dead\n")
		// we've never seen a sensor bar before, so we're screwed
		return sb, false
	case IRGood, IRSingle, IRLost:
		printDebug("IR: trying to keep track of single dot\n")
		// try to find the dot closest to the previous sensor bar
		// position
		for i, accDot := range accDots {
			printDebug("IR: checking dot %d (%.02f, %.02f)\n",
				i, accDot.X, accDot.Y)
			for j, saccDot := range ir.SensorBar.accDots {
				printDebug("      to dot %d (%.02f, %.02f)\n",
					j, saccDot.X, saccDot.Y)
				d = square(accDot.X - saccDot.X)
				d += square(accDot.Y - saccDot.Y)
				if d < best {
					best = d
					closestTo = j
					closest = i
				}
			}
		}
		printDebug("IR: closest dot is %d to %d\n",
			closest, closestTo)
		if ir.Health != IRLost ||
			best < SbSingleNoGuessDistance {
			// now work out where the other dot would be, in the acc
			// frame
			sb.accDots[closestTo] = accDots[closest]
			sb.accDots[closestTo^1].X = (ir.SensorBar.accDots[closestTo^1].X -
				ir.SensorBar.accDots[closestTo].X +
				accDots[closest].X)
			sb.accDots[closestTo^1].Y = (ir.SensorBar.accDots[closestTo^1].Y -
				ir.SensorBar.accDots[closestTo].Y +
				accDots[closest].Y)
			// get the raw frame
			rotateDots(sb.dots[:], sb.accDots[:], -roll)
			if (math.Abs(sb.dots[closestTo^1].X) < SbOffScreenX) &&
				(math.Abs(sb.dots[closestTo^1].Y) < SbOffScreenY) {
				// this dot should be visible but isn't, since the
				// candidate section failed. fall through and try to
				// pick out the sensor bar without previous information
				printDebug("IR: dot falls on screen, falling through\n")
			} else {
				// calculate the rotated dots frame
				// angle tends to drift, so recalculate
				sb.offAngle = -math.Atan2(sb.accDots[1].Y-sb.accDots[0].Y,
					sb.accDots[1].X-sb.accDots[0].X)
				sb.Angle = ir.SensorBar.offAngle + roll
				rotateDots(sb.rotDots[:], sb.accDots[:], ir.SensorBar.offAngle)
				printDebug("IR: kept track of single dot\n")
				break
			}
		} else {
			printDebug("IR: lost the dot and new one is too far away\n")
		}
		// try to find the dot closest to the sensor edge
		printDebug("IR: trying to find best dot\n")
		for i, dot := range dots {
			d = min(1.0-math.Abs(dot.X), Height-math.Abs(dot.Y))
			if d < best {
				best = d
				closest = i
			}
		}
		printDebug("IR: best dot: %d\n", closest)
		// now try it as both places in the sensor bar
		// and pick the one that places the other dot furthest off-screen
		for i := range 2 {
			sbx[i].accDots[i] = accDots[closest]
			sbx[i].accDots[i^1].X = ir.SensorBar.accDots[i^1].X -
				ir.SensorBar.accDots[i].X +
				accDots[closest].X
			sbx[i].accDots[i^1].Y = ir.SensorBar.accDots[i^1].Y -
				ir.SensorBar.accDots[i].Y +
				accDots[closest].Y
			rotateDots(sbx[i].dots[:], sbx[i].accDots[:], -roll)
			dx[i] = max(math.Abs(sbx[i].dots[i^1].X), math.Abs(sbx[i].dots[i^1].Y/Height))
		}
		if dx[0] > dx[1] {
			printDebug("IR: dot is LEFT: %.02f > %.02f\n",
				dx[0], dx[1])
			sb = sbx[0]
		} else {
			printDebug("IR: dot is RIGHT: %.02f < %.02f\n",
				dx[0], dx[1])
			sb = sbx[1]
		}
		// angle tends to drift, so recalculate
		sb.offAngle = -(math.Atan2(sb.accDots[1].Y-sb.accDots[0].Y,
			sb.accDots[1].X-sb.accDots[0].X))
		sb.Angle = ir.SensorBar.offAngle + roll
		rotateDots(sb.rotDots[:], sb.accDots[:], ir.SensorBar.offAngle)
		printDebug("IR: found new dot to track\n")
	}
	sb.score = 0
	return sb, true
}

// raw      *FVec2  // Raw coordinate (-512..512, 0 is center)
// distance float64 // Pixel width of the sensor bar
func (ir *IRPointer) updateSensorbar(slots [4]IRSlot, roll float64) (raw FVec2, distance float64, ok bool) {
	printDebug("IR: angle: %.05f\n", roll)

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

	candidates := findCanditates(dots, accDots, roll)

	var sb SensorBar
	if len(candidates) == 0 {
		sb, ok = ir.guessSingle(dots, accDots, roll)
		if !ok {
			return
		}
		ir.Health = IRSingle
	} else {
		printDebug("IR: finding best candidate\n")
		sb = slices.MaxFunc(candidates, func(left, right SensorBar) int {
			return int(right.score - left.score)
		})
		ir.Health = IRGood
	}

	ir.SensorBar = sb

	raw = FVec2{
		X: ((sb.rotDots[0].X + sb.rotDots[1].X) / 2) * 512.0,
		Y: ((sb.rotDots[0].Y + sb.rotDots[1].Y) / 2) * 512.0,
	}
	distance = (sb.rotDots[1].X - sb.rotDots[0].X) * 512.0
	ok = true
	return
}

func (ir *IRPointer) applySmoothing(raw FVec2) {
	printDebug("Smooth: OK (%.02f, %.02f) LAST (%.02f, %.02f) ",
		raw.X, raw.Y, ir.Smooth.X, ir.Smooth.Y)
	dx := raw.X - ir.Smooth.X
	dy := raw.Y - ir.Smooth.Y
	d := math.Sqrt(square(dx) + square(dy))
	if d <= SmootherDeadzone {
		printDebug("DEADZONE\n")
		return
	}
	if d < SmootherRadius {
		printDebug("INSIDE\n")
		ir.Smooth.X += dx * SmootherSpeed
		ir.Smooth.Y += dy * SmootherSpeed
	} else {
		printDebug("OUTSIDE\n")
		theta := math.Atan2(dy, dx)
		ir.Smooth.X = raw.X - math.Cos(theta)*SmootherRadius
		ir.Smooth.Y = raw.Y - math.Sin(theta)*SmootherRadius
	}
}

// Update processes new dots and roll values
//
// You can calculate the roll from the accel as roll=atan2(x, z). If roll
// data is unreliable (wiimote is significantly accelerating) then you should
// supply the last known good value.
func (ir *IRPointer) Update(slots [4]IRSlot, roll float64) {
	raw, distance, ok := ir.updateSensorbar(slots, roll)

	if ok {
		ir.Distance = distance
		if ir.errorCount >= ErrorMaxCount {
			ir.Smooth = &raw
			ir.glitchCount = 0
		} else {
			d := square(raw.X-ir.Smooth.X) + square(raw.Y-ir.Smooth.Y)
			if d > GlitchDistance {
				if ir.glitchCount > GlitchMaxCount {
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
	} else {
		if ir.errorCount >= ErrorMaxCount {
			ir.Smooth = nil
		} else {
			ir.errorCount++
		}
	}
}

// WiimoteDistance returns distances between Wiimote to sensor bar in meters
func (ir *IRPointer) WiimoteDistance() float64 {
	return SbZCoefficient / ir.Distance
}
