package fsg

import (
	"math"
	"testing"
	"time"

	"github.com/gen2brain/malgo"
)

func TestSetMockDevice(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	mockDevice := NewMockDevice(44100, 2)
	gen.SetMockDevice(mockDevice)

	if gen.device != mockDevice {
		t.Errorf("Mock device not set correctly")
	}

	if !gen.isInitialized {
		t.Errorf("Generator should be marked as initialized after setting mock device")
	}
}

func TestInterpolateFrequency(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)

	testCases := []struct {
		mode     SweepMode
		expected float64
	}{
		{SweepModeLinear, 220},
		{SweepModeSine, 220},
		{SweepModeTriangle, 220},
		{SweepModeExponential, 220},
		{SweepModeLogarithmic, 220},
		{SweepModeSquare, 220},
		{SweepModeSawtooth, 220},
		{SweepModeRandom, 220}, // Random is unpredictable, but test the structure
	}

	for _, tc := range testCases {
		gen.SetSweepMode(tc.mode)
		frequency := gen.interpolateFrequency()
		if tc.mode != SweepModeRandom { // Skip the exact check for random mode
			if math.Abs(frequency-tc.expected) > 1e-5 {
				t.Errorf("Incorrect frequency for mode %v: got %v, want %v", tc.mode, frequency, tc.expected)
			}
		} else {
			if frequency < 220 || frequency > 880 {
				t.Errorf("Frequency out of range for random mode: got %v", frequency)
			}
		}
	}
}

func TestUpdateAmplitude(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)

	// Test fade-in
	gen.isFadingIn = true
	gen.fadeStartTime = time.Now().Add(-gen.fadeInDuration)
	gen.updateAmplitude()
	if gen.currentAmplitude != gen.targetAmplitude {
		t.Errorf("Fade-in did not complete correctly: got %v, want %v", gen.currentAmplitude, gen.targetAmplitude)
	}

	// Test fade-out
	gen.isFadingOut = true
	gen.fadeStartTime = time.Now().Add(-gen.fadeOutDuration)
	gen.updateAmplitude()
	if gen.currentAmplitude != 0 {
		t.Errorf("Fade-out did not complete correctly: got %v, want 0", gen.currentAmplitude)
	}

	// Test gradual change to target amplitude
	gen.isFadingIn = false
	gen.isFadingOut = false
	gen.currentAmplitude = 0.5
	gen.targetAmplitude = 1.0
	gen.updateAmplitude()
	if gen.currentAmplitude <= 0.5 {
		t.Errorf("Amplitude did not increase correctly: got %v, want > 0.5", gen.currentAmplitude)
	}
}

func TestMalgoDeviceWrapper_Uninit(t *testing.T) {
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		println(message)
	})
	if err != nil {
		t.Fatalf("Failed to initialize malgo context: %v", err)
	}
	defer context.Free()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	device, err := malgo.InitDevice(context.Context, deviceConfig, malgo.DeviceCallbacks{})
	if err != nil {
		t.Fatalf("Failed to initialize malgo device: %v", err)
	}

	// Wrap the real device in the MalgoDeviceWrapper
	wrapper := &MalgoDeviceWrapper{Device: device}

	// Start the device (optional but good to test start/stop lifecycle)
	if err := wrapper.Start(); err != nil {
		t.Fatalf("Failed to start device: %v", err)
	}

	// Stop the device before uninitializing
	if err := wrapper.Stop(); err != nil {
		t.Fatalf("Failed to stop device: %v", err)
	}

	// Now, Uninit the device
	err = wrapper.Uninit()
	if err != nil {
		t.Errorf("Uninit should not return an error: %v", err)
	}
}

func TestClose(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	mockDevice := NewMockDevice(44100, 2)
	gen.SetMockDevice(mockDevice)

	// Set short fade durations to avoid long sleep times in tests
	gen.SetFadeDurations(0, 0)

	err := gen.Start()
	if err != nil {
		t.Fatalf("Failed to start generator: %v", err)
	}

	err = gen.Close()
	if err != nil {
		t.Fatalf("Failed to close generator: %v", err)
	}

	if gen.isPlaying {
		t.Errorf("Generator should not be playing after Close()")
	}
}

