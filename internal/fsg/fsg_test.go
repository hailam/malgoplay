package fsg

import (
	"math"
	"testing"
	"time"

	"github.com/gen2brain/malgo"
)

func TestNewFrequencySweepGenerator(t *testing.T) {
	gen, err := NewFrequencySweepGenerator(220, 880, 44100, 2)
	if err != nil {
		t.Fatalf("Failed to create new generator: %v", err)
	}
	if gen == nil {
		t.Fatal("Generator is nil")
	}
	if gen.minFrequency != 220 || gen.maxFrequency != 880 {
		t.Errorf("Incorrect frequency range: got %v-%v, want 220-880", gen.minFrequency, gen.maxFrequency)
	}
	if gen.sampleRate != 44100 || gen.channels != 2 {
		t.Errorf("Incorrect audio settings: got %v Hz, %v channels, want 44100 Hz, 2 channels", gen.sampleRate, gen.channels)
	}
}

func TestSetAmplitude(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetAmplitude(0.5)
	if gen.targetAmplitude != 0.5 {
		t.Errorf("Incorrect amplitude: got %v, want 0.5", gen.targetAmplitude)
	}
	gen.SetAmplitude(1.5) // Should be clamped to 1.0
	if gen.targetAmplitude != 1.0 {
		t.Errorf("Amplitude not clamped: got %v, want 1.0", gen.targetAmplitude)
	}
	gen.SetAmplitude(-0.5) // Should be clamped to 0.0
	if gen.targetAmplitude != 0.0 {
		t.Errorf("Amplitude not clamped: got %v, want 0.0", gen.targetAmplitude)
	}
}

func TestSetSweepRate(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetSweepRate(2.0)
	if gen.sweepRate != 2.0 {
		t.Errorf("Incorrect sweep rate: got %v, want 2.0", gen.sweepRate)
	}
}

func TestSetSweepMode(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetSweepMode(SweepModeSine)
	if gen.sweepMode != SweepModeSine {
		t.Errorf("Incorrect sweep mode: got %v, want %v", gen.sweepMode, SweepModeSine)
	}
}

func TestSetFadeDurations(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetFadeDurations(1*time.Second, 2*time.Second)
	if gen.fadeInDuration != 1*time.Second || gen.fadeOutDuration != 2*time.Second {
		t.Errorf("Incorrect fade durations: got %v in, %v out, want 1s in, 2s out", gen.fadeInDuration, gen.fadeOutDuration)
	}
}

func TestInterpolateFrequency(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	testCases := []struct {
		mode     SweepMode
		phase    float64
		expected float64
	}{
		{SweepModeLinear, 0.5, 550},
		{SweepModeSine, 0.5, 550},
		{SweepModeTriangle, 0.25, 550},
		{SweepModeExponential, 0.5, 493.38095116624275},
		{SweepModeLogarithmic, 0.5, 606.0752504759631},
		{SweepModeSquare, 0.25, 220},
		{SweepModeSquare, 0.75, 880},
		{SweepModeSawtooth, 0.5, 550},
	}

	for _, tc := range testCases {
		gen.sweepMode = tc.mode
		gen.sweepPhase = tc.phase
		result := gen.interpolateFrequency()
		if math.Abs(result-tc.expected) > 0.001 {
			t.Errorf("Incorrect frequency for mode %v at phase %v: got %v, want %v", tc.mode, tc.phase, result, tc.expected)
		}
	}
}

func TestUpdateAmplitude(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.targetAmplitude = 1.0
	gen.currentAmplitude = 0.0
	gen.isFadingIn = true
	gen.fadeInDuration = 1 * time.Second
	gen.fadeStartTime = time.Now().Add(-500 * time.Millisecond)

	gen.updateAmplitude()

	if math.Abs(gen.currentAmplitude-0.5) > 0.001 {
		t.Errorf("Incorrect amplitude during fade in: got %v, want 0.5", gen.currentAmplitude)
	}

	gen.isFadingIn = false
	gen.isFadingOut = true
	gen.fadeOutDuration = 1 * time.Second
	gen.fadeStartTime = time.Now().Add(-750 * time.Millisecond)

	gen.updateAmplitude()

	if math.Abs(gen.currentAmplitude-0.25) > 0.001 {
		t.Errorf("Incorrect amplitude during fade out: got %v, want 0.25", gen.currentAmplitude)
	}
}

func TestStartStop(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	mockDevice := NewMockDevice(44100, 2)

	err := gen.Start(mockDevice)
	if err != nil {
		t.Fatalf("Failed to start generator: %v", err)
	}
	if !gen.isPlaying {
		t.Error("Generator should be playing after Start()")
	}

	err = gen.Stop()
	if err != nil {
		t.Fatalf("Failed to stop generator: %v", err)
	}
	if gen.isPlaying {
		t.Error("Generator should not be playing after Stop()")
	}
}

func TestDataCallback(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	mockDevice := NewMockDevice(44100, 2)

	// Set up the data callback
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = gen.channels
	deviceCallbacks := malgo.DeviceCallbacks{
		Data: gen.DataCallback,
	}
	mockDevice.SetDataCallback(deviceCallbacks.Data)

	err := gen.Start(mockDevice)
	if err != nil {
		t.Fatalf("Failed to start generator: %v", err)
	}

	// Generate some samples
	mockDevice.GenerateSamples(1000)

	samples := mockDevice.GetCapturedSamples()
	if len(samples) != 2000 { // 1000 frames * 2 channels
		t.Errorf("Unexpected number of samples: got %d, want 2000", len(samples))
	}

	// Check if samples are within expected range (-1 to 1)
	for _, sample := range samples {
		if sample < -1 || sample > 1 {
			t.Errorf("Sample out of range: %f", sample)
		}
	}

	// Optional: Add more specific checks here, e.g., checking for frequency content
}

func TestSweepPhaseProgression(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetSweepRate(1) // 1 Hz sweep rate

	// Simulate 1 second of audio generation
	for i := 0; i < 44100; i++ {
		gen.interpolateFrequency()
		gen.sweepPhase += 1.0 / 44100.0
		if gen.sweepPhase > 1.0 {
			gen.sweepPhase -= 1.0
		}
	}

	if math.Abs(gen.sweepPhase-0.0) > 0.01 {
		t.Errorf("Sweep phase did not complete full cycle: got %v, want close to 0", gen.sweepPhase)
	}
}

func TestRandomSweep(t *testing.T) {
	gen, _ := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetSweepMode(SweepModeRandom)

	// Generate a sequence of random frequencies
	frequencies := make([]float64, 100)
	for i := 0; i < 100; i++ {
		frequencies[i] = gen.interpolateFrequency()
	}

	// Check that we have a range of different frequencies
	min, max := frequencies[0], frequencies[0]
	for _, f := range frequencies {
		if f < min {
			min = f
		}
		if f > max {
			max = f
		}
	}

	if max-min < (880-220)/2 {
		t.Errorf("Random sweep doesn't cover enough of the frequency range. Min: %v, Max: %v", min, max)
	}

	// Check that we don't have long sequences of the same frequency
	sameCount := 1
	maxSameCount := 1
	for i := 1; i < len(frequencies); i++ {
		if frequencies[i] == frequencies[i-1] {
			sameCount++
			if sameCount > maxSameCount {
				maxSameCount = sameCount
			}
		} else {
			sameCount = 1
		}
	}

	if maxSameCount > 10 {
		t.Errorf("Random sweep has too many consecutive same frequencies. Max consecutive: %v", maxSameCount)
	}
}
