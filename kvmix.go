package kvmix

import (
	"io"

	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/KVMix/ddc"
	"github.com/marcuswu/KVMix/viewmodel"
	"github.com/marcuswu/gosmartknob"
	"github.com/marcuswu/gosmartknob/pb"
)

type PortOpener func() (io.ReadWriteCloser, error)

type KVMix struct {
	smartknob  *gosmartknob.SmartKnob
	config     config.Config
	ddc        ddc.DDC
	viewModels []viewmodel.ViewModel
	portOpener PortOpener
	Running    bool
}

func New(portOpener PortOpener, config config.Config) *KVMix {
	mixer := &KVMix{config: config, Running: true}
	port, _ := portOpener()
	sk := gosmartknob.New(port, func(message *pb.FromSmartKnob) {
		mixer.handleMessage(message)
	}, func() { mixer.handleClosed() })
	mixer.smartknob = sk
	mixer.ddc = ddc.NewDDC()
	mixer.viewModels = make([]viewmodel.ViewModel, 0, 3)
	home := viewmodel.NewHomeViewModel(0, config, mixer.ddc)
	mixer.viewModels = append(mixer.viewModels, home)
	mixer.smartknob.SendConfig(mixer.generateConfig())
	return mixer
}

func (mix *KVMix) generateConfig() *pb.SmartKnobConfig {
	return mix.viewModels[len(mix.viewModels)-1].GenerateConfig()
}

func (mix *KVMix) handleMessage(message *pb.FromSmartKnob) {
	// ViewModel can change the SmartKnob config based on messages
	// It returns true if a configuration update is necessary
	navAction := mix.viewModels[len(mix.viewModels)-1].HandleMessage(message.GetSmartknobState())
	switch navAction.Navigation {
	case viewmodel.NavTo:
		if navAction.ViewModel != nil {
			mix.viewModels = append(mix.viewModels, navAction.ViewModel)
		}
		navAction.RegenConfig = true
	case viewmodel.NavBack:
		mix.viewModels = mix.viewModels[:len(mix.viewModels)-1]
		// Ensure initial state going back is valid (pressNonce has probably changed)
		mix.viewModels[len(mix.viewModels)-1].HandleMessage(message.GetSmartknobState())
		navAction.RegenConfig = true
	}
	if navAction.RegenConfig {
		mix.generateConfig()
	}
}

func (mix *KVMix) handleClosed() {
	mix.Running = false
	newPort, _ := mix.portOpener()
	mix.smartknob.SetReadWriter(newPort)
	mix.Running = true
}
