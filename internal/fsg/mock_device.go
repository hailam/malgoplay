package fsg

import (
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
	m.isStarted = true
	return nil
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

func (m *MockDevice) SetDataCallback(dataCallback malgo.DataProc) {
	m.dataCallback = dataCallback
}

func (m *MockDevice) GenerateSamples(frameCount uint32) {
	if m.dataCallback == nil {
		return
	}

	byteCount := frameCount * m.channels * 4 // 4 bytes per float32
	output := make([]byte, byteCount)
	m.dataCallback(output, nil, frameCount)

	// Convert bytes back to float32 for easier inspection
	m.capturedSamples = make([]float32, frameCount*m.channels)
	for i := uint32(0); i < frameCount*m.channels; i++ {
		m.capturedSamples[i] = float32frombytes(output[i*4 : (i+1)*4])
	}
}

func (m *MockDevice) GetCapturedSamples() []float32 {
	return m.capturedSamples
}

// Helper function to convert 4 bytes to float32
func float32frombytes(bytes []byte) float32 {
	bits := uint32(bytes[0]) | uint32(bytes[1])<<8 | uint32(bytes[2])<<16 | uint32(bytes[3])<<24
	return math.Float32frombits(bits)
}
