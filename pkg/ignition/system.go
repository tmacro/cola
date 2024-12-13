package ignition

import (
	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

func generateHostname(cfg *config.ApplianceConfig, g *generator) error {
	if cfg.System.Hostname != "" {
		f := ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path:      "/etc/hostname",
				Overwrite: toPtr(true),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl(cfg.System.Hostname + "\n")),
				},
			},
		}

		g.Files = append(g.Files, f)
	}

	return nil
}
