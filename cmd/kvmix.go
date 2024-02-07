package main

import (
	"io"
	"os"

	kvmix "github.com/marcuswu/KVMix"
	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/gosmartknob"
	skSerial "github.com/marcuswu/gosmartknob/serial"
	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
	"gopkg.in/yaml.v3"
)

func main() {
	ports := skSerial.FindPorts(gosmartknob.DefaultDeviceFilters, "", true)
	if len(ports) < 1 {
		log.Error().Msg("Could not find SmartKnob")
		return
	}

	portOpener := func() (io.ReadWriteCloser, error) {
		port, err := serial.Open(ports[0].Name, &serial.Mode{BaudRate: gosmartknob.Baud})
		if err != nil {
			log.Error().Err(err).Str("port", ports[0].Name).Msg("Failed to open port")
			return nil, err
		}
		return port, err
	}

	// read config file
	configData, err := os.ReadFile("config.yml")
	if err != nil {
		log.Error().Err(err).Msg("Unable to read the config file config.yml")
		return
	}
	config := config.Config{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Error().Err(err).Msg("Unable to parse the config file config.yml")
		return
	}

	mixer := kvmix.New(portOpener, config)
	for mixer.Running {
	}
	// mixer.RunMixer()
	// do stuff with mixer...
	// myDDC := ddc.NewDDC()
	// myDDC.SetInputSource(1, ddc.HDMI1) // From Windows to OSX
	// myDDC.SetInputSource(0, ddc.DISPLAY_PORT1) // From OSX to Windows
}
