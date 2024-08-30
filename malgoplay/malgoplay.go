package malgoplay

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"unsafe"

	"github.com/gen2brain/malgo"
	"gonum.org/v1/gonum/dsp/fourier"
)

var (
	SampleRate      = float64(48000)
	Channels        = uint32(1)
	Frequency       float64
	MinFrequency    float64
	MaxFrequency    float64
	Amplitude       = 0.5
	Phase           float64
	DetectedFreqs   = make([]float64, 0, 10)
	DetectedFreqsMu sync.Mutex
	SweepDirection  = 1   // 1 for increasing, -1 for decreasing
	SweepRate       = 1.0 // Hz per second
	FrequencyMu     sync.RWMutex
	VolumeMu        sync.RWMutex
	Volume          = 1.0
	Calibration     = map[float64]float64{
		20:    1.2,
		100:   1.0,
		300:   1.0,
		1000:  1.1,
		5000:  1.0,
		10000: 1.0,
		15000: 1.0,
		20000: 1.0,
	}
)

func interpolateCalibration(freq float64) float64 {
	keys := make([]float64, 0, len(Calibration))
	for k := range Calibration {
		keys = append(keys, k)
	}
	sort.Float64s(keys)

	if freq <= keys[0] {
		return Calibration[keys[0]]
	}
	if freq >= keys[len(keys)-1] {
		return Calibration[keys[len(keys)-1]]
	}

	for i := 0; i < len(keys)-1; i++ {
		if keys[i] <= freq && freq < keys[i+1] {
			t := (freq - keys[i]) / (keys[i+1] - keys[i])
			return Calibration[keys[i]]*(1-t) + Calibration[keys[i+1]]*t
		}
	}
	return 1.0
}

func GenerateSineWave(frames uint32) []float32 {
	data := make([]float32, frames)
	FrequencyMu.RLock()
	currentFrequency := Frequency
	FrequencyMu.RUnlock()
	VolumeMu.RLock()
	currentVolume := Volume
	VolumeMu.RUnlock()
	for i := range data {
		t := float64(i) / SampleRate
		data[i] = float32(Amplitude * currentVolume * math.Sin(2*math.Pi*currentFrequency*t+Phase))
	}
	Phase += 2 * math.Pi * currentFrequency * float64(frames) / SampleRate
	Phase = math.Mod(Phase, 2*math.Pi) // Keep Phase bounded
	return data
}

func HannWindow(size int) []float64 {
	window := make([]float64, size)
	for i := range window {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(size-1)))
	}
	return window
}

func FindPeaks(data []float64, minHeight float64) []int {
	peaks := []int{}
	for i := 1; i < len(data)-1; i++ {
		if data[i] > data[i-1] && data[i] > data[i+1] && data[i] > minHeight {
			peaks = append(peaks, i)
		}
	}
	return peaks
}

func ParabolicInterpolation(f []float64, x int) (float64, float64) {
	xv := 1/2.0*(f[x-1]-f[x+1])/(f[x-1]-2*f[x]+f[x+1]) + float64(x)
	yv := f[x] - 1/4.0*(f[x-1]-f[x+1])*(xv-float64(x))
	return xv, yv
}

func DetectFrequency(inData []float32) float64 {
	data := make([]float64, len(inData))
	for i, v := range inData {
		data[i] = float64(v)
	}

	window := HannWindow(len(data))
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
	peaks := FindPeaks(magnitude, maxMag/10)

	if len(peaks) > 0 {
		maxPeak := peaks[0]
		for _, p := range peaks {
			if magnitude[p] > magnitude[maxPeak] {
				maxPeak = p
			}
		}

		trueI, _ := ParabolicInterpolation(magnitude, maxPeak)
		peakFrequency := trueI * SampleRate / float64(paddedLength)

		return peakFrequency
	}

	return 0
}

func Float32ToBytes(f float32) []byte {
	var buf [4]byte
	*(*float32)(unsafe.Pointer(&buf[0])) = f
	return buf[:]
}

func BytesToFloat32(b []byte) float32 {
	return *(*float32)(unsafe.Pointer(&b[0]))
}

func PlaybackAndAnalyzeCallback(pOutputSample, pInputSample []byte, framecount uint32) {
	sineWave := GenerateSineWave(framecount)

	for i, sample := range sineWave {
		copy(pOutputSample[i*4:(i+1)*4], Float32ToBytes(sample))
	}

	inputFloat := make([]float32, framecount)
	for i := range inputFloat {
		inputFloat[i] = BytesToFloat32(pInputSample[i*4 : (i+1)*4])
	}

	detectedFreq := DetectFrequency(inputFloat)
	CalibrationFactor := interpolateCalibration(detectedFreq)
	calibratedFreq := detectedFreq / CalibrationFactor

	DetectedFreqsMu.Lock()
	DetectedFreqs = append(DetectedFreqs, calibratedFreq)
	if len(DetectedFreqs) > 10 {
		DetectedFreqs = DetectedFreqs[1:]
	}
	avgFreq := 0.0
	for _, f := range DetectedFreqs {
		avgFreq += f
	}
	avgFreq /= float64(len(DetectedFreqs))
	DetectedFreqsMu.Unlock()

	FrequencyMu.RLock()
	currentFrequency := Frequency
	FrequencyMu.RUnlock()

	errorMargin := 0.05 * currentFrequency
	status := "MISMATCH"
	if math.Abs(avgFreq-currentFrequency) <= errorMargin {
		status = "MATCH"
	}

	VolumeMu.RLock()
	currentVolume := Volume
	VolumeMu.RUnlock()

	fmt.Printf("\rPlayed: %.2f Hz, Detected: %.2f Hz, Status: %s, Volume: %.2f", currentFrequency, avgFreq, status, currentVolume)
}

func UpdateFrequency() {
	FrequencyMu.Lock()
	defer FrequencyMu.Unlock()

	Frequency += float64(SweepDirection) * SweepRate
	if SweepDirection == 1 && Frequency >= MaxFrequency {
		Frequency = MaxFrequency
		SweepDirection = -1
	} else if SweepDirection == -1 && Frequency <= MinFrequency {
		Frequency = MinFrequency
		SweepDirection = 1
	}
}

func SetFrequency(freq float64) {
	FrequencyMu.Lock()
	defer FrequencyMu.Unlock()
	Frequency = freq
}

func SetVolume(vol float64) {
	VolumeMu.Lock()
	defer VolumeMu.Unlock()
	Volume = vol
}

func GetFrequency() float64 {
	FrequencyMu.RLock()
	defer FrequencyMu.RUnlock()
	return Frequency
}

func GetVolume() float64 {
	VolumeMu.RLock()
	defer VolumeMu.RUnlock()
	return Volume
}

func InitDevice() (*malgo.Device, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Duplex)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = Channels
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = Channels
	deviceConfig.SampleRate = uint32(SampleRate)
	deviceConfig.Alsa.NoMMap = 1

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: PlaybackAndAnalyzeCallback,
	})
	if err != nil {
		return nil, err
	}

	return device, nil
}

func CleanupDevice(device *malgo.Device) {
	device.Uninit()
}
