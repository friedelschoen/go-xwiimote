package xwiimote

import "testing"

func TestIRSlotValid(t *testing.T) {
	slot := IRSlot{Vec2{
		X: 0,
		Y: 0,
	}}

	if !slot.Valid() {
		t.Errorf("IRSlot{%v, %v} should be valid but is not", slot.X, slot.Y)
	}
}

func TestIRSlotInvalid(t *testing.T) {
	slot := IRSlot{Vec2{
		X: 1023,
		Y: 1023,
	}}

	if slot.Valid() {
		t.Errorf("IRSlot{%v, %v} should be invalid but is not", slot.X, slot.Y)
	}
}

func TestIRSlotMixedvalid(t *testing.T) {
	// only if both fields are 1023, the slot is invalid!
	slot := IRSlot{Vec2{
		X: 1023,
		Y: 1024,
	}}

	if !slot.Valid() {
		t.Errorf("IRSlot{%v, %v} should be valid but is not", slot.X, slot.Y)
	}
}
