package irpointer

import (
	"math"
	"testing"

	xwiimote "github.com/friedelschoen/go-xwiimote"
)

// These tests validate the IR pointer algorithm as a pure state machine +
// geometry transform. They avoid depending on real devices and focus on
// expected behavior and invariants:
//
// - Coordinate mapping (IR slots -> normalized dots)
// - Candidate sensorbar selection logic
// - Health transitions (dead/lost/single/good)
// - Smoothing, deadzone, glitch filtering, and dropout on repeated errors

const (
	epsFloat = 1e-8
)

// Helpers

func almost(a, b float64) bool {
	return math.Abs(a-b) <= epsFloat
}

func almostVec(a, b FVec2) bool {
	return almost(a.X, b.X) && almost(a.Y, b.Y)
}

func norm2(v FVec2) float64 {
	return v.X*v.X + v.Y*v.Y
}

func mkSlotValid(x, y int32) xwiimote.IRSlot {
	return xwiimote.IRSlot{Vec2: xwiimote.Vec2{X: x, Y: y}}
}

func mkSlotInvalid() xwiimote.IRSlot {
	// Your IRSlot.Valid() treats (1023,1023) as invalid ("no track").
	return xwiimote.IRSlot{Vec2: xwiimote.Vec2{X: 1023, Y: 1023}}
}

func mkSlots(valid ...xwiimote.IRSlot) [4]xwiimote.IRSlot {
	var s [4]xwiimote.IRSlot
	for i := range s {
		s[i] = mkSlotInvalid()
	}
	for i := 0; i < len(valid) && i < 4; i++ {
		s[i] = valid[i]
	}
	return s
}

func copyDots(in []FVec2) []FVec2 {
	out := make([]FVec2, len(in))
	copy(out, in)
	return out
}

// rotateDots

func TestRotateDots_ThetaZeroIsCopy(t *testing.T) {
	in := []FVec2{{1, 2}, {-3, 4}, {0.5, -0.25}}
	out := make([]FVec2, len(in))
	rotateDots(out, in, 0)

	for i := range in {
		if !almostVec(out[i], in[i]) {
			t.Fatalf("index %d: expected %v, got %v", i, in[i], out[i])
		}
	}
}

func TestRotateDots_PreservesNormAndInverse(t *testing.T) {
	in := []FVec2{{1, 2}, {-3, 4}, {0.5, -0.25}}
	tmp := make([]FVec2, len(in))
	out := make([]FVec2, len(in))

	theta := 0.42
	rotateDots(tmp, in, theta)
	rotateDots(out, tmp, -theta)

	for i := range in {
		// Inverse rotation should recover the original vector (within tolerance).
		if math.Abs(out[i].X-in[i].X) > 1e-7 || math.Abs(out[i].Y-in[i].Y) > 1e-7 {
			t.Fatalf("index %d: inverse mismatch: in=%v out=%v", i, in[i], out[i])
		}
		// Rotation should preserve vector length.
		if math.Abs(norm2(tmp[i])-norm2(in[i])) > 1e-7 {
			t.Fatalf("index %d: norm not preserved: in=%v tmp=%v", i, in[i], tmp[i])
		}
	}
}

// findDots

func TestFindDots_FiltersInvalidAndMapsCoordinates(t *testing.T) {
	// Mapping:
	// dot.X = -(x - 512)/512
	// dot.Y =  (y - 384)/512
	slots := mkSlots(
		mkSlotValid(512, 384),  // center -> (0,0)
		mkSlotValid(0, 384),    // left edge -> X = +1
		mkSlotValid(1023, 384), // near right edge -> X ~ -0.9980
		// last slot invalid
	)

	dots := findDots(slots)
	if len(dots) != 3 {
		t.Fatalf("expected 3 dots, got %d: %v", len(dots), dots)
	}

	if !almostVec(dots[0], FVec2{0, 0}) {
		t.Fatalf("expected center dot (0,0), got %v", dots[0])
	}

	if !almost(dots[1].X, 1.0) || !almost(dots[1].Y, 0.0) {
		t.Fatalf("expected dot[1] approx (1,0), got %v", dots[1])
	}

	wantX := -(float64(1023) - 512.0) / 512.0
	if !almost(dots[2].X, wantX) || !almost(dots[2].Y, 0.0) {
		t.Fatalf("expected dot[2] approx (%v,0), got %v", wantX, dots[2])
	}
}

// findCanditates

