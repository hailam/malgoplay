package fsg

import "github.com/gen2brain/malgo"

type MalgoDeviceWrapper struct {
	*malgo.Device
}

func (w *MalgoDeviceWrapper) Uninit() error {
	w.Device.Uninit()
	return nil
}
