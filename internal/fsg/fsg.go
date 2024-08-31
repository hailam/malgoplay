package fsg

import (
	"encoding/binary"
	"math"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
)

type SweepMode int

const (
	SweepModeLinear SweepMode = iota
	SweepModeSine
	SweepModeTriangle
	SweepModeExponential
	SweepModeLogarithmic
	SweepModeSquare
	SweepModeSawtooth
	SweepModeRandom
)

type audioDevice interface {
	Start() error
	Stop() error
	Uninit() error
}

type FrequencySweepGenerator struct {
	minFrequency     float64
	maxFrequency     float64
	sampleRate       uint32
	channels         uint32
	isPlaying        bool
	currentFreq      float64
	phase            float64
	currentAmplitude float64
	targetAmplitude  float64
	sweepRate        float64
	sweepMode        SweepMode
	sweepPhase       float64
	sweepDirection   int
	fadeInDuration   time.Duration
	fadeOutDuration  time.Duration
	fadeStartTime    time.Time
	isFadingIn       bool
	isFadingOut      bool
	randomSeed       int64
	context          *malgo.AllocatedContext
	device           audioDevice
	mutex            sync.Mutex
}

func NewFrequencySweepGenerator(minFreq, maxFreq float64, sampleRate, channels uint32) (*FrequencySweepGenerator, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		println(message)
	})
	if err != nil {
		return nil, err
	}

	return &FrequencySweepGenerator{
		minFrequency:     minFreq,
		maxFrequency:     maxFreq,
		sampleRate:       sampleRate,
		channels:         channels,
		currentFreq:      minFreq,
		currentAmplitude: 0,
		targetAmplitude:  1,
		sweepRate:        1,
		sweepMode:        SweepModeLinear,
		sweepDirection:   1,
		fadeInDuration:   500 * time.Millisecond,
		fadeOutDuration:  500 * time.Millisecond,
		randomSeed:       time.Now().UnixNano(),
		context:          ctx,
	}, nil
}

func (g *FrequencySweepGenerator) Start(device audioDevice) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.isPlaying {
		return nil
	}

	g.device = device
	err := g.device.Start()
	if err != nil {
		return err
	}

	g.isPlaying = true
	g.isFadingIn = true
	g.isFadingOut = false
	g.fadeStartTime = time.Now()
	g.currentAmplitude = 0

	return nil
}

func (g *FrequencySweepGenerator) Stop() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.isPlaying {
		return nil
	}

	g.isFadingOut = true
	g.fadeStartTime = time.Now()

	// Wait for fade out to complete
	time.Sleep(g.fadeOutDuration)

	g.isPlaying = false
	err := g.device.Stop()
	if err != nil {
		return err
	}

	g.device.Uninit()
	g.device = nil

	return nil
}

func (g *FrequencySweepGenerator) SetAmplitude(amplitude float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.targetAmplitude = math.Max(0, math.Min(1, amplitude))
}

func (g *FrequencySweepGenerator) SetSweepRate(rate float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.sweepRate = rate
}

func (g *FrequencySweepGenerator) SetSweepMode(mode SweepMode) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.sweepMode = mode
	if mode == SweepModeRandom {
		g.randomSeed = time.Now().UnixNano()
	}
}

func (g *FrequencySweepGenerator) SetFadeDurations(fadeIn, fadeOut time.Duration) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.fadeInDuration = fadeIn
	g.fadeOutDuration = fadeOut
}

func (g *FrequencySweepGenerator) DataCallback(pOutputSample, pInputSamples []byte, framecount uint32) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	samples := framecount * g.channels
	output := make([]float32, samples)

	for i := uint32(0); i < samples; i += g.channels {
		g.currentFreq = g.interpolateFrequency()
		g.updateAmplitude()

		g.phase += 2.0 * math.Pi * g.currentFreq / float64(g.sampleRate)
		if g.phase > 2.0*math.Pi {
			g.phase -= 2.0 * math.Pi
		}

		sample := float32(math.Sin(g.phase) * g.currentAmplitude)

		for c := uint32(0); c < g.channels; c++ {
			output[i+c] = sample
		}

		g.sweepPhase += float64(g.sweepDirection) * g.sweepRate / float64(g.sampleRate)
		if g.sweepPhase > 1.0 || g.sweepPhase < 0.0 {
			g.sweepDirection *= -1
			g.sweepPhase = math.Max(0.0, math.Min(1.0, g.sweepPhase))
		}
	}

	// Convert float32 samples to bytes
	for i, sample := range output {
		binary.LittleEndian.PutUint32(pOutputSample[i*4:(i+1)*4], math.Float32bits(sample))
	}
}

func (g *FrequencySweepGenerator) interpolateFrequency() float64 {
	freqRange := g.maxFrequency - g.minFrequency
	var t float64

	switch g.sweepMode {
	case SweepModeLinear:
		t = g.sweepPhase
	case SweepModeSine:
		t = 0.5 + 0.5*math.Sin(math.Pi*(g.sweepPhase-0.5))
	case SweepModeTriangle:
		if g.sweepPhase < 0.5 {
			t = g.sweepPhase * 2.0
		} else {
			t = 2.0 - g.sweepPhase*2.0
		}
	case SweepModeExponential:
		t = math.Pow(2, g.sweepPhase) - 1
	case SweepModeLogarithmic:
		t = math.Log2(1 + g.sweepPhase)
	case SweepModeSquare:
		if g.sweepPhase < 0.5 {
			t = 0
		} else {
			t = 1
		}
	case SweepModeSawtooth:
		t = g.sweepPhase
	case SweepModeRandom:
		g.randomSeed = (g.randomSeed*1103515245 + 12345) & 0x7fffffff
		t = float64(g.randomSeed) / float64(0x7fffffff)
	}

	return g.minFrequency + t*freqRange
}

func (g *FrequencySweepGenerator) updateAmplitude() {
	if g.isFadingIn {
		elapsed := time.Since(g.fadeStartTime)
		if elapsed >= g.fadeInDuration {
			g.currentAmplitude = g.targetAmplitude
			g.isFadingIn = false
			return
		}

		g.currentAmplitude = g.targetAmplitude * float64(elapsed) / float64(g.fadeInDuration)
		return
	}

	if g.isFadingOut {
		elapsed := time.Since(g.fadeStartTime)
		if elapsed >= g.fadeOutDuration {
			g.currentAmplitude = 0
			g.isFadingOut = false
			return
		}

		g.currentAmplitude = g.targetAmplitude * (1 - float64(elapsed)/float64(g.fadeOutDuration))
		return
	}

	// Gradual change to target amplitude
	step := 0.001 // Adjust this value to control the speed of amplitude change
	if g.currentAmplitude < g.targetAmplitude {
		g.currentAmplitude = math.Min(g.currentAmplitude+step, g.targetAmplitude)
	} else if g.currentAmplitude > g.targetAmplitude {
		g.currentAmplitude = math.Max(g.currentAmplitude-step, g.targetAmplitude)
	}
}

func (g *FrequencySweepGenerator) Close() error {
	g.Stop()
	g.context.Free()
	return nil
}
