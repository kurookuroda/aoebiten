//go:build js

package main

import (
	"syscall/js"
)

type SoundManager struct {
	context js.Value
}

func newSoundManager() (*SoundManager, error) {
	// WASMではAudioContextの作成を遅延させる
	// 実際の作成は最初のplayTone呼び出し時に行う
	return &SoundManager{}, nil
}

func (sm *SoundManager) ensureContext() {
	if sm.context.IsUndefined() || sm.context.IsNull() {
		audioCtxClass := js.Global().Get("AudioContext")
		if audioCtxClass.IsUndefined() || audioCtxClass.IsNull() {
			audioCtxClass = js.Global().Get("webkitAudioContext")
		}
		if !audioCtxClass.IsUndefined() && !audioCtxClass.IsNull() {
			sm.context = audioCtxClass.New()
		}
	}
}

func (sm *SoundManager) playTone(freq float64, durationSec float64) {
	sm.ensureContext()
	if sm.context.IsUndefined() || sm.context.IsNull() {
		return
	}

	// AudioContextがsuspendedならresume
	if sm.context.Get("state").String() == "suspended" {
		sm.context.Call("resume")
	}

	ctx := sm.context
	osc := ctx.Call("createOscillator")
	gain := ctx.Call("createGain")

	osc.Set("type", "square")
	osc.Get("frequency").Set("value", freq)

	now := ctx.Get("currentTime").Float()
	gain.Get("gain").Call("setValueAtTime", 0.1, now)
	gain.Get("gain").Call("exponentialRampToValueAtTime", 0.01, now+durationSec)

	osc.Call("connect", gain)
	gain.Call("connect", ctx.Get("destination"))

	osc.Call("start", now)
	osc.Call("stop", now+durationSec)
}

func (sm *SoundManager) playTalk(interval, charsPerTick int) {
	switch {
	case interval == 0 && charsPerTick >= 3:
		sm.playTone(523.25, 0.05) // C5
	case interval == 0 && charsPerTick == 1:
		sm.playTone(392.00, 0.05) // G4
	case interval == FastInterval:
		sm.playTone(329.63, 0.05) // E4
	default:
		sm.playTone(261.63, 0.05) // C4
	}
}
