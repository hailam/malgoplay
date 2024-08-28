package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/gen2brain/malgo"
	"gonum.org/v1/gonum/dsp/fourier"
)

var (
	sampleRate      = float64(48000)
	channels        = uint32(1)
	frequency       float64
	minFrequency    float64
	maxFrequency    float64
	amplitude       = 0.5
	phase           float64
	detectedFreqs   = make([]float64, 0, 10)
	detectedFreqsMu sync.Mutex
	shouldRun       = true
	sweepDirection  = 1   // 1 for increasing, -1 for decreasing
	sweepRate       = 1.0 // Hz per second
	frequencyMu     sync.RWMutex
	volumeMu        sync.RWMutex
	volume          = 1.0
)

// Calibration data
var calibration = map[float64]float64{
	20:    0.01,
	100:   1.0,
	300:   1.0,
	1000:  1.1,
	5000:  1.0,
	10000: 1.0,
	15000: 1.0,
	20000: 1.0,
}

func interpolateCalibration(freq float64) float64 {
	keys := make([]float64, 0, len(calibration))
	for k := range calibration {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	if freq <= keys[0] {
		return calibration[keys[0]]
	}
	if freq >= keys[len(keys)-1] {
		return calibration[keys[len(keys)-1]]
	}

	for i := 0; i < len(keys)-1; i++ {
		if keys[i] <= freq && freq < keys[i+1] {
			t := (freq - keys[i]) / (keys[i+1] - keys[i])
			return calibration[keys[i]]*(1-t) + calibration[keys[i+1]]*t
		}
	}
	return 1.0
}

func generateSineWave(frames uint32) []float32 {
	data := make([]float32, frames)
	frequencyMu.RLock()
	currentFrequency := frequency
	frequencyMu.RUnlock()
	volumeMu.RLock()
	currentVolume := volume
	volumeMu.RUnlock()
	for i := range data {
		t := float64(i) / sampleRate
		data[i] = float32(amplitude * currentVolume * math.Sin(2*math.Pi*currentFrequency*t+phase))
	}
	phase += 2 * math.Pi * currentFrequency * float64(frames) / sampleRate
	phase = math.Mod(phase, 2*math.Pi) // Keep phase bounded
	return data
}

func hannWindow(size int) []float64 {
	window := make([]float64, size)
	for i := range window {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(size-1)))
	}
	return window
}

func findPeaks(data []float64, minHeight float64) []int {
	peaks := []int{}
	for i := 1; i < len(data)-1; i++ {
		if data[i] > data[i-1] && data[i] > data[i+1] && data[i] > minHeight {
			peaks = append(peaks, i)
		}
	}
	return peaks
}

func parabolicInterpolation(f []float64, x int) (float64, float64) {
	xv := 1/2.0*(f[x-1]-f[x+1])/(f[x-1]-2*f[x]+f[x+1]) + float64(x)
	yv := f[x] - 1/4.0*(f[x-1]-f[x+1])*(xv-float64(x))
	return xv, yv
}

func detectFrequency(inData []float32) float64 {
	data := make([]float64, len(inData))
	for i, v := range inData {
		data[i] = float64(v)
	}

	window := hannWindow(len(data))
	for i := range data {
		data[i] *= window[i]
	}

	paddedLength := len(data) * 4
	paddedData := make([]float64, paddedLength)
	copy(paddedData, data)

	fft := fourier.NewFFT(paddedLength)
	coeffs := fft.Coefficients(nil, paddedData)
	magnitude := make([]float64, len(coeffs))
	for i, c := range coeffs {
		magnitude[i] = math.Sqrt(real(c)*real(c) + imag(c)*imag(c))
	}

	maxMag := 0.0
	for _, m := range magnitude {
		if m > maxMag {
			maxMag = m
		}
	}
	peaks := findPeaks(magnitude, maxMag/10)

	if len(peaks) > 0 {
		maxPeak := peaks[0]
		for _, p := range peaks {
			if magnitude[p] > magnitude[maxPeak] {
				maxPeak = p
			}
		}

		trueI, _ := parabolicInterpolation(magnitude, maxPeak)
		peakFrequency := trueI * sampleRate / float64(paddedLength)

		return peakFrequency
	}

	return 0
}

