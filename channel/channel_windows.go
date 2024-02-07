package channel

import (
	"errors"
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/mitchellh/go-ps"
	"github.com/moutend/go-wca/pkg/wca"
)

type WindowsChannel struct {
	name        string
	pid         uint32
	processName string

	control *wca.IAudioSessionControl2
	volume  *wca.ISimpleAudioVolume

	eventCtx *ole.GUID
}

func newWindowsChannel(
	control *wca.IAudioSessionControl2,
	volume *wca.ISimpleAudioVolume,
	pid uint32,
	eventCtx *ole.GUID,
) (*WindowsChannel, error) {

	ch := &WindowsChannel{
		control:  control,
		volume:   volume,
		pid:      pid,
		eventCtx: eventCtx,
	}

	if pid == 0 {
		ch.name = "main"
		return ch, nil
	}

	process, err := ps.FindProcess(int(pid))
	if err != nil {
		defer ch.Release()

		return nil, fmt.Errorf("find process name by pid: %w", err)
	}

	if process == nil {
		return nil, fmt.Errorf("cound not find PID %d: %w", pid, err)
	}

	ch.processName = process.Executable()
	ch.name = ch.processName

	return ch, nil
}

func (ch *WindowsChannel) Name() string {
	return ch.name
}

func (ch *WindowsChannel) GetVolume() float64 {
	var level float32

	err := ch.volume.GetMasterVolume(&level)
	if err != nil {
		return 0
	}

	return float64(level)
}

func (ch *WindowsChannel) SetVolume(v float64) error {
	err := ch.volume.SetMasterVolume(float32(v), ch.eventCtx)
	if err != nil {
		return fmt.Errorf("failed to adjust session volume: %w", err)
	}

	var state uint32

	err = ch.control.GetState(&state)
	if err != nil {
		return fmt.Errorf("failed to get session state: %w", err)
	}

	if state == wca.AudioSessionStateExpired {
		return errors.New("audio session expired")
	}

	return nil
}

func (ch *WindowsChannel) Release() {
	ch.volume.Release()
	ch.control.Release()
}

func (ch *WindowsChannel) String() string {
	return fmt.Sprintf("%s: %f", ch.name, ch.GetVolume())
}

type WindowsMainChannel struct {
	volume *wca.IAudioEndpointVolume

	eventCtx *ole.GUID
}

func newWindowsMainChannel(volume *wca.IAudioEndpointVolume, eventCtx *ole.GUID) (*WindowsMainChannel, error) {
	return &WindowsMainChannel{
		volume:   volume,
		eventCtx: eventCtx,
	}, nil
}

func (ch *WindowsMainChannel) Name() string {
	return "Main Volume"
}

func (ch *WindowsMainChannel) GetVolume() float64 {
	var volume float32

	err := ch.volume.GetMasterVolumeLevelScalar(&volume)
	if err != nil {
		return 0
	}

	return float64(volume)
}

func (ch *WindowsMainChannel) SetVolume(v float64) error {
	err := ch.volume.SetMasterVolumeLevelScalar(float32(v), ch.eventCtx)
	if err != nil {
		return fmt.Errorf("failed to adjust session volume: %w", err)
	}

	return nil
}

func (ch *WindowsMainChannel) Release() {
	ch.volume.Release()
}

func (ch *WindowsMainChannel) String() string {
	return fmt.Sprintf("Main Channel: %f", ch.GetVolume())
}
