package malgoplay_android_main

import (
	"log"
	"time"

	"github.com/gen2brain/malgo"
	"github.com/hailam/malgoplay/malgoplay"
)

var (
	device *malgo.Device
	//shouldRun   bool
	//shouldRunMu sync.Mutex
)

// Expose Go functions to Android via gomobile bindings

func SetFrequency(freq float64) {
	malgoplay.SetFrequency(freq)
}

func SetSweepRate(rate float64) {
	malgoplay.SweepRate = rate
}

func SetAmplitude(amp float64) {
	malgoplay.Amplitude = amp
}

func SetMinFrequency(freq float64) {
	malgoplay.MinFrequency = freq
}

func SetVolume(vol float64) {
	malgoplay.SetVolume(vol)
}

func GetFrequency() float64 {
	return malgoplay.GetFrequency()
}

func GetVolume() float64 {
	return malgoplay.GetVolume()
}

func GetMinFrequency() float64 {
	return malgoplay.MinFrequency
}

func GetMaxFrequency() float64 {
	return malgoplay.MaxFrequency
}

func ReverseSweepDirection() {
	malgoplay.SweepDirection *= -1
}

func UpdateFrequency() {
	malgoplay.UpdateFrequency()
}

func InitDevice() error {
	var err error
	device, err = malgoplay.InitDevice()
	if err != nil {
		log.Printf("InitDevice: %v", err)
		return err
	}

	return nil
}

func CleanupDevice() {
	if device != nil {
		malgoplay.CleanupDevice(device)
		device = nil
	}
}

func StartDevice() error {
	if device != nil {
		//shouldRunMu.Lock()
		//shouldRun = true
		//shouldRunMu.Unlock()

		err := device.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func StopDevice() error {
	if device != nil {
		//shouldRunMu.Lock()
		//shouldRun = false
		//shouldRunMu.Unlock()

		// Gradually decrease the volume before stopping
		for i := 100; i >= 0; i-- {
			malgoplay.SetVolume(float64(i) / 100.0)
			time.Sleep(10 * time.Millisecond)
		}

		return device.Stop()
	}
	return nil
}
