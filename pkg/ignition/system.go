package ignition

import (
	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

func generateHostname(cfg *config.ApplianceConfig, g *generator) error {
	if cfg.System.Hostname != "" {
		g.Files = append(g.Files, ignitionTypes.File{
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
		})
	}

	return nil
}

func generateUpdateConfig(cfg *config.ApplianceConfig, g *generator) error {
	if cfg.System.Updates != nil {
		g.Files = append(g.Files, ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path:      "/etc/flatcar/update.conf",
				Overwrite: toPtr(true),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl("REBOOT_STRATEGY=" + cfg.System.Updates.RebootStrategy + "\n")),
				},
			},
		})
	}

	return nil
}
