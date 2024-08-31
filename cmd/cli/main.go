package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	audio "github.com/hailam/malgoplay/internal/fsg"

	"github.com/gen2brain/malgo"
)

func main() {
	var (
		minFreq    float64
		maxFreq    float64
		sampleRate uint
		channels   uint
		duration   int
		sweepRate  float64
		sweepMode  string
	)

	flag.Float64Var(&minFreq, "min", 220, "Minimum frequency")
	flag.Float64Var(&minFreq, "m", 220, "Minimum frequency (shorthand)")
	flag.Float64Var(&maxFreq, "max", 880, "Maximum frequency")
	flag.Float64Var(&maxFreq, "M", 880, "Maximum frequency (shorthand)")
	flag.UintVar(&sampleRate, "rate", 44100, "Sample rate")
	flag.UintVar(&sampleRate, "r", 44100, "Sample rate (shorthand)")
	flag.UintVar(&channels, "channels", 2, "Number of channels")
	flag.UintVar(&channels, "c", 2, "Number of channels (shorthand)")
	flag.IntVar(&duration, "duration", 10, "Duration in seconds")
	flag.IntVar(&duration, "d", 10, "Duration in seconds (shorthand)")
	flag.Float64Var(&sweepRate, "sweep", 1, "Sweep rate in Hz")
	flag.Float64Var(&sweepRate, "s", 1, "Sweep rate in Hz (shorthand)")
	flag.StringVar(&sweepMode, "mode", "linear", "Sweep mode (linear, sine, triangle, exponential, logarithmic, square, sawtooth, random)")
	flag.StringVar(&sweepMode, "o", "linear", "Sweep mode (shorthand)")
	flag.Parse()

	gen := audio.NewFrequencySweepGenerator(minFreq, maxFreq, uint32(sampleRate), uint32(channels))
	defer gen.Close()

	gen.SetSweepRate(sweepRate)

	switch sweepMode {
	case "linear":
		gen.SetSweepMode(audio.SweepModeLinear)
	case "sine":
		gen.SetSweepMode(audio.SweepModeSine)
	case "triangle":
		gen.SetSweepMode(audio.SweepModeTriangle)
	case "exponential":
		gen.SetSweepMode(audio.SweepModeExponential)
	case "logarithmic":
		gen.SetSweepMode(audio.SweepModeLogarithmic)
	case "square":
		gen.SetSweepMode(audio.SweepModeSquare)
	case "sawtooth":
		gen.SetSweepMode(audio.SweepModeSawtooth)
	case "random":
		gen.SetSweepMode(audio.SweepModeRandom)
	default:
		log.Fatalf("Unknown sweep mode: %s", sweepMode)
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("Log: %v\n", message)
	})
	if err != nil {
		log.Fatalf("Failed to initialize context: %v", err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = uint32(channels)
	deviceConfig.SampleRate = uint32(sampleRate)

	// Set the device configuration for the generator
	gen.SetDeviceConfig(deviceConfig)

	// Start the generator, which starts the device internally
	if err := gen.Start(); err != nil {
		log.Fatalf("Failed to start frequency sweep generator: %v", err)
	}

	if duration == 0 {
		fmt.Println("Playing sweep indefinitely. Press Ctrl+C to stop.")
		// Set up channel to listen for interrupt signal
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		// Block until we receive an interrupt signal
		<-c
	} else {
		fmt.Printf("Playing sweep for %d seconds...\n", duration)
		time.Sleep(time.Duration(duration) * time.Second)
	}

	fmt.Println("Stopping playback...")
	if err := gen.Stop(); err != nil {
		log.Fatalf("Failed to stop frequency sweep generator: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
}
