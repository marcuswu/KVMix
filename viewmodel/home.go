package viewmodel

import (
	"math"

	"github.com/marcuswu/KVMix/channel"
	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/KVMix/ddc"
	"github.com/marcuswu/gosmartknob/pb"
)

type HomePositions int

const (
	ChangeVolumePosition HomePositions = iota
	ChangeComputerPosition
)

type HomeViewModel struct {
	position      int32
	positionNonce uint32
	positionNames []string
	pressNonce    uint32
	config        config.Config
	ddc           ddc.DDC
	setNonce      bool
}

func NewHomeViewModel(config config.Config, ddc ddc.DDC) *HomeViewModel {
	return &HomeViewModel{
		position:      0,
		positionNonce: 0,
		positionNames: []string{"Change Volume", "Change Computer"},
		pressNonce:    0,
		config:        config,
		ddc:           ddc,
		setNonce:      false,
	}
}

func (hvm *HomeViewModel) getNonceSet() bool {
	return hvm.setNonce
}

func (hvm *HomeViewModel) setNonceSet(isSet bool) {
	hvm.setNonce = isSet
}

func (hvm *HomeViewModel) getPositionNonce() uint32 {
	return hvm.positionNonce
}

func (hvm *HomeViewModel) setPositionNonce(nonce uint32) {
	hvm.positionNonce = nonce
}

func (hvm *HomeViewModel) getPosition() int32 {
	return hvm.position
}

func (hvm *HomeViewModel) setPosition(pos int32) {
	if len(hvm.positionNames) <= int(pos) {
		hvm.position = 0
	}
	hvm.position = pos
}

func (hvm *HomeViewModel) getPressNonce() uint32 {
	return hvm.pressNonce
}

func (hvm *HomeViewModel) setPressNonce(nonce uint32) {
	hvm.pressNonce = nonce
}

func (hvm *HomeViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Check for select of ChangeVolume or ChangeComputer
	// Return new ViewModel accordingly
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	handleNonces(hvm, state)
	// The knob position has updated
	if state.CurrentPosition != hvm.getPosition() {
		hvm.setPosition(state.CurrentPosition)
		ret.RegenConfig = true
	}
	if state.PressNonce != hvm.pressNonce {
		hvm.pressNonce = state.PressNonce
		switch hvm.position {
		case int32(ChangeVolumePosition):
			ret.ViewModel = NewVolumeAppViewModel(hvm.config.VolumeApps, channel.New())
		case int32(ChangeComputerPosition):
			ret.ViewModel = NewComputersViewModel(hvm.config.Computers, hvm.ddc)
		default:
			// Shouldn't land here, but if we do, do nothing
			return ret
		}
		ret.Navigation = NavTo
		ret.RegenConfig = true
	}
	return ret
}

func (hvm *HomeViewModel) Restore(state *pb.SmartKnobState) {
	restorePosition(hvm, state)
}

func (hvm *HomeViewModel) GenerateConfig() *pb.SmartKnobConfig {
	return &pb.SmartKnobConfig{
		MinPosition:          0,
		MaxPosition:          int32(len(hvm.positionNames)) - 1,
		PositionWidthRadians: (30 / 180.0) * math.Pi,
		EndstopStrengthUnit:  1,
		DetentStrengthUnit:   0.4,
		SnapPoint:            0.7,
		Text:                 hvm.positionNames[hvm.position],
	}
}
