package fsg

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
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

type AudioDevice interface {
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
	device           AudioDevice
	deviceConfig     malgo.DeviceConfig
	isInitialized    bool
	mutex            sync.Mutex
	initMutex        sync.Mutex
	enableLogging    bool
}

func NewFrequencySweepGenerator(minFreq, maxFreq float64, sampleRate, channels uint32) *FrequencySweepGenerator {
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
		isInitialized:    false,
		enableLogging:    false,
	}
}

func (g *FrequencySweepGenerator) SetDeviceConfig(config malgo.DeviceConfig) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.deviceConfig = config
}

func (g *FrequencySweepGenerator) Initialize() error {
	g.initMutex.Lock()
	defer g.initMutex.Unlock()

	if g.isInitialized {
		return nil
	}

	var err error
	g.context, err = malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		g.Log(fmt.Sprintf("Malgo Log: %v", message))
	})
	if err != nil {
		return err
	}

	g.deviceConfig = malgo.DefaultDeviceConfig(malgo.Playback)
	g.deviceConfig.Playback.Format = malgo.FormatF32
	g.deviceConfig.Playback.Channels = g.channels
	g.deviceConfig.SampleRate = g.sampleRate

	device, err := malgo.InitDevice(g.context.Context, g.deviceConfig, malgo.DeviceCallbacks{
		Data: g.DataCallback,
	})
	if err != nil {
		g.context.Free()
		return err
	}

	g.device = &MalgoDeviceWrapper{Device: device}
	g.isInitialized = true
	return nil
}

// New method to set a mock device for testing
func (g *FrequencySweepGenerator) SetMockDevice(device AudioDevice) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.device = device
	g.isInitialized = true
}

func (g *FrequencySweepGenerator) Start() error {
	g.mutex.Lock()
	g.Log("Start: Acquired lock")

	if !g.isInitialized {
		g.Log("Initializing the generator...")
		if err := g.Initialize(); err != nil {
			g.mutex.Unlock()
			return err
		}

		g.Log("Generator initialized successfully.")
	}

	if g.isPlaying {
		g.mutex.Unlock()
		g.Log("Generator is already playing.")
		return nil
	}

	// Ensure device callback is registered and ready before starting playback
	if g.device == nil {
		g.mutex.Unlock()
		return fmt.Errorf("audio device is not initialized")
	}

	// Set isPlaying to true before starting the device
	g.isPlaying = true
	g.isFadingIn = true
	g.isFadingOut = false
	g.fadeStartTime = time.Now()
	g.currentAmplitude = 0

	// Set the callback for the mock device
	if mockDevice, ok := g.device.(*MockDevice); ok {
		g.Log("Setting mock device callback...")
		mockDevice.SetCallback(g.DataCallback)
	}

	time.Sleep(100 * time.Millisecond)
	g.Log("Starting the audio device...")
	err := g.device.Start()
	if err != nil {
		g.isPlaying = false // Revert if the device failed to start
		g.mutex.Unlock()
		g.Log(fmt.Sprintf("Failed to start device: %v", err))
		return err
	}

	g.Log("Generator started. Fade-in begins.")
	g.mutex.Unlock()

	// Optional: Add a small delay if the real device needs setup time
	time.Sleep(50 * time.Millisecond)

	g.isPlaying = true
	g.isFadingIn = true
	g.isFadingOut = false
	g.fadeStartTime = time.Now()
	g.currentAmplitude = 0

	g.Log("Generator is now playing.")
	return nil
}

