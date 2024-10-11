package ignition

import (
	"fmt"
	"strings"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/internal/templates"
	"github.com/tmacro/cola/pkg/config"
)

func generateInterfaces(cfg *config.ApplianceConfig, g *generator) error {
	for _, iface := range cfg.Interfaces {
		ifaceNet, err := templates.SystemdNetwork(iface)
		if err != nil {
			return fmt.Errorf("failed to format systemd network contents: %v", err)
		}

		ifaceName := strings.ReplaceAll(iface.Name, "*", "")
		filename := fmt.Sprintf("/etc/systemd/network/10-%s.network", ifaceName)

		g.Files = append(g.Files, ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: filename,
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl(ifaceNet)),
				},
			},
		})

		for _, vlan := range iface.VLANs {
			vlanNet, err := templates.SystemdVlanNetwork(vlan)
			if err != nil {
				return fmt.Errorf("failed to format systemd network contents: %v", err)
			}

			g.Files = append(g.Files, ignitionTypes.File{
				Node: ignitionTypes.Node{
					Path: fmt.Sprintf("/etc/systemd/network/20-%s.network", vlan.Name),
				},
				FileEmbedded1: ignitionTypes.FileEmbedded1{
					Mode: toPtr(0644),
					Contents: ignitionTypes.Resource{
						Source: toPtr(toDataUrl(vlanNet)),
					},
				},
			})

			vlanNetdev, err := templates.SystemdVlanNetDev(vlan)
			if err != nil {
				return fmt.Errorf("failed to format systemd netdev contents: %v", err)
			}

			g.Files = append(g.Files, ignitionTypes.File{
				Node: ignitionTypes.Node{
					Path: fmt.Sprintf("/etc/systemd/network/00-%s.netdev", vlan.Name),
				},
				FileEmbedded1: ignitionTypes.FileEmbedded1{
					Mode: toPtr(0644),
					Contents: ignitionTypes.Resource{
						Source: toPtr(toDataUrl(vlanNetdev)),
					},
				},
			})
		}
	}

	return nil
}
