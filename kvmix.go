package kvmix

import (
	"io"

	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/KVMix/ddc"
	"github.com/marcuswu/KVMix/viewmodel"
	"github.com/marcuswu/gosmartknob"
	"github.com/marcuswu/gosmartknob/pb"
	"github.com/rs/zerolog/log"
)

type PortOpener func() (io.ReadWriteCloser, error)

type KVMix struct {
	smartknob   *gosmartknob.SmartKnob
	config      config.Config
	ddc         ddc.DDC
	viewModels  []viewmodel.ViewModel
	portOpener  PortOpener
	Running     bool
	configNonce int64
}

func New(portOpener PortOpener, config config.Config) *KVMix {
	mixer := &KVMix{config: config, Running: true}
	mixer.ddc = ddc.NewDDC()
	mixer.viewModels = make([]viewmodel.ViewModel, 0, 3)
	home := viewmodel.NewHomeViewModel(config, mixer.ddc)
	mixer.viewModels = append(mixer.viewModels, home)
	mixer.portOpener = portOpener
	port, _ := mixer.portOpener()
	sk := gosmartknob.New(port, func(message *pb.FromSmartKnob) {
		mixer.handleMessage(message)
	}, func() { log.Debug().Msg("connection closed"); mixer.handleClosed() })
	mixer.smartknob = sk
	nonce, _ := mixer.smartknob.SendConfig(mixer.generateConfig())
	mixer.configNonce = int64(nonce)
	return mixer
}

func (mix *KVMix) generateConfig() *pb.SmartKnobConfig {
	return mix.viewModels[len(mix.viewModels)-1].GenerateConfig()
}

func (mix *KVMix) handleMessage(message *pb.FromSmartKnob) {
	if mix.configNonce > 0 {
		log.Debug().Int64("desiredNonce", mix.configNonce).Msg("handleMessage looking for ack")
	}
	ack := message.GetAck()
	if ack != nil {
		log.Debug().Int64("nonce", int64(ack.Nonce)).Msg("handleMessage got ack")
	}
	if ack != nil && mix.configNonce == int64(ack.Nonce) {
		mix.configNonce = -1
	}
	// Discard everything from the SmartKnob until the last sent config is acknowledged
	if mix.configNonce > 0 {
		return
	}

	// ViewModel can change the SmartKnob config based on messages
	// It returns true if a configuration update is necessary
	state := message.GetSmartknobState()
	if state == nil {
		return
	}
	navAction := mix.viewModels[len(mix.viewModels)-1].HandleMessage(state)
	switch navAction.Navigation {
	case viewmodel.NavTo:
		if navAction.ViewModel != nil {
			mix.viewModels = append(mix.viewModels, navAction.ViewModel)
		}
		navAction.RegenConfig = true
	case viewmodel.NavBack:
		mix.viewModels = mix.viewModels[:len(mix.viewModels)-1]
		// Ensure initial state going back is valid (pressNonce has probably changed)
		mix.viewModels[len(mix.viewModels)-1].Restore(message.GetSmartknobState())
		navAction.RegenConfig = true
	}
	if navAction.RegenConfig {
		nonce, _ := mix.smartknob.SendConfig(mix.generateConfig())
		mix.configNonce = int64(nonce)
	}
}

func (mix *KVMix) handleClosed() {
	newPort, _ := mix.portOpener()
	mix.smartknob.SetReadWriter(newPort)
}

func (mix *KVMix) Stop() {
	mix.Running = false
}
