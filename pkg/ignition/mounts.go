package ignition

import (
	"fmt"
	"strings"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/internal/templates"
	"github.com/tmacro/cola/pkg/config"
)

func generateMounts(cfg *config.ApplianceConfig, g *generator) error {
	for _, mount := range cfg.Mounts {
		contents, err := templates.SystemdMount(mount)
		if err != nil {
			return fmt.Errorf("failed to format systemd mount contents: %v", err)
		}

		mountName := strings.TrimPrefix(mount.Where, "/")
		mountName = strings.ReplaceAll(mountName, "/", "-")

		g.Files = append(g.Files, ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: fmt.Sprintf("/etc/systemd/system/%s.mount", mountName),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl(contents)),
				},
			},
		})

	}

	return nil
}
