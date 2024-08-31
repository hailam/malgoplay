package fsg

import (
	"encoding/binary"
	"math"
	"sync"

	"github.com/gen2brain/malgo"
)

type MockDevice struct {
	mutex           sync.Mutex
	isStarted       bool
	dataCallback    malgo.DataProc
	sampleRate      uint32
	channels        uint32
	capturedSamples []float32
}

func NewMockDevice(sampleRate, channels uint32) *MockDevice {
	return &MockDevice{
		sampleRate: sampleRate,
		channels:   channels,
	}
}

func (m *MockDevice) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.dataCallback == nil {
		// Set the callback to a default no-op if it's not already set
		m.dataCallback = func(pOutputSample, pInputSamples []byte, framecount uint32) {
			// Default no-op callback
		}
	}

	m.isStarted = true
	println("Mock device started.")
	return nil
}

func (m *MockDevice) SetCallback(dataCallback malgo.DataProc) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.dataCallback = dataCallback
	println("MockDevice: Callback set.")

	if m.isStarted {
		// If the device is already started, ensure callback is set
		println("Callback set for running mock device.")
	}
}

func (m *MockDevice) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.isStarted = false
	return nil
}

func (m *MockDevice) Uninit() error {
	return nil
}

func (m *MockDevice) GenerateSamples(frameCount uint32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.dataCallback == nil || !m.isStarted {
		println("MockDevice: No callback or device not started.")
		return
	}

	byteCount := frameCount * m.channels * 4 // 4 bytes per float32
	output := make([]byte, byteCount)

	println("MockDevice generating samples.")
	m.dataCallback(output, nil, frameCount)

	m.capturedSamples = make([]float32, frameCount*m.channels)
	for i := uint32(0); i < frameCount*m.channels; i++ {
		m.capturedSamples[i] = math.Float32frombits(binary.LittleEndian.Uint32(output[i*4 : (i+1)*4]))
	}

	if len(m.capturedSamples) == 0 {
		println("MockDevice: No samples captured.")
	} else {
		println("MockDevice: Samples captured.")
	}
}

func (m *MockDevice) GetCapturedSamples() []float32 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.capturedSamples
}

// Helper function to convert 4 bytes to float32
func float32frombytes(bytes []byte) float32 {
	bits := uint32(bytes[0]) | uint32(bytes[1])<<8 | uint32(bytes[2])<<16 | uint32(bytes[3])<<24
	return math.Float32frombits(bits)
}
