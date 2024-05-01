package channel

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca/pkg/wca"
)

const (
	GUID                  = "{374cac77-3ffc-4010-9ac4-88ebd23e3e00}"
	deviceChangeThreshold = 200 * time.Millisecond
	NoCurrentProcess      = 0x889000D
)

var ErrNoProcess = errors.New("no process")

type WindowsChannelFactory struct {
	eventCtx *ole.GUID
}

func New() ChannelFactory {
	return WindowsChannelFactory{eventCtx: ole.NewGUID(GUID)}
}

func (wcf WindowsChannelFactory) ChannelsMatching(ids map[string]interface{}) ([]Channel, error) {
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		const oleFalse = 1
		oleError := &ole.OleError{}

		errOk := errors.As(err, &oleError)
		if errOk && oleError.Code() != oleFalse {
			return []Channel{}, fmt.Errorf("redundant call to CoInitializeEx: %w", err)
		}

		if !errOk {
			return []Channel{}, fmt.Errorf("redundant call to CoInitializeEx: %w", err)
		}
	}
	defer ole.CoUninitialize()

	var deviceEnumerator *wca.IMMDeviceEnumerator
	err := wca.CoCreateInstance(
		wca.CLSID_MMDeviceEnumerator,
		0,
		wca.CLSCTX_ALL,
		wca.IID_IMMDeviceEnumerator,
		&deviceEnumerator,
	)
	if err != nil {
		return []Channel{}, err
	}
	defer deviceEnumerator.Release()

	var defaultOutputEndpoint *wca.IMMDevice
	err = deviceEnumerator.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &defaultOutputEndpoint)
	if err != nil {
		return []Channel{}, fmt.Errorf("failed to get default audio endpoints: %w", err)
	}
	defer defaultOutputEndpoint.Release()

	mainOut, err := wcf.getMainChannel(defaultOutputEndpoint)
	if err != nil {
		return []Channel{}, fmt.Errorf("failed to get master audio output channel: %w", err)
	}

	appChannels, err := wcf.getChannels(ids, defaultOutputEndpoint)
	channels := make([]Channel, 0, len(appChannels)+1)
	if err != nil {
		return []Channel{}, fmt.Errorf("enumerate device sessions: %w", err)
	}
	channels = append(channels, mainOut)
	channels = append(channels, appChannels...)

	return channels, nil
}

func (wcf WindowsChannelFactory) getMainChannel(device *wca.IMMDevice) (*WindowsMainChannel, error) {

	var volume *wca.IAudioEndpointVolume

	err := device.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &volume)
	if err != nil {
		return nil, fmt.Errorf("failed to activate master channel: %w", err)
	}

	main, err := newWindowsMainChannel(volume, wcf.eventCtx)
	if err != nil {
		return nil, fmt.Errorf("create master session: %w", err)
	}

	return main, nil
}

func (wcf WindowsChannelFactory) getChannels(ids map[string]interface{}, device *wca.IMMDevice) ([]Channel, error) {
	channels := make([]Channel, 0)

	dispatch, err := device.QueryInterface(wca.IID_IMMEndpoint)
	if err != nil {
		return channels, err
	}

	endpointType := (*wca.IMMEndpoint)(unsafe.Pointer(dispatch))
	defer endpointType.Release()

	var dataFlow uint32
	if err := endpointType.GetDataFlow(&dataFlow); err != nil {
		return channels, err
	}

	if dataFlow == wca.ERender {
		deviceChannels, err := wcf.enumerateDeviceChannels(device, ids)
		if err != nil {
			return channels, err
		}
		return deviceChannels, nil
	}

	return channels, nil
}

func (wcf WindowsChannelFactory) enumerateDeviceChannels(
	endpoint *wca.IMMDevice,
	ids map[string]interface{},
) ([]Channel, error) {
	var audioSessionManager2 *wca.IAudioSessionManager2

	err := endpoint.Activate(
		wca.IID_IAudioSessionManager2,
		wca.CLSCTX_ALL,
		nil,
		&audioSessionManager2,
	)
	if err != nil {
		return []Channel{}, fmt.Errorf("activate endpoint: %w", err)
	}
	defer audioSessionManager2.Release()

	var sessionEnumerator *wca.IAudioSessionEnumerator

	if err := audioSessionManager2.GetSessionEnumerator(&sessionEnumerator); err != nil {
		return []Channel{}, err
	}
	defer sessionEnumerator.Release()

	var channelCount int

	if err := sessionEnumerator.GetCount(&channelCount); err != nil {
		return []Channel{}, fmt.Errorf("get session count: %w", err)
	}

	channels := make([]Channel, 0, channelCount)
	for channelIdx := 0; channelIdx < channelCount; channelIdx++ {

		var audioSessionControl *wca.IAudioSessionControl
		if err := sessionEnumerator.GetSession(channelIdx, &audioSessionControl); err != nil {
			return []Channel{}, fmt.Errorf("get channel %d from enumerator: %w", channelIdx, err)
		}

		dispatch, err := audioSessionControl.QueryInterface(wca.IID_IAudioSessionControl2)
		if err != nil {
			return []Channel{}, fmt.Errorf("query channel %d IAudioSessionControl2: %w", channelIdx, err)
		}

		audioSessionControl.Release()

		audioSessionControl2 := (*wca.IAudioSessionControl2)(unsafe.Pointer(dispatch))

		var pid uint32

		if err := audioSessionControl2.GetProcessId(&pid); err != nil {

			isSystemSoundsErr := audioSessionControl2.IsSystemSoundsSession()
			if isSystemSoundsErr != nil && !strings.Contains(err.Error(), fmt.Sprintf("%d", NoCurrentProcess)) {
				continue
			}
		}

		dispatch, err = audioSessionControl2.QueryInterface(wca.IID_ISimpleAudioVolume)
		if err != nil {
			continue
		}

		simpleAudioVolume := (*wca.ISimpleAudioVolume)(unsafe.Pointer(dispatch))

		newChannel, err := newWindowsChannel(audioSessionControl2, simpleAudioVolume, pid, wcf.eventCtx)
		if err != nil {

			if errors.Is(err, ErrNoProcess) {
				continue
			}

			audioSessionControl2.Release()
			simpleAudioVolume.Release()

			continue
		}

		// add it to our slice
		_, nameOk := ids[newChannel.name]
		_, pnameOk := ids[newChannel.processName]
		if nameOk || pnameOk {
			channels = append(channels, newChannel)
			delete(ids, newChannel.name)
			delete(ids, newChannel.processName)
		}
	}

	return channels, nil
}
