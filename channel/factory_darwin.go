package channel

import (
	"github.com/progrium/macdriver/objc"
)

const maxVolume = 100

type DarwinChannelFactory struct {
}

func New() ChannelFactory {
	return DarwinChannelFactory{}
}

func (dcf DarwinChannelFactory) Channels() ([]Channel, error) {
	apps := runningApplications()
	channels := make([]Channel, 0, len(apps))

	channels = append(channels, DarwinMainChannel{name: "Main Volume"})

	for _, name := range apps {
		channel := DarwinChannel{name: name}
		channels = append(channels, channel)
	}
	return channels, nil
}

func runningApplications() []string {
	applications := []string{}
	apps := objc.Get("NSWorkspace").Get("sharedWorkspace").Get("runningApplications")

	for i := int64(0); i < apps.Get("count").Int(); i++ {
		applications = append(applications, apps.Send("objectAtIndex:", i).Send("localizedName").String())
	}

	return applications
}
