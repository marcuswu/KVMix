package viewmodel

import (
	"math"

	"github.com/marcuswu/KVMix/channel"
	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/gosmartknob/pb"
)

type VolumeAppViewModel struct {
	position      int32
	positionNonce uint32
	pressNonce    uint32
	appFactory    channel.ChannelFactory
	channels      []channel.Channel
	setNonce      bool
}

func NewVolumeAppViewModel(apps config.VolumeAppList, factory channel.ChannelFactory) *VolumeAppViewModel {
	channels, _ := factory.ChannelsMatching(apps.ToMatchMap())
	return &VolumeAppViewModel{
		position:      0,
		positionNonce: 0,
		pressNonce:    0,
		appFactory:    factory,
		channels:      channels,
		setNonce:      false,
	}
}

func (vavm *VolumeAppViewModel) getNonceSet() bool {
	return vavm.setNonce
}

func (vavm *VolumeAppViewModel) setNonceSet(isSet bool) {
	vavm.setNonce = isSet
}

func (vavm *VolumeAppViewModel) getPositionNonce() uint32 {
	return vavm.positionNonce
}

func (vavm *VolumeAppViewModel) setPositionNonce(nonce uint32) {
	vavm.positionNonce = nonce
}

func (vavm *VolumeAppViewModel) getPosition() int32 {
	return vavm.position
}

func (vavm *VolumeAppViewModel) setPosition(pos int32) {
	vavm.position = pos
}

func (vavm *VolumeAppViewModel) getPressNonce() uint32 {
	return vavm.pressNonce
}

func (vavm *VolumeAppViewModel) setPressNonce(nonce uint32) {
	vavm.pressNonce = nonce
}

func (vavm *VolumeAppViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Check for select of App Channel or Back
	// Return new ViewModel accordingly
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	ret.RegenConfig = handleNonces(vavm, state)
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

func (vavm *VolumeAppViewModel) Restore(state *pb.SmartKnobState) {
	restorePosition(vavm, state)
}

func (vavm *VolumeAppViewModel) GenerateConfig() *pb.SmartKnobConfig {
	title := "Back"
	if vavm.position > 0 {
		title = vavm.channels[vavm.position-1].Name()
	}
	return &pb.SmartKnobConfig{
		Position:             vavm.position,
		PositionNonce:        vavm.positionNonce,
		MinPosition:          0,
		MaxPosition:          int32(len(vavm.channels)),
		PositionWidthRadians: (30 / 180.0) * math.Pi,
		EndstopStrengthUnit:  1,
		DetentStrengthUnit:   0.4,
		SnapPoint:            0.7,
		Text:                 title,
	}
}