func TestNewFrequencySweepGenerator(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
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

func TestSetDeviceConfig(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	config := malgo.DefaultDeviceConfig(malgo.Playback)
	config.SampleRate = 48000
	gen.SetDeviceConfig(config)
	if gen.deviceConfig.SampleRate != 48000 {
		t.Errorf("Device config not set correctly: got %v, want 48000", gen.deviceConfig.SampleRate)
	}
}

func TestSetAmplitude(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetAmplitude(0.5)
	if gen.targetAmplitude != 0.5 {
		t.Errorf("Incorrect amplitude: got %v, want 0.5", gen.targetAmplitude)
	}
	gen.SetAmplitude(1.5) // Should be clamped to 1.0
	if gen.targetAmplitude != 1.0 {
		t.Errorf("Amplitude not clamped: got %v, want 1.0", gen.targetAmplitude)
	}
}

func TestSetSweepRate(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetSweepRate(2.0)
	if gen.sweepRate != 2.0 {
		t.Errorf("Incorrect sweep rate: got %v, want 2.0", gen.sweepRate)
	}
}

func TestSetSweepMode(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetSweepMode(SweepModeSine)
	if gen.sweepMode != SweepModeSine {
		t.Errorf("Incorrect sweep mode: got %v, want %v", gen.sweepMode, SweepModeSine)
	}
}

func TestSetFadeDurations(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	gen.SetFadeDurations(1*time.Second, 2*time.Second)
	if gen.fadeInDuration != 1*time.Second || gen.fadeOutDuration != 2*time.Second {
		t.Errorf("Incorrect fade durations: got %v in, %v out, want 1s in, 2s out", gen.fadeInDuration, gen.fadeOutDuration)
	}
}

func TestStartStop(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	mockDevice := NewMockDevice(44100, 2)
	gen.SetMockDevice(mockDevice)

	err := gen.Start()
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
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)
	mockDevice := NewMockDevice(44100, 2)

	// Set the mock device and initialize the generator
	gen.SetMockDevice(mockDevice)
	err := gen.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize generator: %v", err)
	}

	// Manually set the callback on the mock device
	mockDevice.SetCallback(gen.DataCallback)

	err = gen.Start()
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

func TestFrequencySweep(t *testing.T) {
	minFreq, maxFreq := 220.0, 880.0
	sampleRate := uint32(44100)
	channels := uint32(2)
	sweepRate := 1.0 // 1 Hz sweep rate

	gen := NewFrequencySweepGenerator(minFreq, maxFreq, sampleRate, channels)
	mockDevice := NewMockDevice(sampleRate, channels)

	gen.SetMockDevice(mockDevice)
	err := gen.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize generator: %v", err)
	}

	mockDevice.SetCallback(gen.DataCallback)
	gen.SetSweepRate(sweepRate)

	err = gen.Start()
	if err != nil {
		t.Fatalf("Failed to start generator: %v", err)
	}

	// Generate 5 seconds of samples to ensure we complete a full sweep
	sampleDuration := 5 * time.Second
	framesToGenerate := uint32(float64(sampleRate) * sampleDuration.Seconds())
	mockDevice.GenerateSamples(framesToGenerate)

	samples := mockDevice.GetCapturedSamples()
	if len(samples) == 0 {
		t.Fatal("No samples were generated")
	}

	// Analyze frequency at multiple points
	analysisDuration := 0.1 // 100ms for each analysis window
	analysisFrames := int(float64(sampleRate) * analysisDuration)

	for i := 0; i < 5; i++ {
		startFrame := i * int(sampleRate)
		endFrame := startFrame + analysisFrames
		if endFrame > len(samples) {
			endFrame = len(samples)
		}
		freq := estimateFrequency(samples[startFrame:endFrame], int(sampleRate))
		t.Logf("Frequency at %v seconds: %v Hz", float64(i), freq)
	}

	// Check start and end frequencies
	firstFreq := estimateFrequency(samples[:analysisFrames], int(sampleRate))
	lastFreq := estimateFrequency(samples[len(samples)-analysisFrames:], int(sampleRate))

	t.Logf("First frequency: %v Hz, Last frequency: %v Hz", firstFreq, lastFreq)

	tolerance := 20.0 // Allow 20 Hz tolerance
	if math.Abs(firstFreq-minFreq) > tolerance {
		t.Errorf("Initial frequency out of expected range: got %v, want close to %v", firstFreq, minFreq)
	}

	if math.Abs(lastFreq-maxFreq) > tolerance {
		t.Errorf("Final frequency out of expected range: got %v, want close to %v", lastFreq, maxFreq)
	}
}

