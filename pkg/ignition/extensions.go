package ignition

import (
	"fmt"
	"net/url"
	"strings"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/internal/files"
	"github.com/tmacro/cola/pkg/config"
	"github.com/vincent-petithory/dataurl"
)

var defaultExtensionFiles = []ignitionTypes.File{
	{
		Node: ignitionTypes.Node{
			Path: "/opt/bin/sysext-update",
		},
		FileEmbedded1: ignitionTypes.FileEmbedded1{
			Mode: toPtr(0644),
			Contents: ignitionTypes.Resource{
				Source: toPtr(toDataUrl(files.MustGetEmbeddedFile("sysext-update.sh"))),
			},
		},
	},
	{
		Node: ignitionTypes.Node{
			Path: "/etc/sysupdate.d/noop.conf",
		},
		FileEmbedded1: ignitionTypes.FileEmbedded1{
			Mode: toPtr(0644),
			Contents: ignitionTypes.Resource{
				Source: toPtr("https://github.com/flatcar/sysext-bakery/releases/download/latest/noop.conf"),
			},
		},
	},
}

var defaultExtensionUnits = []ignitionTypes.Unit{
	{
		Name:    "systemd-sysupdate.timer",
		Enabled: toPtr(true),
	},
	{
		Name:     "sysext-pre-update.service",
		Enabled:  toPtr(true),
		Contents: toPtr(files.MustGetEmbeddedFile("sysext-pre-update.service")),
	},
	{
		Name:     "sysext-post-update.service",
		Enabled:  toPtr(true),
		Contents: toPtr(files.MustGetEmbeddedFile("sysext-post-update.service")),
	},
}

func generateExtensions(cfg *config.ApplianceConfig, g *generator) error {
	if len(cfg.Extensions) == 0 {
		return nil
	}

	g.Files = append(g.Files, defaultExtensionFiles...)
	g.Units = append(g.Units, defaultExtensionUnits...)

	dropins := []ignitionTypes.Dropin{
		{
			Name:     "sysext.conf",
			Contents: toPtr("[Service]\nExecStartPost=systemctl restart systemd-sysext"),
		},
	}

	for _, ext := range cfg.Extensions {
		extPath := fmt.Sprintf("/opt/extensions/%s/%s", ext.Name, formatExtensionName(ext.Name, ext.Version, ext.Arch))

		g.Links = append(g.Links, ignitionTypes.Link{
			Node: ignitionTypes.Node{
				Path: "/etc/extensions/" + ext.Name + ".raw",
			},
			LinkEmbedded1: ignitionTypes.LinkEmbedded1{
				Hard:   toPtr(false),
				Target: toPtr(extPath),
			},
		})

		g.Files = append(g.Files, ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: extPath,
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(formatExtensionURL(ext.BakeryUrl, ext.Name, ext.Version, ext.Arch)),
				},
			},
		})

		g.Files = append(g.Files, ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: fmt.Sprintf("/etc/sysupdate.%s.d/%s.conf", ext.Name, ext.Name),
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(0644),
				Contents: ignitionTypes.Resource{
					Source: toPtr(formatExtensionTransferConfigUrl(ext.BakeryUrl, ext.Name)),
				},
			},
		})

		dropins = append(dropins, formatExtensionUpdateDropIn(ext.Name))
	}

	g.Units = append(g.Units, ignitionTypes.Unit{
		Name:    "systemd-sysupdate.service",
		Dropins: dropins,
	})

	return nil
}

func toDataUrl(data string) string {
	return (&url.URL{
		Scheme: "data",
		Opaque: "," + dataurl.EscapeString(data),
	}).String()
}

func formatExtensionName(name, version, arch string) string {
	if arch != "" {
		arch = "-" + arch
	}
	return fmt.Sprintf("%s-%s%s", name, version, arch)
}

func formatExtensionURL(bakeryURL, name, version, arch string) string {
	if !strings.HasSuffix(bakeryURL, "/") {
		bakeryURL += "/"
	}
	return fmt.Sprintf("%s%s-%s-%s.raw", bakeryURL, name, version, arch)
}

func formatExtensionTransferConfigUrl(bakery, name string) string {
	if !strings.HasSuffix(bakery, "/") {
		bakery += "/"
	}
	return fmt.Sprintf("%s%s.conf", bakery, name)
}

func formatExtensionUpdateDropIn(name string) ignitionTypes.Dropin {
	return ignitionTypes.Dropin{
		Name:     name + ".conf",
		Contents: toPtr(fmt.Sprintf("[Service]\nExecStartPre=/usr/lib/systemd/systemd-sysupdate -C %s update", name)),
	}
}