func (g *FrequencySweepGenerator) Stop() error {
	g.Log("Stop: Attempting to stop generator...")

	g.mutex.Lock()
	g.Log("Stop: Acquired lock")

	if !g.isPlaying {
		g.Log("Generator is not playing.")
		g.mutex.Unlock()
		return nil
	}

	// Start the fade-out process
	g.isFadingOut = true
	g.fadeStartTime = time.Now()
	fadeDuration := g.fadeOutDuration

	g.mutex.Unlock()

	// Wait for the fade-out duration (without holding the lock)
	time.Sleep(fadeDuration)

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Finalize the stopping process
	g.isPlaying = false
	g.isFadingIn = false
	g.isFadingOut = false
	return g.device.Stop()
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
	g.Log("DataCallback: Acquired lock")
	if !g.mutex.TryLock() {
		g.Log("DataCallback: Failed to acquire lock")
		return
	}

	defer func() {
		g.Log("DataCallback: Releasing lock")
		g.mutex.Unlock()
	}()

	g.Log("DataCallback invoked.")

	if !g.isPlaying {
		g.Log("DataCallback invoked, but generator is not playing.")
		return
	}

	g.Log("DataCallback is generating samples.")

	samples := framecount * g.channels
	if samples == 0 {
		g.Log("DataCallback: No samples expected to be generated.")
		return
	}

	output := make([]float32, samples)

	for i := uint32(0); i < samples; i++ {
		g.currentFreq = g.interpolateFrequency()
		g.updateAmplitude()

		g.phase += 2.0 * math.Pi * g.currentFreq / float64(g.sampleRate)
		if g.phase > 2.0*math.Pi {
			g.phase -= 2.0 * math.Pi
		}

		sample := float32(math.Sin(g.phase) * g.currentAmplitude)
		output[i] = sample

		if i%g.channels == 0 {
			switch g.sweepMode {
			case SweepModeLinear, SweepModeExponential, SweepModeLogarithmic:
				g.sweepPhase += g.sweepRate / float64(g.sampleRate)
				if g.sweepPhase > 1.0 {
					g.sweepPhase = 1.0
				}
			case SweepModeTriangle, SweepModeSine:
				g.sweepPhase += float64(g.sweepDirection) * g.sweepRate / float64(g.sampleRate)
				if g.sweepPhase > 1.0 {
					g.sweepPhase = 1.0
					g.sweepDirection = -1
				} else if g.sweepPhase < 0.0 {
					g.sweepPhase = 0.0
					g.sweepDirection = 1
				}
			case SweepModeSawtooth:
				g.sweepPhase += g.sweepRate / float64(g.sampleRate)
				if g.sweepPhase > 1.0 {
					g.sweepPhase = 0.0
				}
			case SweepModeRandom:
				g.sweepPhase = rand.Float64()
			}
		}
	}

	g.Log("DataCallback generated samples.")

	// Convert float32 samples to bytes
	for i, sample := range output {
		binary.LittleEndian.PutUint32(pOutputSample[i*4:(i+1)*4], math.Float32bits(sample))
	}
}

func (g *FrequencySweepGenerator) interpolateFrequency() float64 {
	freqRange := g.maxFrequency - g.minFrequency
	var t float64

	switch g.sweepMode {
	case SweepModeSawtooth, SweepModeLinear:
		t = g.sweepPhase
	case SweepModeSine:
		t = 0.5 + 0.5*math.Sin(math.Pi*(g.sweepPhase-0.5))
	case SweepModeTriangle:
		if g.sweepPhase < 0.5 {
			t = g.sweepPhase * 2.0
			break
		}
		t = 2.0 - g.sweepPhase*2.0
	case SweepModeExponential:
		t = math.Pow(2, g.sweepPhase) - 1
	case SweepModeLogarithmic:
		const base = 10.0
		t = math.Log(g.sweepPhase*(base-1)+1) / math.Log(base)
	case SweepModeSquare:
		if g.sweepPhase < 0.5 {
			t = 0
			break
		}
		t = 1
	case SweepModeRandom:
		g.randomSeed = (g.randomSeed*1103515245 + 12345) & 0x7fffffff
		//randGen := rand.New(rand.NewSource(g.randomSeed))
		//g.randomSeed = randGen.Int63()
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
			g.Log(fmt.Sprintf("Fade-in complete. Amplitude set to target: %v", g.currentAmplitude))
			return
		}
		g.currentAmplitude = g.targetAmplitude * float64(elapsed) / float64(g.fadeInDuration)
		g.Log(fmt.Sprintf("Fading in. Current amplitude: %v", g.currentAmplitude))
		return
	}

	if g.isFadingOut {
		elapsed := time.Since(g.fadeStartTime)
		if elapsed >= g.fadeOutDuration {
			g.currentAmplitude = 0
			g.isFadingOut = false
			g.Log("Fade-out complete. Amplitude set to 0")
			return
		}
		g.currentAmplitude = g.targetAmplitude * (1 - float64(elapsed)/float64(g.fadeOutDuration))
		g.Log(fmt.Sprintf("Fading out. Current amplitude: %v", g.currentAmplitude))
		return
	}

	// Gradual change to target amplitude
	step := 0.001
	if g.currentAmplitude < g.targetAmplitude {
		g.currentAmplitude = math.Min(g.currentAmplitude+step, g.targetAmplitude)
	} else if g.currentAmplitude > g.targetAmplitude {
		g.currentAmplitude = math.Max(g.currentAmplitude-step, g.targetAmplitude)
	}
}

func (g *FrequencySweepGenerator) Close() error {
	g.mutex.Lock()
	isPlaying := g.isPlaying
	g.mutex.Unlock()

	// Stop the generator if it is currently playing
	if isPlaying {
		if err := g.Stop(); err != nil {
			return err
		}
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Uninitialize the device if it exists
	if g.device != nil {
		g.device.Uninit()
		g.device = nil
	}

	// Free the context if it exists
	if g.context != nil {
		g.context.Free()
		g.context = nil
	}

	g.isInitialized = false
	return nil
}

func (g *FrequencySweepGenerator) Log(message string) {
	if g.enableLogging {
		fmt.Println(message)
	}
}
