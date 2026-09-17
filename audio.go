package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const audioSampleRate = 44100

type SoundManager struct {
	context       *audio.Context
	sndTalk       *audio.Player
	sndTalkSpace  *audio.Player
	sndTalkFast   *audio.Player
	sndTalkFaster *audio.Player
}

func newSoundManager() (*SoundManager, error) {
	ctx := audio.NewContext(audioSampleRate)
	sm := &SoundManager{context: ctx}

	var err error
	// 音の長さを十分に確保（WASMでは短すぎると再生されない）
	sm.sndTalk, err = createSquareWavePlayer(ctx, 261.63, 0.12, 0.5) // C4
	if err != nil {
		return nil, fmt.Errorf("sndTalk: %w", err)
	}
	sm.sndTalkSpace, err = createSquareWavePlayer(ctx, 329.63, 0.12, 0.5) // E4
	if err != nil {
		return nil, fmt.Errorf("sndTalkSpace: %w", err)
	}
	sm.sndTalkFast, err = createSquareWavePlayer(ctx, 392.00, 0.12, 0.5) // G4
	if err != nil {
		return nil, fmt.Errorf("sndTalkFast: %w", err)
	}
	sm.sndTalkFaster, err = createSquareWavePlayer(ctx, 523.25, 0.12, 0.5) // C5
	if err != nil {
		return nil, fmt.Errorf("sndTalkFaster: %w", err)
	}
	return sm, nil
}

func createSquareWavePlayer(ctx *audio.Context, freq, durationSec, duty float64) (*audio.Player, error) {
	wavData := generateSquareWaveWAV(freq, durationSec, duty)
	stream, err := wav.Decode(ctx, bytes.NewReader(wavData))
	if err != nil {
		return nil, fmt.Errorf("wav decode: %w", err)
	}
	player, err := ctx.NewPlayer(stream)
	if err != nil {
		return nil, fmt.Errorf("new player: %w", err)
	}
	return player, nil
}

func generateSquareWaveWAV(freq, durationSec, duty float64) []byte {
	sampleRate := float64(audioSampleRate)
	numSamples := int(durationSec * sampleRate)
	if numSamples < 1 {
		numSamples = 1
	}
	data := make([]int16, numSamples)
	period := sampleRate / freq
	amp := int16(28000) // 最大振幅に近づける（32767がmax）

	fadeStart := int(float64(numSamples) * 0.8) // 最後20%でフェードアウト

	for i := 0; i < numSamples; i++ {
		phase := (float64(i) / period) - math.Floor(float64(i)/period)
		var sample int16
		if phase < duty {
			sample = amp
		} else {
			sample = -amp
		}

		// フェードアウトでクリックノイズ防止
		if i >= fadeStart {
			fadeRatio := float64(numSamples-i) / float64(numSamples-fadeStart)
			sample = int16(float64(sample) * fadeRatio)
		}
		data[i] = sample
	}

	byteData := make([]byte, len(data)*2)
	for i, v := range data {
		binary.LittleEndian.PutUint16(byteData[i*2:], uint16(v))
	}

	return buildWAVHeader(byteData, 1, 16, audioSampleRate)
}

func buildWAVHeader(pcmData []byte, channels, bitsPerSample, sampleRate int) []byte {
	dataLen := len(pcmData)
	header := &bytes.Buffer{}

	header.WriteString("RIFF")
	binary.Write(header, binary.LittleEndian, uint32(36+dataLen))
	header.WriteString("WAVE")

	header.WriteString("fmt ")
	binary.Write(header, binary.LittleEndian, uint32(16)) // Subchunk1Size
	binary.Write(header, binary.LittleEndian, uint16(1))  // AudioFormat = PCM
	binary.Write(header, binary.LittleEndian, uint16(channels))
	binary.Write(header, binary.LittleEndian, uint32(sampleRate))
	binary.Write(header, binary.LittleEndian, uint32(sampleRate*channels*bitsPerSample/8)) // ByteRate
	binary.Write(header, binary.LittleEndian, uint16(channels*bitsPerSample/8))            // BlockAlign
	binary.Write(header, binary.LittleEndian, uint16(bitsPerSample))

	header.WriteString("data")
	binary.Write(header, binary.LittleEndian, uint32(dataLen))
	header.Write(pcmData)

	return header.Bytes()
}

func (sm *SoundManager) playTalk(interval, charsPerTick int) {
	var p *audio.Player
	switch {
	case interval == 0 && charsPerTick >= 3:
		p = sm.sndTalkFaster
	case interval == 0 && charsPerTick == 1:
		p = sm.sndTalkFast
	case interval == FastInterval:
		p = sm.sndTalkSpace
	default:
		p = sm.sndTalk
	}

	if p != nil {
		p.Rewind()
		p.Play()
	}
}
