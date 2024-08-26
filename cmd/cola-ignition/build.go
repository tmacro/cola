package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	clConfig "github.com/flatcar/container-linux-config-transpiler/config"
	clcTypes "github.com/flatcar/container-linux-config-transpiler/config/types"
	ignitionTypes "github.com/flatcar/ignition/config/v2_3/types"
	"github.com/tmacro/cola/internal/files"
	"github.com/tmacro/cola/internal/templates"
	"github.com/tmacro/cola/pkg/config"
	"gopkg.in/yaml.v3"
)

func defaultCLConfig() *clcTypes.Config {
	return &clcTypes.Config{
		Storage: clcTypes.Storage{
			Directories: []clcTypes.Directory{},
			Links:       []clcTypes.Link{},
			Files: []clcTypes.File{
				{
					Path: "/etc/flatcar/enabled-sysext.conf",
					Mode: Ptr(0644),
					Contents: clcTypes.FileContents{
						Inline: "podman\n",
					},
				},
				{
					Path: "/etc/tmpfiles.d/podman.conf",
					Mode: Ptr(0644),
					Contents: clcTypes.FileContents{
						Inline: "C+ /etc/containers - - - - /usr/share/podman/etc/containers\n",
					},
				},
				{
					Path: "/etc/default/cpu_governor",
					Mode: Ptr(0644),
					Contents: clcTypes.FileContents{
						Inline: "conservative\n",
					},
				},
			},
		},
		Systemd: clcTypes.Systemd{
			Units: []clcTypes.SystemdUnit{
				{
					Name:     "cpu-governor.service",
					Enabled:  Ptr(true),
					Contents: files.MustGetEmbeddedFile("cpu-governor.service"),
				},
				{
					Name:     "cpu-governor.sh",
					Enabled:  Ptr(true),
					Contents: files.MustGetEmbeddedFile("cpu-governor.sh"),
				},
			},
		},
	}
}

func defaultExtensionFiles() []clcTypes.File {
	return []clcTypes.File{
		{
			Path: "/opt/bin/sysext-update",
			Mode: Ptr(0755),
			Contents: clcTypes.FileContents{
				Inline: files.MustGetEmbeddedFile("sysext-update.sh"),
			},
		},
		{
			Path: "/etc/sysupdate.d/noop.conf",
			Mode: Ptr(0644),
			Contents: clcTypes.FileContents{
				Remote: clcTypes.Remote{
					Url: "https://github.com/flatcar/sysext-bakery/releases/download/latest/noop.conf",
				},
			},
		},
	}
}

func defaultExtensionUnits() []clcTypes.SystemdUnit {
	return []clcTypes.SystemdUnit{
		{
			Name:    "systemd-sysupdate.timer",
			Enabled: Ptr(true),
		},
		{
			Name:     "sysext-pre-update.service",
			Enabled:  Ptr(true),
			Contents: files.MustGetEmbeddedFile("sysext-pre-update.service"),
		},
		{
			Name:     "sysext-post-update.service",
			Enabled:  Ptr(true),
			Contents: files.MustGetEmbeddedFile("sysext-post-update.service"),
		},
	}
}

func formatExtensionLinkPath(name string) string {
	return "/etc/extensions/" + name + ".raw"
}

func formatExtensionTargetPath(name, version, arch string) string {
	if arch != "" {
		arch = "-" + arch
	}
	return fmt.Sprintf("/opt/extensions/%s/%s-%s%s.raw", name, name, version, arch)
}

func formatExtensionURL(bakeryURL, name, version, arch string) string {
	if !strings.HasSuffix(bakeryURL, "/") {
		bakeryURL += "/"
	}
	return fmt.Sprintf("%s%s-%s-%s.raw", bakeryURL, name, version, arch)
}

func formatExtensionUpdateDropIn(name string) clcTypes.SystemdUnitDropIn {
	return clcTypes.SystemdUnitDropIn{
		Name:     name + ".conf",
		Contents: fmt.Sprintf("[Service]\nExecStartPre=/usr/lib/systemd/systemd-sysupdate -C %s update", name),
	}
}

