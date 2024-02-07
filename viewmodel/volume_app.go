package viewmodel

import (
	"github.com/marcuswu/KVMix/channel"
	"github.com/marcuswu/gosmartknob/pb"
)

type VolumeAppViewModel struct {
	position   int32
	pressNonce uint32
	appFactory channel.ChannelFactory
	channels   []channel.Channel
}

func NewVolumeAppViewModel(pressNonce uint32, factory channel.ChannelFactory) *VolumeAppViewModel {
	channels, _ := factory.Channels()
	return &VolumeAppViewModel{
		position:   0,
		pressNonce: pressNonce,
		appFactory: factory,
		channels:   channels,
	}
}

func (vavm *VolumeAppViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Check for select of App Channel or Back
	// Return new ViewModel accordingly
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	if state.CurrentPosition != vavm.position {
		vavm.position = state.CurrentPosition
		ret.RegenConfig = true
	}
	if state.PressNonce != vavm.pressNonce {
		ret.Navigation = NavBack
		ret.pressNonce = state.PressNonce
		ret.RegenConfig = true
		if vavm.position > 0 {
			// Construct VolumeViewModel using vavm.channels[vavm.position]
			ret.Navigation = NavTo
			ret.ViewModel = NewVolumeViewModel(ret.pressNonce, vavm.channels[vavm.position-1])
		}
	}
	return ret
}

func (vavm *VolumeAppViewModel) GenerateConfig() *pb.SmartKnobConfig {
	title := "Back"
	if vavm.position > 0 {
		title = vavm.channels[vavm.position-1].Name()
	}
	return &pb.SmartKnobConfig{
		Position:           vavm.position,
		MinPosition:        0,
		MaxPosition:        int32(len(vavm.channels)),
		DetentStrengthUnit: 0.4,
		SnapPoint:          0.7,
		Text:               title,
	}
}
