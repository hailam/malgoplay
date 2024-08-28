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
	sampleRate      = 48000
	channels        = uint32(1)
	frequency       float64
	amplitude       = 0.5
	phase           float64
	detectedFreqs   = make([]float64, 0, 10)
	detectedFreqsMu sync.Mutex
)

// Calibration data
// use this to adjust the calibration of the frequency detection based on the microphone used
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
	for i := range data {
		t := float64(i) / float64(sampleRate)
		data[i] = float32(amplitude * math.Sin(2*math.Pi*frequency*t+phase))
	}
	phase += 2 * math.Pi * frequency * float64(frames) / float64(sampleRate)
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
	// Convert input to float64
	data := make([]float64, len(inData))
	for i, v := range inData {
		data[i] = float64(v)
	}

	// Apply Hann window
	window := hannWindow(len(data))
	for i := range data {
		data[i] *= window[i]
	}

	// Pad data
	paddedLength := len(data) * 3
	paddedData := make([]float64, paddedLength)
	copy(paddedData, data)

	// Perform FFT
	fft := fourier.NewFFT(paddedLength)
	coeffs := fft.Coefficients(nil, paddedData)
	magnitude := make([]float64, len(coeffs))
	for i, c := range coeffs {
		magnitude[i] = math.Sqrt(real(c)*real(c) + imag(c)*imag(c))
	}

	// Find peaks
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

		// Parabolic interpolation
		trueI, _ := parabolicInterpolation(magnitude, maxPeak)
		peakFrequency := trueI * float64(sampleRate) / float64(paddedLength)

		return peakFrequency
	}

	return 0 // No peak found
}

// float32ToBytes converts a float32 to a byte slice
func float32ToBytes(f float32) []byte {
	var buf [4]byte
	*(*float32)(unsafe.Pointer(&buf[0])) = f
	return buf[:]
}

// bytesToFloat32 converts a byte slice to a float32
func bytesToFloat32(b []byte) float32 {
	return *(*float32)(unsafe.Pointer(&b[0]))
}

func playbackAndAnalyzeCallback(pOutputSample, pInputSample []byte, framecount uint32) {
	// Generate sine wave
	sineWave := generateSineWave(framecount)

	// Copy sine wave to output
	for i, sample := range sineWave {
		copy(pOutputSample[i*4:(i+1)*4], float32ToBytes(sample))
	}

	// Analyze input
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

	errorMargin := 0.05 * frequency
	status := "MISMATCH"
	if math.Abs(avgFreq-frequency) <= errorMargin {
		status = "MATCH"
	}

	fmt.Printf("\rPlayed: %.2f Hz, Detected: %.2f Hz, Status: %s", frequency, avgFreq, status)
}

func main() {
	// Parse command-line arguments
	flag.Float64Var(&frequency, "f", 440, "Frequency of the sine wave in Hz")
	flag.Parse()

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

	fmt.Printf("Playing and analyzing %f Hz. Press Ctrl+C to stop.\n", frequency)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	fmt.Println("\nStopping...")
	time.Sleep(time.Second)
	fmt.Println("Playback and analysis stopped.")
}
