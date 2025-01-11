package ignition

import (
	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/internal/files"
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

func generatePowerProfile(cfg *config.ApplianceConfig, g *generator) error {
	if cfg.System.PowerProfile == "" {
		return nil
	}

	g.Files = append(g.Files,
		ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: "/opt/bin/set-cpu-governor.sh",
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0755),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl(files.MustGetEmbeddedFile("set-cpu-governor.sh"))),
				},
			},
		},
		ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: "/etc/default/set_cpu_governor",
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl("POWER_PROFILE=" + cfg.System.PowerProfile + "\n")),
				},
			},
		},
	)

	g.Units = append(g.Units, ignitionTypes.Unit{
		Name:     "set-cpu-governor.service",
		Enabled:  toPtr(true),
		Contents: toPtr(files.MustGetEmbeddedFile("set-cpu-governor.service")),
	})

	return nil
}
