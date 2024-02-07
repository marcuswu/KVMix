package viewmodel

import (
	"fmt"

	"github.com/marcuswu/KVMix/channel"
	"github.com/marcuswu/gosmartknob/pb"
)

type VolumeViewModel struct {
	position   int32
	pressNonce uint32
	appChannel channel.Channel
}

func NewVolumeViewModel(pressNonce uint32, appChannel channel.Channel) *VolumeViewModel {
	return &VolumeViewModel{
		position:   0,
		pressNonce: pressNonce,
		appChannel: appChannel,
	}
}

func (vvm *VolumeViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Change volume of app, setting the volume as we go
	// Press sends us back one ViewModel
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	if state.CurrentPosition != vvm.position {
		vvm.position = state.CurrentPosition
		vvm.appChannel.SetVolume(float64(vvm.position) / 100.0)
	}
	if state.PressNonce != vvm.pressNonce {
		ret.Navigation = NavBack
		ret.RegenConfig = true
		ret.pressNonce = state.PressNonce
	}
	return ret
}

func (vvm *VolumeViewModel) GenerateConfig() *pb.SmartKnobConfig {
	title := fmt.Sprintf("Volume for %s", vvm.appChannel.Name())
	return &pb.SmartKnobConfig{
		Position:            vvm.position,
		MinPosition:         0,
		MaxPosition:         100,
		DetentStrengthUnit:  0.0,
		EndstopStrengthUnit: 1.0,
		SnapPoint:           0.7,
		Text:                title,
	}
}
