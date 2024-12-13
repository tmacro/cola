package ignition

import (
	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
)

func defaultConfig() ignitionTypes.Config {
	return ignitionTypes.Config{
		Ignition: ignitionTypes.Ignition{
			Version: "3.4.0",
		},
	}
}
