package ignition

import (
	"fmt"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/internal/files"
	"github.com/tmacro/cola/internal/templates"
	"github.com/tmacro/cola/pkg/config"
)

func enablePodmanSysext(g *generator) {
	g.Files = append(g.Files,
		ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path:      "/etc/flatcar/enabled-sysext.conf",
				Overwrite: toPtr(true),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl("podman\n")),
				},
			},
		},
		ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path:      "/opt/cola/podman-tmpfiles-fix.sh",
				Overwrite: toPtr(true),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0755),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl(files.MustGetEmbeddedFile("podman-tmpfiles-fix.sh"))),
				},
			},
		},
		ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path:      "/etc/systemd/system/podman-tmpfiles-fix.service",
				Overwrite: toPtr(true),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0755),
				Contents: ignitionTypes.Resource{
					Source: toPtr(toDataUrl(files.MustGetEmbeddedFile("podman-tmpfiles-fix.service"))),
				},
			},
		},
	)
	g.Links = append(g.Links,
		ignitionTypes.Link{
			Node: ignitionTypes.Node{
				Path: "/etc/systemd/system/sysinit.target.wants/podman-tmpfiles-fix.service",
			},
			LinkEmbedded1: ignitionTypes.LinkEmbedded1{
				Hard:   toPtr(false),
				Target: toPtr("/etc/systemd/system/podman-tmpfiles-fix.service"),
			},
		},
	)
}

func generateContainers(cfg *config.ApplianceConfig, g *generator) error {
	// We need to enable the podman sysext to get the systemd generator
	if len(cfg.Containers) > 0 {
		enablePodmanSysext(g)
	}

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
