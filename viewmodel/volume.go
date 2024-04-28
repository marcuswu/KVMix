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
	initialVolume, err := appChannel.GetVolume()
	initialVolume *= 100
	log.Debug().Int32("Initial Volume", int32(initialVolume)).Str("channel", appChannel.Name()).Msg("Volume ViewModel Initializer")
	if err != nil {
		log.Error().Err(err).Str("channel", appChannel.Name()).Msg("Error retrieving volume for channel")
	}
	return &VolumeViewModel{
		position:          int32(initialVolume),
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
		return
	}
	log.Debug().Int("volume", int(vvm.position)).Float64("float volume", float64(vvm.position)/100.0).Str("channel", vvm.appChannel.Name()).Msg("Setting volume")
	vvm.appChannel.SetVolume(float64(vvm.position) / 100.0)
	vvm.updateVolumeTimer = time.AfterFunc(time.Duration(300)*time.Millisecond, func() {
		vvm.updateVolumeTimer = nil
	})
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
	// We don't need to handle positionNonce (handleNonces) -- we aren't changing title
	if !vvm.getNonceSet() {
		handleNonces(vvm, state)
		ret.RegenConfig = true
		return ret
	}
	vvm.setPosition(state.CurrentPosition)
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

func (vvm *VolumeViewModel) GenerateConfigBeforeBack() bool {
	return false
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
