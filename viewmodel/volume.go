package viewmodel

import (
	"fmt"
	"math"
	"time"

	"github.com/marcuswu/KVMix/channel"
	"github.com/marcuswu/gosmartknob/pb"
	"github.com/rs/zerolog/log"
)

type VolumeViewModel struct {
	position          int32
	positionNonce     uint32
	pressNonce        uint32
	appChannel        channel.Channel
	setNonce          bool
	updateVolumeTimer *time.Timer
}

func NewVolumeViewModel(pressNonce uint32, appChannel channel.Channel) *VolumeViewModel {
	return &VolumeViewModel{
		position:          0,
		positionNonce:     0,
		pressNonce:        0,
		appChannel:        appChannel,
		setNonce:          false,
		updateVolumeTimer: nil,
	}
}

func (vvm *VolumeViewModel) getNonceSet() bool {
	return vvm.setNonce
}

func (vvm *VolumeViewModel) setNonceSet(isSet bool) {
	vvm.setNonce = isSet
}

func (vvm *VolumeViewModel) getPositionNonce() uint32 {
	return vvm.positionNonce
}

func (vvm *VolumeViewModel) setPositionNonce(nonce uint32) {
	vvm.positionNonce = nonce
}

func (vvm *VolumeViewModel) getPosition() int32 {
	return vvm.position
}

func (vvm *VolumeViewModel) setPosition(pos int32) {
	vvm.position = pos
	if vvm.updateVolumeTimer != nil {
		log.Debug().Msg("Resetting volume timer")
		vvm.updateVolumeTimer.Reset(time.Duration(250) * time.Millisecond)
	} else {
		log.Debug().Msg("Updating volume timer")
		vvm.updateVolumeTimer = time.AfterFunc(time.Duration(250)*time.Millisecond, func() {
			log.Debug().Int("volume", int(vvm.position)).Msg("Setting volume")
			vvm.appChannel.SetVolume(float64(vvm.position) / 100.0)
			vvm.updateVolumeTimer = nil
		})
	}
}

func (vvm *VolumeViewModel) getPressNonce() uint32 {
	return vvm.pressNonce
}

func (vvm *VolumeViewModel) setPressNonce(nonce uint32) {
	vvm.pressNonce = nonce
}

func (vvm *VolumeViewModel) HandleMessage(state *pb.SmartKnobState) NavAction {
	// Change volume of app, setting the volume as we go
	// Press sends us back one ViewModel
	ret := NavAction{
		Navigation:  NavNone,
		ViewModel:   nil,
		RegenConfig: false,
	}
	// Don't set RegenConfig from handleNonces because the title doesn't change
	handleNonces(vvm, state)
	if state.PressNonce != vvm.pressNonce {
		ret.Navigation = NavBack
		ret.RegenConfig = true
		ret.pressNonce = state.PressNonce
	}
	return ret
}

func (vvm *VolumeViewModel) Restore(state *pb.SmartKnobState) {
	restorePosition(vvm, state)
}

func (vvm *VolumeViewModel) GenerateConfig() *pb.SmartKnobConfig {
	title := fmt.Sprintf("Volume for\n%s", vvm.appChannel.Name())
	return &pb.SmartKnobConfig{
		Position:             vvm.position,
		PositionNonce:        vvm.positionNonce,
		MinPosition:          0,
		MaxPosition:          100,
		PositionWidthRadians: (3 / 180.0) * math.Pi,
		DetentStrengthUnit:   0.0,
		EndstopStrengthUnit:  1.0,
		SnapPoint:            0.7,
		Text:                 title,
	}
}
