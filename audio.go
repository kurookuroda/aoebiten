package main

import (
	"bytes"
	"encoding/binary"
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
	sm.sndTalk, err = createSquareWavePlayer(ctx, 130.81, 0.08, 0.5)
	if err != nil {
		return nil, err
	}
	sm.sndTalkSpace, err = createSquareWavePlayer(ctx, 130.81, 0.10, 0.5)
	if err != nil {
		return nil, err
	}
	sm.sndTalkFast, err = createSquareWavePlayer(ctx, 130.81, 0.06, 0.5)
	if err != nil {
		return nil, err
	}
	sm.sndTalkFaster, err = createSquareWavePlayer(ctx, 130.81, 0.04, 0.5)
	if err != nil {
		return nil, err
	}
	return sm, nil
}

func createSquareWavePlayer(ctx *audio.Context, freq, durationSec, duty float64) (*audio.Player, error) {
	wavData := generateSquareWaveWAV(freq, durationSec, duty)
	stream, err := wav.Decode(ctx, bytes.NewReader(wavData))
	if err != nil {
		return nil, err
	}
	player, err := ctx.NewPlayer(stream)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func generateSquareWaveWAV(freq, durationSec, duty float64) []byte {
	sampleRate := float64(audioSampleRate)
	numSamples := int(durationSec * sampleRate)
	data := make([]int16, numSamples)
	period := sampleRate / freq
	amp := int16(8000)

	for i := 0; i < numSamples; i++ {
		phase := (float64(i) / period) - math.Floor(float64(i)/period)
		if phase < duty {
			data[i] = amp
		} else {
			data[i] = -amp
		}
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
	binary.Write(header, binary.LittleEndian, uint32(16))
	binary.Write(header, binary.LittleEndian, uint16(1))
	binary.Write(header, binary.LittleEndian, uint16(channels))
	binary.Write(header, binary.LittleEndian, uint32(sampleRate))
	binary.Write(header, binary.LittleEndian, uint32(sampleRate*channels*bitsPerSample/8))
	binary.Write(header, binary.LittleEndian, uint16(channels*bitsPerSample/8))
	binary.Write(header, binary.LittleEndian, uint16(bitsPerSample))

	header.WriteString("data")
	binary.Write(header, binary.LittleEndian, uint32(dataLen))
	header.Write(pcmData)

	return header.Bytes()
}

func (sm *SoundManager) playTalk(interval, charsPerTick int) {
	switch {
	case interval == 0 && charsPerTick >= 3:
		sm.sndTalkFaster.Rewind()
		sm.sndTalkFaster.Play()
	case interval == 0 && charsPerTick == 1:
		sm.sndTalkFast.Rewind()
		sm.sndTalkFast.Play()
	case interval == FastInterval:
		sm.sndTalkSpace.Rewind()
		sm.sndTalkSpace.Play()
	default:
		sm.sndTalk.Rewind()
		sm.sndTalk.Play()
	}
}
