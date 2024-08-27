package ignition

import (
	"fmt"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/internal/templates"
	"github.com/tmacro/cola/pkg/config"
)

func generateContainers(cfg *config.ApplianceConfig, g *generator) error {
	for _, container := range cfg.Containers {
		contents, err := templates.SystemdContainer(container)
		if err != nil {
			return fmt.Errorf("failed to format container unit contents: %v", err)
		}

		g.Files = append(g.Files, ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: fmt.Sprintf("/etc/containers/systemd/%s.container", container.Name),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0640),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl(contents)),
				},
			},
		})
	}
	return nil
}