func formatExtensionTransferConfigPath(name string) string {
	return fmt.Sprintf("/etc/sysupdate.%s.d/%s.conf", name, name)
}

func formatExtensionTransferConfigUrl(bakery, name string) string {
	if !strings.HasSuffix(bakery, "/") {
		bakery += "/"
	}
	return fmt.Sprintf("%s%s.conf", bakery, name)
}

func formatContainerUnitPath(name string) string {
	return fmt.Sprintf("/etc/containers/systemd/%s.container", name)
}

func formatMountUnitPath(mountPoint string) string {
	mountName := strings.ReplaceAll(mountPoint[1:], "/", "-")
	return fmt.Sprintf("/etc/systemd/system/%s.mount", mountName)
}

func buildIgnitionConfig(cfg *config.ApplianceConfig) (*ignitionTypes.Config, error) {
	clc := defaultCLConfig()

	clc.Storage.Files = append(clc.Storage.Files, clcTypes.File{
		Path: "/etc/hostname",
		Mode: Ptr(0644),
		Contents: clcTypes.FileContents{
			Inline: cfg.System.Hostname + "\n",
		},
	})

	for _, user := range cfg.Users {
		clc.Passwd.Users = append(clc.Passwd.Users, clcTypes.User{
			Name:              user.Username,
			Groups:            user.Groups,
			HomeDir:           user.HomeDir,
			Shell:             user.Shell,
			SSHAuthorizedKeys: user.SSHAuthorizedKeys,
		})
	}

	if len(cfg.Extensions) > 0 {
		clc.Storage.Files = append(clc.Storage.Files, defaultExtensionFiles()...)
		clc.Systemd.Units = append(clc.Systemd.Units, defaultExtensionUnits()...)

		updateDropIns := []clcTypes.SystemdUnitDropIn{
			{
				Name:     "sysext.conf",
				Contents: "[Service]\nExecStartPost=systemctl restart systemd-sysext",
			},
		}

		for _, extension := range cfg.Extensions {
			clc.Storage.Links = append(clc.Storage.Links, clcTypes.Link{
				Path:   formatExtensionLinkPath(extension.Name),
				Target: formatExtensionTargetPath(extension.Name, extension.Version, extension.Arch),
				Hard:   false,
			})

			clc.Storage.Files = append(clc.Storage.Files, clcTypes.File{
				Path: formatExtensionTargetPath(extension.Name, extension.Version, extension.Arch),
				Mode: Ptr(0644),
				Contents: clcTypes.FileContents{
					Remote: clcTypes.Remote{
						Url: formatExtensionURL(extension.BakeryUrl, extension.Name, extension.Version, extension.Arch),
					},
				},
			})

			clc.Storage.Files = append(clc.Storage.Files, clcTypes.File{
				Path: formatExtensionTransferConfigPath(extension.Name),
				Mode: Ptr(0644),
				Contents: clcTypes.FileContents{
					Remote: clcTypes.Remote{
						Url: formatExtensionTransferConfigUrl(extension.BakeryUrl, extension.Name),
					},
				},
			})

			updateDropIns = append(updateDropIns, formatExtensionUpdateDropIn(extension.Name))
		}

		clc.Systemd.Units = append(clc.Systemd.Units, clcTypes.SystemdUnit{
			Name:    "systemd-sysupdate.service",
			Dropins: updateDropIns,
		})
	}

	for _, container := range cfg.Containers {
		contents, err := templates.SystemdContainer(container)
		if err != nil {
			return nil, fmt.Errorf("failed to format container unit contents: %v", err)
		}

		clc.Storage.Files = append(clc.Storage.Files, clcTypes.File{
			Path:     formatContainerUnitPath(container.Name),
			Mode:     Ptr(0644),
			Contents: clcTypes.FileContents{Inline: contents},
		})
	}

	for _, file := range cfg.Files {
		mode, err := parseOctal(file.Mode)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file mode %s: %v", file.Mode, err)
		}
		f := clcTypes.File{
			Path: file.Path,
			Mode: Ptr(mode),
		}

		if file.Owner != "" {
			f.User = &clcTypes.FileUser{Name: file.Owner}
		}

		if file.Group != "" {
			f.Group = &clcTypes.FileGroup{Name: file.Group}
		}

		if file.Inline != "" {
			f.Contents.Inline = file.Inline
		} else if file.SourcePath != "" {
			content, err := os.ReadFile(file.SourcePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %v", file.SourcePath, err)
			}
			f.Contents.Inline = string(content)
		} else if file.URL != "" {
			f.Contents.Remote.Url = file.URL
		}

		clc.Storage.Files = append(clc.Storage.Files, f)
	}

	for _, dir := range cfg.Directories {
		mode, err := parseOctal(dir.Mode)
		if err != nil {
			return nil, fmt.Errorf("failed to parse file mode %s: %v", dir.Mode, err)
		}
		d := clcTypes.Directory{
			Path: dir.Path,
			Mode: Ptr(mode),
		}

		if dir.Owner != "" {
			d.User = &clcTypes.FileUser{Name: dir.Owner}
		}

		if dir.Group != "" {
			d.Group = &clcTypes.FileGroup{Name: dir.Group}
		}

		clc.Storage.Directories = append(clc.Storage.Directories, d)
	}

	for _, mount := range cfg.Mounts {
		contents, err := templates.SystemdMount(mount)
		if err != nil {
			return nil, fmt.Errorf("failed to format systemd mount contents: %v", err)
		}

		clc.Systemd.Units = append(clc.Systemd.Units, clcTypes.SystemdUnit{
			Name:     formatMountUnitPath(mount.MountPoint),
			Enabled:  Ptr(true),
			Contents: contents,
		})
	}

	for _, iface := range cfg.Interfaces {
		ifaceNet, err := templates.SystemdNetwork(iface)
		if err != nil {
			return nil, fmt.Errorf("failed to format systemd network contents: %v", err)
		}

		clc.Networkd.Units = append(clc.Networkd.Units, clcTypes.NetworkdUnit{
			Name:     "10-" + iface.Name + ".network",
			Contents: ifaceNet,
		})

		for _, vlan := range iface.VLANs {
			vlanNet, err := templates.SystemdVlanNetwork(vlan)
			if err != nil {
				return nil, fmt.Errorf("failed to format systemd vlan network contents: %v", err)
			}

			clc.Networkd.Units = append(clc.Networkd.Units, clcTypes.NetworkdUnit{
				Name:     "20-" + vlan.Name + ".network",
				Contents: vlanNet,
			})

			vlanNetdev, err := templates.SystemdVlanNetDev(vlan)
			if err != nil {
				return nil, fmt.Errorf("failed to format systemd vlan netdev contents: %v", err)
			}

			clc.Networkd.Units = append(clc.Networkd.Units, clcTypes.NetworkdUnit{
				Name:     "00-" + vlan.Name + ".netdev",
				Contents: vlanNetdev,
			})
		}

	}

	return convertCLConfig(clc)
}

func convertCLConfig(cfg *clcTypes.Config) (*ignitionTypes.Config, error) {
	serialized, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CL config: %v", err)
	}

	parsed, ast, report := clConfig.Parse(serialized)
	if len(report.Entries) > 0 {
		os.Stdout.Write(serialized)
		return nil, fmt.Errorf("failed to parse CL config: %v", report)
	}

	ignCfg, report := clConfig.Convert(parsed, "", ast)
	if len(report.Entries) > 0 {
		return nil, fmt.Errorf("failed to convert CL config: %v", report)
	}

	return &ignCfg, nil
}

func Ptr[T any](v T) *T {
	return &v
}

func parseOctal(s string) (int, error) {
	val, err := strconv.ParseInt(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}