func TestFindCandidates_GoodBarOneCandidate(t *testing.T) {
	ir := NewIRPointer(nil)

	// Two dots horizontally aligned => slope ~0 and width > MinSbWidth.
	dots := []FVec2{{-0.4, 0.0}, {0.4, 0.0}}
	accDots := copyDots(dots)

	cands := ir.findCanditates(dots, accDots, 0)
	if len(cands) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(cands))
	}

	sb := cands[0]
	if sb.rotDots[1].X <= sb.rotDots[0].X {
		t.Fatalf("expected ordered rotDots, got %+v", sb.rotDots)
	}
	if sb.score <= 0 {
		t.Fatalf("expected positive score, got %v", sb.score)
	}
}

func TestFindCandidates_TooSteepRejected(t *testing.T) {
	params := DefaultIRParams
	params.MaxSbSlope = 0.2 // make slope check stricter
	ir := NewIRPointer(&params)

	dots := []FVec2{{-0.4, -0.4}, {0.4, 0.4}} // slope ~ 1.0
	accDots := copyDots(dots)

	cands := ir.findCanditates(dots, accDots, 0)
	if len(cands) != 0 {
		t.Fatalf("expected 0 candidates, got %d", len(cands))
	}
}

func TestFindCandidates_TooNarrowRejected(t *testing.T) {
	params := DefaultIRParams
	params.MinSbWidth = 0.9 // wider than our dot separation
	ir := NewIRPointer(&params)

	dots := []FVec2{{-0.2, 0.0}, {0.2, 0.0}} // width 0.4
	accDots := copyDots(dots)

	cands := ir.findCanditates(dots, accDots, 0)
	if len(cands) != 0 {
		t.Fatalf("expected 0 candidates, got %d", len(cands))
	}
}

// func TestFindCandidates_MiddleDotReject(t *testing.T) {
// 	params := DefaultIRParams
// 	// Increase dot extents to make the middle-dot rejection box very large,
// 	// guaranteeing that a dot in the middle will be considered inside.
// 	params.SbDotWidth = 10.0
// 	params.SbDotHeight = 10.0
// 	ir := NewIRPointer(&params)

// 	// Left + middle + right dot: pairs should be rejected due to a middle dot inside.
// 	dots := []FVec2{{-0.6, 0.0}, {0.0, 0.0}, {0.6, 0.0}}
// 	accDots := copyDots(dots)

// 	cands := ir.findCanditates(dots, accDots, 0)
// 	if len(cands) != 0 {
// 		t.Fatalf("expected 0 candidates due to middle-dot check, got %d", len(cands))
// 	}
// }

// updateSensorbar + health

func TestUpdateSensorbar_NoDotsHealthTransitions(t *testing.T) {
	ir := NewIRPointer(nil)

	// Starting IRDead with no dots should remain IRDead and report ok=false.
	raw, ok := ir.updateSensorbar(mkSlots(), 0)
	if ok {
		t.Fatalf("expected ok=false with no dots, got ok=true raw=%v", raw)
	}
	if ir.Health != IRDead {
		t.Fatalf("expected Health=IRDead, got %v", ir.Health)
	}

	// If we previously had a signal, losing all dots should transition to IRLost.
	ir.Health = IRGood
	raw, ok = ir.updateSensorbar(mkSlots(), 0)
	if ok {
		t.Fatalf("expected ok=false with no dots, got ok=true raw=%v", raw)
	}
	if ir.Health != IRLost {
		t.Fatalf("expected Health=IRLost after losing signal, got %v", ir.Health)
	}
}

func TestUpdateSensorbar_TwoDotsGivesGoodAndDistance(t *testing.T) {
	ir := NewIRPointer(nil)

	// Two visible points at y=384 => normalized Y ~ 0.
	slots := mkSlots(
		mkSlotValid(400, 384),
		mkSlotValid(624, 384),
	)

	raw, ok := ir.updateSensorbar(slots, 0)
	if !ok {
		t.Fatalf("expected ok=true, got ok=false")
	}
	if ir.Health != IRGood {
		t.Fatalf("expected Health=IRGood, got %v", ir.Health)
	}
	if ir.Distance <= 0 {
		t.Fatalf("expected positive Distance, got %v", ir.Distance)
	}
	if math.Abs(raw.Y) > 1e-6 {
		t.Fatalf("expected raw.Y ~ 0, got %v", raw.Y)
	}
}

// Candidate selection correctness (best score within the same frame)

