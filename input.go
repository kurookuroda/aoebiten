package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Input struct {
	gamepadIDs []ebiten.GamepadID
}

func (i *Input) update() {
	i.updateGamepadIDs()
}

func (i *Input) updateGamepadIDs() {
	i.gamepadIDs = i.gamepadIDs[:0]
	for _, id := range ebiten.GamepadIDs() {
		i.gamepadIDs = append(i.gamepadIDs, id)
	}
}

func (i *Input) isKeyOrButtonPressed(key ebiten.Key, btn ebiten.StandardGamepadButton) bool {
	if ebiten.IsKeyPressed(key) {
		return true
	}
	for _, id := range i.gamepadIDs {
		if ebiten.IsStandardGamepadButtonPressed(id, btn) {
			return true
		}
	}
	return false
}

func (i *Input) isKeyOrButtonJustPressed(key ebiten.Key, btn ebiten.StandardGamepadButton) bool {
	if inpututil.IsKeyJustPressed(key) {
		return true
	}
	for _, id := range i.gamepadIDs {
		if inpututil.IsStandardGamepadButtonJustPressed(id, btn) {
			return true
		}
	}
	return false
}

func (i *Input) justTouched() bool {
	ids := inpututil.AppendJustPressedTouchIDs(nil)
	return len(ids) > 0
}

func (i *Input) btnA() bool {
	return i.isKeyOrButtonPressed(ebiten.KeyZ, ebiten.StandardGamepadButtonRightBottom)
}

func (i *Input) btnB() bool {
	return i.isKeyOrButtonPressed(ebiten.KeyX, ebiten.StandardGamepadButtonRightRight)
}

func (i *Input) btnX() bool {
	return i.isKeyOrButtonPressed(ebiten.KeyF, ebiten.StandardGamepadButtonRightLeft)
}

func (i *Input) btnDown() bool {
	return i.isKeyOrButtonPressed(ebiten.KeyDown, ebiten.StandardGamepadButtonLeftBottom)
}

func (i *Input) btnUp() bool {
	return i.isKeyOrButtonPressed(ebiten.KeyUp, ebiten.StandardGamepadButtonLeftTop)
}

func (i *Input) btnpDown() bool {
	return i.isKeyOrButtonJustPressed(ebiten.KeyDown, ebiten.StandardGamepadButtonLeftBottom)
}

func (i *Input) btnpUp() bool {
	return i.isKeyOrButtonJustPressed(ebiten.KeyUp, ebiten.StandardGamepadButtonLeftTop)
}

func (i *Input) btnpX() bool {
	return i.isKeyOrButtonJustPressed(ebiten.KeyF, ebiten.StandardGamepadButtonRightLeft)
}

func (i *Input) btnpSelect() bool {
	return i.isKeyOrButtonJustPressed(ebiten.KeyR, ebiten.StandardGamepadButtonCenterLeft)
}

func (i *Input) btnpStart() bool {
	return i.isKeyOrButtonJustPressed(ebiten.KeyR, ebiten.StandardGamepadButtonCenterRight)
}

func (i *Input) downAlonePressed() bool {
	down := i.btnpDown()
	aHeld := i.btnA()
	bHeld := i.btnB()
	return down && !(aHeld || bHeld)
}

func (i *Input) skipPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		i.justTouched() ||
		i.downAlonePressed()
}

func (i *Input) backPressed() bool {
	return i.btnpUp()
}

func (i *Input) nextPressed() bool {
	return i.downAlonePressed() ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		i.justTouched()
}

func (i *Input) superSpeedCombo() bool {
	return i.btnDown() && i.btnB()
}

func (i *Input) speedCombo() bool {
	return i.btnDown() && i.btnA()
}

func (i *Input) spaceHeld() bool {
	return ebiten.IsKeyPressed(ebiten.KeySpace)
}

func (i *Input) resetPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyR) ||
		i.btnpSelect() || i.btnpStart()
}

func (i *Input) fontTogglePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyF) || i.btnpX()
}
