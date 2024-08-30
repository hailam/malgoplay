package main

import (
	"github.com/hailam/malgoplay/malgoplay"
)

// Expose Go functions to Android via gomobile bindings
func SetFrequency(freq float64) {
	malgoplay.SetFrequency(freq)
}

func SetVolume(vol float64) {
	malgoplay.SetVolume(vol)
}

func GenerateSineWave(frames uint32) []float32 {
	return malgoplay.GenerateSineWave(frames)
}

func DetectFrequency(data []float32) float64 {
	return malgoplay.DetectFrequency(data)
}
