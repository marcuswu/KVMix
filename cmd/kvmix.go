package main

import (
	"io"
	"os"
	"os/signal"
	"sync"

	kvmix "github.com/marcuswu/KVMix"
	"github.com/marcuswu/KVMix/config"
	"github.com/marcuswu/gosmartknob"
	skSerial "github.com/marcuswu/gosmartknob/serial"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/tarm/serial"
)

func main() {
	var wg sync.WaitGroup
	exitchan := make(chan bool, 1)

	ports := skSerial.FindPorts(gosmartknob.DefaultDeviceFilters, "", true)
	if len(ports) < 1 {
		log.Error().Msg("Could not find SmartKnob")
		return
	}

	portOpener := func() (io.ReadWriteCloser, error) {
		log.Debug().Str("port", ports[0].Name).Msg("Opening port")
		port, err := serial.OpenPort(&serial.Config{Name: ports[0].Name, Baud: gosmartknob.Baud})
		if err != nil {
			log.Error().Err(err).Str("port", ports[0].Name).Msg("Failed to open port")
			return nil, err
		}
		log.Info().Str("port", ports[0].Name).Msg("Opened port")
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

	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan // Wait for a SIGINT

		// Close out our go routine gracefully
		exitchan <- true
		wg.Wait()
		mixer.Stop()

		os.Exit(0)
	}()

	<-exitchan

	// mixer.RunMixer()
	// do stuff with mixer...
	// myDDC := ddc.NewDDC()
	// myDDC.SetInputSource(1, ddc.HDMI1)         // From Windows to OSX
	// myDDC.SetInputSource(0, ddc.DISPLAY_PORT1) // From OSX to Windows
}
