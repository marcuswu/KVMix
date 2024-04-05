package channel

import (
	"github.com/progrium/macdriver/objc"
	"golang.org/x/exp/maps"
)

const maxVolume = 100

type DarwinChannelFactory struct {
}

func New() ChannelFactory {
	return DarwinChannelFactory{}
}

func (dcf DarwinChannelFactory) ChannelsMatching(ids map[string]interface{}) ([]Channel, error) {
	apps := runningApplications()
	channels := make([]Channel, 0, len(apps))

	channels = append(channels, DarwinMainChannel{name: "Main Volume"})

	for _, name := range apps {
		if _, ok := ids[name]; ok {
			channel := DarwinChannel{name: name}
			channels = append(channels, channel)
		}
	}
	return channels, nil
}

func runningApplications() []string {
	applications := make(map[string]struct{})
	apps := objc.Get("NSWorkspace").Get("sharedWorkspace").Get("runningApplications")

	for i := int64(0); i < apps.Get("count").Int(); i++ {
		appName := apps.Send("objectAtIndex:", i).Send("localizedName").String()
		if _, ok := applications[appName]; ok {
			continue
		}
		applications[appName] = struct{}{}
	}

	return maps.Keys(applications)
}
