package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hailam/malgoplay/malgoplay"
)

var (
	shouldRun = true
)

func flagVar(p *float64, name string, shorthand string, value float64, usage string) {
	if shorthand != "" {
		flag.Float64Var(p, shorthand, value, usage)
	}
	flag.Float64Var(p, name, value, usage)
}

func initFlags() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flagVar(&malgoplay.MaxFrequency, "f", "frequency", 1000, "Maximum frequency of the sine wave in Hz")
	flagVar(&malgoplay.Amplitude, "a", "amplitude", 0.5, "Amplitude of the sine wave")
	flagVar(&malgoplay.SampleRate, "r", "sample-rate", 48000, "Sample rate in Hz")
	flagVar(&malgoplay.MinFrequency, "m", "min-frequency", 0, "Minimum frequency to start sweeping from in Hz")
	flagVar(&malgoplay.SweepRate, "s", "sweep-rate", 1.0, "Frequency change rate in Hz per second")

	flag.Parse()
}

func main() {
	initFlags()

	if malgoplay.MinFrequency == 0 {
		malgoplay.MinFrequency = malgoplay.MaxFrequency // If no min frequency is set, use a fixed frequency
	}
	malgoplay.SetFrequency(malgoplay.MinFrequency) // Start from the minimum frequency

	device, err := malgoplay.InitDevice()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer malgoplay.CleanupDevice(device)

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if malgoplay.MinFrequency < malgoplay.MaxFrequency {
		fmt.Printf("Sweeping frequency from %f Hz to %f Hz. Press Ctrl+C to stop.\n", malgoplay.MinFrequency, malgoplay.MaxFrequency)
	} else {
		fmt.Printf("Playing and analyzing %f Hz. Press Ctrl+C to stop.\n", malgoplay.GetFrequency())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for shouldRun {
		select {
		case <-sig:
			fmt.Println("\nGot interrupt signal. Stopping...")
			for i := 100; i >= 0; i-- {
				malgoplay.SetVolume(float64(i) / 100.0)
				time.Sleep(10 * time.Millisecond)
			}
			shouldRun = false
		case <-ticker.C:
			if malgoplay.MinFrequency < malgoplay.MaxFrequency {
				malgoplay.UpdateFrequency()
			}
		}
	}

	time.Sleep(time.Second) // intentional
	fmt.Println("Playback and analysis stopped.")
}
