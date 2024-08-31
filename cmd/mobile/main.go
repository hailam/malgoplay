//go:build (linux && cgo) || (darwin && cgo) || windows
// +build linux,cgo darwin,cgo windows

package mobile_fsg_main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	audio "github.com/hailam/malgoplay/internal/fsg"
)

var (
	gen       *audio.FrequencySweepGenerator
	mutex     sync.Mutex
	isPlaying bool
	stopChan  chan struct{}
)

// InitializeAudio sets up the audio generator and device
func InitializeAudio(minFreq, maxFreq float64, sampleRate, channels int) error {
	mutex.Lock()
	defer mutex.Unlock()

	sampleRate32 := uint32(sampleRate)
	channels32 := uint32(channels)

	gen = audio.NewFrequencySweepGenerator(minFreq, maxFreq, sampleRate32, channels32)

	if err := gen.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize audio: %w", err)
	}

	stopChan = make(chan struct{})
	return nil
}

// StartAudio begins playback
func StartAudio(durationSeconds int) error {
	mutex.Lock()
	defer mutex.Unlock()

	if isPlaying {
		return errors.New("audio is already playing")
	}

	if gen == nil {
		return errors.New("audio not initialized")
	}

	if err := gen.Start(); err != nil {
		fmt.Println("Failed to start device", err)
		return fmt.Errorf("failed to start device: %w", err)
	}

	isPlaying = true

	go func() {
		if durationSeconds > 0 {
			select {
			case <-time.After(time.Duration(durationSeconds) * time.Second):
			case <-stopChan:
			}
		} else {
			<-stopChan
		}

		mutex.Lock()
		if isPlaying {
			_ = gen.Stop()
			isPlaying = false
		}
		mutex.Unlock()
	}()

	return nil
}

// StopAudio stops playback
func StopAudio() error {
	mutex.Lock()
	defer mutex.Unlock()

	if !isPlaying {
		return nil
	}

	close(stopChan)
	isPlaying = false
	return gen.Stop()
}

// SetSweepRate sets the sweep rate
func SetSweepRate(rate float64) {
	mutex.Lock()
	defer mutex.Unlock()

	if gen != nil {
		gen.SetSweepRate(rate)
	}
}

// SetSweepMode sets the sweep mode
func SetSweepMode(mode int) {
	mutex.Lock()
	defer mutex.Unlock()

	if gen != nil {
		gen.SetSweepMode(audio.SweepMode(mode))
	}
}

// IsPlaying returns the current playback state
func IsPlaying() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return isPlaying
}

// CleanupAudio frees resources
func CleanupAudio() {
	mutex.Lock()
	defer mutex.Unlock()

	if isPlaying {
		_ = StopAudio()
	}

	if gen != nil {
		gen.Close()
		gen = nil
	}

	isPlaying = false
}