func float32ToBytes(f float32) []byte {
	var buf [4]byte
	*(*float32)(unsafe.Pointer(&buf[0])) = f
	return buf[:]
}

func bytesToFloat32(b []byte) float32 {
	return *(*float32)(unsafe.Pointer(&b[0]))
}

func playbackAndAnalyzeCallback(pOutputSample, pInputSample []byte, framecount uint32) {
	sineWave := generateSineWave(framecount)

	for i, sample := range sineWave {
		copy(pOutputSample[i*4:(i+1)*4], float32ToBytes(sample))
	}

	inputFloat := make([]float32, framecount)
	for i := range inputFloat {
		inputFloat[i] = bytesToFloat32(pInputSample[i*4 : (i+1)*4])
	}

	detectedFreq := detectFrequency(inputFloat)
	calibrationFactor := interpolateCalibration(detectedFreq)
	calibratedFreq := detectedFreq / calibrationFactor

	detectedFreqsMu.Lock()
	detectedFreqs = append(detectedFreqs, calibratedFreq)
	if len(detectedFreqs) > 10 {
		detectedFreqs = detectedFreqs[1:]
	}
	avgFreq := 0.0
	for _, f := range detectedFreqs {
		avgFreq += f
	}
	avgFreq /= float64(len(detectedFreqs))
	detectedFreqsMu.Unlock()

	frequencyMu.RLock()
	currentFrequency := frequency
	frequencyMu.RUnlock()

	errorMargin := 0.05 * currentFrequency
	status := "MISMATCH"
	if math.Abs(avgFreq-currentFrequency) <= errorMargin {
		status = "MATCH"
	}

	volumeMu.RLock()
	currentVolume := volume
	volumeMu.RUnlock()

	fmt.Printf("\rPlayed: %.2f Hz, Detected: %.2f Hz, Status: %s, Volume: %.2f", currentFrequency, avgFreq, status, currentVolume)
}

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

	flagVar(&maxFrequency, "frequency", "f", 1000, "Maximum frequency of the sine wave in Hz")
	flagVar(&amplitude, "amplitude", "a", 0.5, "Amplitude of the sine wave")
	flagVar(&sampleRate, "sample-rate", "r", 48000, "Sample rate in Hz")
	flagVar(&minFrequency, "min-frequency", "m", 0, "Minimum frequency to start sweeping from in Hz")
	flagVar(&sweepRate, "sweep-rate", "s", 1.0, "Frequency change rate in Hz per second")

	flag.Parse()
}

func updateFrequency() {
	frequencyMu.Lock()
	defer frequencyMu.Unlock()

	frequency += float64(sweepDirection) * sweepRate
	if sweepDirection == 1 && frequency >= maxFrequency {
		frequency = maxFrequency
		sweepDirection = -1
	} else if sweepDirection == -1 && frequency <= minFrequency {
		frequency = minFrequency
		sweepDirection = 1
	}
}

func main() {
	initFlags()

	if minFrequency == 0 {
		minFrequency = maxFrequency // If no min frequency is set, use a fixed frequency
	}
	frequency = minFrequency // Start from the minimum frequency

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = channels
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = channels
	deviceConfig.SampleRate = uint32(sampleRate)
	deviceConfig.Alsa.NoMMap = 1

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: playbackAndAnalyzeCallback,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer device.Uninit()

	err = device.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if minFrequency < maxFrequency {
		fmt.Printf("Sweeping frequency from %f Hz to %f Hz. Press Ctrl+C to stop.\n", minFrequency, maxFrequency)
	} else {
		fmt.Printf("Playing and analyzing %f Hz. Press Ctrl+C to stop.\n", frequency)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(10 * time.Millisecond) // Update frequency more frequently
	defer ticker.Stop()

	for shouldRun {
		select {
		case <-sig:
			fmt.Println("\nGot interrupt signal. Stopping...")
			for i := 100; i >= 0; i-- {
				volumeMu.Lock()
				volume = float64(i) / 100.0
				volumeMu.Unlock()
				time.Sleep(10 * time.Millisecond)
			}
			shouldRun = false
		case <-ticker.C:
			if minFrequency < maxFrequency {
				updateFrequency()
			}
		}
	}

	fmt.Println("\nStopping...")
	time.Sleep(time.Second)
	fmt.Println("Playback and analysis stopped.")
}
