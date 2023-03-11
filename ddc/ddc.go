package ddc

type InputSource int

const (
	VGA1 InputSource = iota + 1
	VGA2
	DVI1
	DVI2
	COMPOSITE_VIDEO1
	COMPOSITE_VIDEO2
	SVIDEO1
	SVIDEO2
	TUNER1
	TUNER2
	TUNER3
	COMPONENT_VIDEO1
	COMPONENT_VIDEO2
	COMPONENT_VIDEO3
	DISPLAY_PORT1
	DISPLAY_PORT2
	HDMI1
	HDMI2
)

const DISPLAY_DEVICE_ATTACHED_TO_DESKTOP int32 = 0x01

const SELECT_SOURCE = '\x60'

type DDC interface {
	SendCommand(int, int, int) error
	SetInputSource(int, InputSource) error
}