// func TestUpdateSensorbar_SelectsBestCandidateInFrame(t *testing.T) {
// 	// This test constructs a dot set that yields multiple valid candidates.
// 	// The expected behavior is to pick the candidate with the highest score,
// 	// where score = 1/width in rotated frame (narrower bar => higher score).

// 	params := DefaultIRParams
// 	params.MaxSbSlope = 10.0  // permissive
// 	params.MinSbWidth = 0.01  // permissive
// 	params.SbDotWidth = 0.01  // small, so middle-dot check is less likely to reject
// 	params.SbDotHeight = 0.01 // small
// 	ir := NewIRPointer(&params)

// 	// Create 4 dots on a line, producing multiple pairs.
// 	// Narrow pair: x = -0.05 and +0.05 => width 0.10 => score 10
// 	// Wide pair:   x = -0.40 and +0.40 => width 0.80 => score 1.25
// 	// We want the algorithm to choose the narrow pair.
// 	dots := []FVec2{
// 		{-0.40, 0.00},
// 		{-0.05, 0.00},
// 		{+0.05, 0.00},
// 		{+0.40, 0.00},
// 	}
// 	accDots := copyDots(dots)

// 	cands := ir.findCanditates(dots, accDots, 0)
// 	if len(cands) < 2 {
// 		t.Fatalf("expected at least 2 candidates, got %d", len(cands))
// 	}

// 	// Compute the best expected score directly from candidates (ground truth).
// 	bestIdx := 0
// 	bestScore := cands[0].score
// 	for i := 1; i < len(cands); i++ {
// 		if cands[i].score > bestScore {
// 			bestScore = cands[i].score
// 			bestIdx = i
// 		}
// 	}
// 	want := cands[bestIdx]

// 	// Now feed equivalent data through updateSensorbar via IR slots.
// 	//
// 	// We need to map normalized dot coordinates back into slot coordinates.
// 	// findDots uses:
// 	//   dot.X = -(x - 512)/512  => x = 512 - dot.X*512
// 	//   dot.Y =  (y - 384)/512  => y = 384 + dot.Y*512
// 	//
// 	// Here dot.Y=0, so y=384.
// 	toSlot := func(d FVec2) xwiimote.IRSlot {
// 		x := int32(math.Round(512.0 - d.X*512.0))
// 		y := int32(384.0) // since d.Y=0
// 		// Clamp to valid sensor ranges.
// 		if x < 0 {
// 			x = 0
// 		}
// 		if x > 1023 {
// 			x = 1023
// 		}
// 		return mkSlotValid(x, y)
// 	}

// 	slots := mkSlots(
// 		toSlot(dots[0]),
// 		toSlot(dots[1]),
// 		toSlot(dots[2]),
// 		toSlot(dots[3]),
// 	)

// 	raw, ok := ir.updateSensorbar(slots, 0)
// 	if !ok {
// 		t.Fatalf("expected ok=true")
// 	}
// 	if ir.Health != IRGood {
// 		t.Fatalf("expected Health=IRGood, got %v", ir.Health)
// 	}
// 	_ = raw // we mainly care about selected SensorBar here.

// 	// Verify that the chosen SensorBar has the best score from this frame.
// 	// Use score comparison rather than exact dot equality to avoid float brittleness.
// 	if math.Abs(ir.SensorBar.score-want.score) > 1e-6 {
// 		t.Fatalf("expected best candidate score %v, got %v (candidates=%d)", want.score, ir.SensorBar.score, len(cands))
// 	}

// 	// Additionally, verify that chosen width corresponds to the maximum score.
// 	chosenWidth := ir.SensorBar.rotDots[1].X - ir.SensorBar.rotDots[0].X
// 	wantWidth := want.rotDots[1].X - want.rotDots[0].X
// 	if math.Abs(chosenWidth-wantWidth) > 1e-6 {
// 		t.Fatalf("expected chosen width %v, got %v", wantWidth, chosenWidth)
// 	}
// }

// Smoothing + glitch + dropout

// func TestUpdateRoll_SmoothingDeadzoneAndMovement(t *testing.T) {
// 	params := DefaultIRParams
// 	// Make initialization deterministic and reduce noise in the test.
// 	params.ErrorMaxCount = 1
// 	params.SmootherDeadzone = 10.0
// 	params.SmootherRadius = 8.0
// 	params.SmootherSpeed = 0.5
// 	ir := NewIRPointer(&params)