func TestAmplitudeFade(t *testing.T) {
	// Create a mock device with a sample rate of 44100 and 2 channels (stereo)
	mockDevice := NewMockDevice(44100, 2)

	// Initialize the frequency sweep generator with the mock device
	gen := NewFrequencySweepGenerator(20.0, 20000.0, 44100, 2)
	gen.SetMockDevice(mockDevice)

	// Set fade durations for a quick test
	gen.SetFadeDurations(100*time.Millisecond, 100*time.Millisecond)

	// Start the generator (this will initiate the fade-in)
	if err := gen.Start(); err != nil {
		t.Fatalf("Failed to start generator: %v", err)
	}

	// Simulate the generation of samples
	mockDevice.GenerateSamples(4410) // Generate 100ms worth of samples

	// Check if any samples were generated during fade-in
	capturedSamples := mockDevice.GetCapturedSamples()
	if len(capturedSamples) == 0 {
		t.Errorf("No samples were generated during fade in")
	}

	// Stop the generator to test fade-out
	if err := gen.Stop(); err != nil {
		t.Fatalf("Failed to stop generator: %v", err)
	}

	// Ensure that the generator is not playing after stop
	if gen.isPlaying {
		t.Errorf("Generator is still playing after stop")
	}
}

func TestInitialize(t *testing.T) {
	gen := NewFrequencySweepGenerator(220, 880, 44100, 2)

	// Ensure the generator is not initialized
	if gen.isInitialized {
		t.Fatalf("Generator should not be initialized initially")
	}

	// Initialize the generator
	err := gen.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize generator: %v", err)
	}

	// Check if the generator is marked as initialized
	if !gen.isInitialized {
		t.Fatalf("Generator should be marked as initialized after Initialize()")
	}

	// Check if the default device config is set correctly
	if gen.deviceConfig.Playback.Format != malgo.FormatF32 {
		t.Errorf("Device config format not set correctly: got %v, want %v", gen.deviceConfig.Playback.Format, malgo.FormatF32)
	}

	if gen.deviceConfig.Playback.Channels != gen.channels {
		t.Errorf("Device config channels not set correctly: got %v, want %v", gen.deviceConfig.Playback.Channels, gen.channels)
	}

	if gen.deviceConfig.SampleRate != gen.sampleRate {
		t.Errorf("Device config sample rate not set correctly: got %v, want %v", gen.deviceConfig.SampleRate, gen.sampleRate)
	}

	// Attempt to initialize again and ensure no errors
	err = gen.Initialize()
	if err != nil {
		t.Fatalf("Re-initializing the generator should not fail: %v", err)
	}

	// Ensure the generator is still marked as initialized
	if !gen.isInitialized {
		t.Fatalf("Generator should remain initialized after a second Initialize() call")
	}
}

// Helper function to estimate frequency from samples
func estimateFrequency(samples []float32, sampleRate int) float64 {
	if len(samples) == 0 {
		return 0
	}
	// Simple zero-crossing method
	crossings := 0
	for i := 1; i < len(samples); i++ {
		if (samples[i-1] < 0 && samples[i] >= 0) || (samples[i-1] >= 0 && samples[i] < 0) {
			crossings++
		}
	}
	return float64(crossings) * float64(sampleRate) / float64(2*len(samples))
}
