package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type ApplianceConfig struct {
	System      *System     `hcl:"system,block"`
	Users       []User      `hcl:"user,block"`
	Extensions  []Extension `hcl:"extension,block"`
	Containers  []Container `hcl:"container,block"`
	Files       []File      `hcl:"file,block"`
	Directories []Directory `hcl:"directory,block"`
	Mounts      []Mount     `hcl:"mount,block"`
}

type System struct {
	Hostname string `hcl:"hostname"`
	Timezone string `hcl:"timezone,optional"`
}

type User struct {
	Username          string   `hcl:"username,label"`
	Groups            []string `hcl:"groups,optional"`
	HomeDir           string   `hcl:"home_dir,optional"`
	Shell             string   `hcl:"shell,optional"`
	SSHAuthorizedKeys []string `hcl:"ssh_authorized_keys,optional"`
}

type Extension struct {
	Name      string `hcl:"name,label"`
	Version   string `hcl:"version"`
	Arch      string `hcl:"arch,optional"`
	BakeryUrl string `hcl:"bakery_url"`
}

type Container struct {
	Name    string   `hcl:"name,label"`
	Image   string   `hcl:"image"`
	Args    []string `hcl:"args,optional"`
	Volumes []Volume `hcl:"volume,block"`
}

type Volume struct {
	Source string `hcl:"source,label"`
	Target string `hcl:"target"`
}

type File struct {
	Path       string `hcl:"path,label"`
	Owner      string `hcl:"owner,optional"`
	Group      string `hcl:"group,optional"`
	Mode       string `hcl:"mode"`
	Inline     string `hcl:"inline,optional"`
	SourcePath string `hcl:"source_path,optional"`
	URL        string `hcl:"url,optional"`
}

type Directory struct {
	Path  string `hcl:"path,label"`
	Owner string `hcl:"owner,optional"`
	Group string `hcl:"group,optional"`
	Mode  string `hcl:"mode"`
}

type Mount struct {
	MountPoint string `hcl:"mount_point,label"`
	Type       string `hcl:"type"`
	What       string `hcl:"what"`
	Where      string `hcl:"where"`
	Options    string `hcl:"options,optional"`
}

func ReadConfig(path string, strict bool) (*ApplianceConfig, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return readConfigDir(path, strict)
	}

	return readConfigFile(path, strict)
}

func readConfigDir(path string, strict bool) (*ApplianceConfig, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var merged *ApplianceConfig

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".hcl" {
			continue
		}

		config, err := readConfigFile(path+"/"+file.Name(), false)
		if err != nil {
			return nil, err
		}

		merged = mergeConfigs(merged, config)
	}

	if merged == nil {
		return nil, fmt.Errorf("no config files found")
	}

	if strict && merged.System == nil {
		return nil, fmt.Errorf("system block is required")
	}

	return merged, nil
}

func readConfigFile(path string, strict bool) (*ApplianceConfig, error) {
	var config ApplianceConfig
	err := hclsimple.DecodeFile(path, nil, &config)
	if err != nil {
		return nil, err
	}

	if strict && config.System == nil {
		return nil, fmt.Errorf("system block is required")
	}

	return &config, nil
}

func mergeConfigs(base, override *ApplianceConfig) *ApplianceConfig {
	if base == nil {
		return override
	}

	if base.System == nil {
		base.System = override.System
	} else {
		if override.System != nil && override.System.Hostname != "" {
			base.System.Hostname = override.System.Hostname
		}

		if override.System != nil && override.System.Timezone != "" {
			base.System.Timezone = override.System.Timezone
		}
	}

	base.Users = append(base.Users, override.Users...)
	base.Extensions = append(base.Extensions, override.Extensions...)
	base.Containers = append(base.Containers, override.Containers...)

	return base
}