// 	// Initialize position with a valid sensorbar.
// 	slots1 := mkSlots(
// 		mkSlotValid(400, 384),
// 		mkSlotValid(624, 384),
// 	)
// 	ir.UpdateRoll(slots1, 0)
// 	if ir.Position == nil {
// 		t.Fatalf("expected Position initialized")
// 	}
// 	start := *ir.Position

// 	// Small movement below deadzone should not change Position.
// 	slots2 := mkSlots(
// 		mkSlotValid(402, 384),
// 		mkSlotValid(626, 384),
// 	)
// 	ir.UpdateRoll(slots2, 0)
// 	if ir.Position == nil {
// 		t.Fatalf("expected Position not nil")
// 	}
// 	if math.Abs(ir.Position.X-start.X) > 1e-6 || math.Abs(ir.Position.Y-start.Y) > 1e-6 {
// 		t.Fatalf("expected no movement in deadzone, start=%v now=%v", start, *ir.Position)
// 	}

// 	// A larger change should move Position.
// 	slots3 := mkSlots(
// 		mkSlotValid(200, 300),
// 		mkSlotValid(824, 468),
// 	)
// 	ir.UpdateRoll(slots3, 0)
// 	if ir.Position == nil {
// 		t.Fatalf("expected Position not nil")
// 	}
// 	if math.Abs(ir.Position.X-start.X) < 1e-3 && math.Abs(ir.Position.Y-start.Y) < 1e-3 {
// 		t.Fatalf("expected Position to move with large change, start=%v now=%v", start, *ir.Position)
// 	}
// }

// func TestUpdateRoll_GlitchFiltering(t *testing.T) {
// 	params := DefaultIRParams
// 	params.ErrorMaxCount = 1
// 	params.GlitchMaxCount = 2
// 	params.GlitchDistance = 1.0 // very low so almost any jump counts as glitch
// 	params.SmootherDeadzone = 0
// 	params.SmootherRadius = 8
// 	params.SmootherSpeed = 1.0
// 	ir := NewIRPointer(&params)

// 	// Initialize position.
// 	stable := mkSlots(
// 		mkSlotValid(400, 384),
// 		mkSlotValid(624, 384),
// 	)
// 	ir.UpdateRoll(stable, 0)
// 	if ir.Position == nil {
// 		t.Fatalf("expected Position initialized")
// 	}
// 	base := *ir.Position

// 	// Large jump that will be treated as a glitch.
// 	jump := mkSlots(
// 		mkSlotValid(100, 100),
// 		mkSlotValid(900, 700),
// 	)

// 	// For the first GlitchMaxCount+? updates, position should not update.
// 	// After passing the threshold, smoothing should apply and the position should move.
// 	ir.UpdateRoll(jump, 0)
// 	pos1 := *ir.Position
// 	ir.UpdateRoll(jump, 0)
// 	pos2 := *ir.Position

// 	if math.Abs(pos1.X-base.X) > 1e-3 || math.Abs(pos1.Y-base.Y) > 1e-3 {
// 		t.Fatalf("expected glitch filtered (no move on first), base=%v pos1=%v", base, pos1)
// 	}
// 	if math.Abs(pos2.X-base.X) > 1e-3 || math.Abs(pos2.Y-base.Y) > 1e-3 {
// 		t.Fatalf("expected glitch filtered (no move on second), base=%v pos2=%v", base, pos2)
// 	}

// 	// Threshold passed: now movement should occur.
// 	ir.UpdateRoll(jump, 0)
// 	pos3 := *ir.Position
// 	if almostVec(base, pos3) {
// 		t.Fatalf("expected movement after glitch threshold, base=%v pos3=%v", base, pos3)
// 	}
// }

func TestUpdateRoll_ErrorDropoutSetsPositionNil(t *testing.T) {
	params := DefaultIRParams
	params.ErrorMaxCount = 3
	ir := NewIRPointer(&params)

	// Initialize position first.
	okSlots := mkSlots(
		mkSlotValid(400, 384),
		mkSlotValid(624, 384),
	)
	ir.UpdateRoll(okSlots, 0)
	if ir.Position == nil {
		t.Fatalf("expected Position initialized")
	}

	// After enough consecutive failures, Position should be dropped (nil).
	none := mkSlots()
	ir.UpdateRoll(none, 0)
	ir.UpdateRoll(none, 0)
	ir.UpdateRoll(none, 0)
	ir.UpdateRoll(none, 0)

	if ir.Position != nil {
		t.Fatalf("expected Position=nil after enough errors, got %v", *ir.Position)
	}
}
