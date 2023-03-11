package ddc

import (
	"unsafe"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/windows"
)

const (
	MOUSE_X = 1000
	MOUSE_Y = 1000
	M_POINT = (MOUSE_X & 0xFFFFFFFF) | (MOUSE_Y << 32)
)

type DDCControl struct {
	user32               *windows.LazyDLL
	dxva2                *windows.LazyDLL
	enumDisplayDevices   *windows.LazyProc
	enumDisplayMonitors  *windows.LazyProc
	getMonitorInfo       *windows.LazyProc
	monitorsFromHMONITOR *windows.LazyProc
	setVCPFeature        *windows.LazyProc
	destroyMonitor       *windows.LazyProc
}

func NewDDC() DDC {
	user32 := windows.NewLazySystemDLL("user32.dll")
	dxva2 := windows.NewLazySystemDLL("dxva2.dll")

	enumDisplayDevices := user32.NewProc("EnumDisplayDevicesW")
	enumDisplayMonitors := user32.NewProc("EnumDisplayMonitors")
	getMonitorInfo := user32.NewProc("GetMonitorInfoW")
	monitorsFromHMONITOR := dxva2.NewProc("GetPhysicalMonitorsFromHMONITOR")
	setVCPFeature := dxva2.NewProc("SetVCPFeature")
	destroyMonitor := dxva2.NewProc("DestroyPhysicalMonitor")

	return &DDCControl{
		user32:               user32,
		dxva2:                dxva2,
		enumDisplayDevices:   enumDisplayDevices,
		enumDisplayMonitors:  enumDisplayMonitors,
		getMonitorInfo:       getMonitorInfo,
		monitorsFromHMONITOR: monitorsFromHMONITOR,
		setVCPFeature:        setVCPFeature,
		destroyMonitor:       destroyMonitor,
	}
}

type displayDeviceW struct {
	cb           int32
	DeviceName   [32]byte
	DeviceString [128]byte
	StateFlags   int32
	DeviceID     [128]byte
	DeviceKey    [128]byte
}

type monitorData struct {
	monitor    uintptr
	deviceName uintptr
}

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type monitorInfo struct {
	cbSize    int32
	rcMonitor rect
	rcWork    rect
	dwFlags   int32
	szDevice  [64]byte
}

func (d *DDCControl) monitorHandleFromDeviceName(deviceName string) (result uintptr, err error) {
	monitor := monitorData{}
	monitorName, _ := windows.UTF16PtrFromString(deviceName)
	monitor.deviceName = uintptr(unsafe.Pointer(monitorName))
	callback := windows.NewCallback(func(hMonitor uintptr, hDC uintptr, rect uintptr, data uintptr) uintptr {
		monitor := *(**monitorData)(unsafe.Pointer(&data))
		monitorName := windows.UTF16PtrToString(*(**uint16)(unsafe.Pointer(&monitor.deviceName)))
		mInfo := monitorInfo{}
		mInfo.cbSize = int32(unsafe.Sizeof(mInfo))
		blah, _, err := d.getMonitorInfo.Call(hMonitor, uintptr(unsafe.Pointer(&mInfo)))
		log.Debug().Int("result", int(blah)).Err(err).Msg("Getting Monitor Info")
		monitorInfoName := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(&mInfo.szDevice)))
		log.Debug().Str("name", monitorInfoName).Msg("Checking Monitor")
		if monitorInfoName == monitorName {
			log.Debug().Str("name", monitorName).Msg("Found Monitor")
			monitor.monitor = hMonitor
		}

		return 0
	})
	result, _, err = d.enumDisplayMonitors.Call(uintptr(0), uintptr(0), callback, uintptr(unsafe.Pointer(&monitor)))

	if result != 0 || monitor.monitor == 0 {
		result = uintptr(0)
	} else {
		result = monitor.monitor
	}

	return
}

func (d *DDCControl) monitorFromIndex(index int) (result uintptr, err error) {
	dd := displayDeviceW{}
	dd.cb = int32(unsafe.Sizeof(dd))
	result, _, err = d.enumDisplayDevices.Call(0, uintptr(index), uintptr(unsafe.Pointer(&dd)), 0)
	if result != 0 /*&& (dd.StateFlags&DISPLAY_DEVICE_ATTACHED_TO_DESKTOP) != 0*/ {
		deviceName := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(&dd.DeviceName)))
		log.Debug().Str("name", deviceName).Msg("Looking for Monitor")
		return d.monitorHandleFromDeviceName(deviceName)
	}
	return
}

func (d *DDCControl) getPhysicalMonitor(handle uintptr) (result uintptr, err error) {
	b := make([]byte, 256)
	_, _, callErr := d.monitorsFromHMONITOR.Call(handle, uintptr(1), uintptr(unsafe.Pointer(&b[0])))

	result = uintptr(b[0])
	err = callErr
	return
}

func (d *DDCControl) SendCommand(index int, command int, value int) error {
	hMon, callErr := d.monitorFromIndex(index)
	if callErr != windows.NOERROR {
		return callErr
	}
	mHandle, callErr := d.getPhysicalMonitor(hMon)
	if callErr != windows.NOERROR {
		return callErr
	}
	defer d.destroyPhysicalMonitor(mHandle)

	_, _, callErr = d.setVCPFeature.Call(mHandle, uintptr(command), uintptr(value))

	return callErr
}

func (d *DDCControl) destroyPhysicalMonitor(monitor uintptr) error {
	_, _, callErr := d.destroyMonitor.Call(monitor, 0, 0)

	return callErr
}

func (d *DDCControl) SetInputSource(index int, source InputSource) error {
	log.Debug().Int("monitor index", index).Int("source", int(source)).Msg("Setting Input Source on Windows")
	err := d.SendCommand(index, SELECT_SOURCE, int(source))
	if err != nil {
		log.Error().Msg(err.Error())
	}
	return err
}
