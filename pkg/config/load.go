package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

var (
	ErrNoSystemBlock = fmt.Errorf("system block is required")
)

func getConfigFilesInDir(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0)

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".hcl" {
			continue
		}

		paths = append(paths, path+"/"+file.Name())
	}

	return paths, nil
}

func ReadConfig(paths, bases []string, strict bool) (*ApplianceConfig, error) {
	config, err := readConfig(paths, strict)
	if err != nil {
		return nil, err
	}

	if len(bases) > 0 {
		baseConfig, err := readConfig(bases, false)
		if err != nil {
			return nil, err
		}

		config = MergeConfigs(baseConfig, config)
	}

	return config, nil
}

func readConfig(paths []string, strict bool) (*ApplianceConfig, error) {
	configFilePaths := make([]string, 0)
	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if fi.IsDir() {
			dirPaths, err := getConfigFilesInDir(path)
			if err != nil {
				return nil, err
			}

			configFilePaths = append(configFilePaths, dirPaths...)
		} else {
			configFilePaths = append(configFilePaths, path)
		}
	}

	return readConfigFiles(configFilePaths, strict)
}

func readConfigFile(path string, strict bool) (*ApplianceConfig, error) {
	var config ApplianceConfig
	err := hclsimple.DecodeFile(path, nil, &config)
	if err != nil {
		return nil, ParseError{Err: err, Path: path}
	}

	if strict && config.System == nil {
		return nil, ErrNoSystemBlock
	}

	if len(config.Files) > 0 {
		files := make([]File, len(config.Files))
		for i, file := range config.Files {
			f := file
			if file.SourcePath != "" && !filepath.IsAbs(file.SourcePath) {
				f.SourcePath = filepath.Join(filepath.Dir(path), file.SourcePath)
			}
			files[i] = f
		}

		config.Files = files
	}

	if len(config.Services) > 0 {
		services := make([]Service, len(config.Services))
		for i, service := range config.Services {
			svc := service
			if service.SourcePath != "" && !filepath.IsAbs(service.SourcePath) {
				svc.SourcePath = filepath.Join(filepath.Dir(path), service.SourcePath)
			}

			if len(service.DropIns) > 0 {
				dropins := make([]DropIn, len(service.DropIns))
				for j, dropin := range service.DropIns {
					drp := dropin
					if dropin.SourcePath != "" && !filepath.IsAbs(dropin.SourcePath) {
						drp.SourcePath = filepath.Join(filepath.Dir(path), dropin.SourcePath)
					}
					dropins[j] = drp
				}
				svc.DropIns = dropins
			}
			services[i] = svc
		}
		config.Services = services
	}

	return &config, nil
}

func readConfigFiles(paths []string, strict bool) (*ApplianceConfig, error) {
	var merged *ApplianceConfig

	for _, path := range paths {
		config, err := readConfigFile(path, false)
		if err != nil {
			return nil, ParseError{Err: err, Path: path}
		}

		merged = MergeConfigs(merged, config)
	}

	if strict && merged.System == nil {
		return nil, ErrNoSystemBlock
	}

	return merged, nil
}

func MergeConfigs(base, override *ApplianceConfig) *ApplianceConfig {
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

	if base.Etcd == nil {
		base.Etcd = override.Etcd
	} else {
		if override.Etcd != nil && override.Etcd.Server {
			base.Etcd.Server = true
		}

		if override.Etcd != nil && override.Etcd.Gateway {
			base.Etcd.Gateway = true
		}

		if override.Etcd != nil && override.Etcd.Name != "" {
			base.Etcd.Name = override.Etcd.Name
		}

		if override.Etcd != nil && override.Etcd.ListenAddress != "" {
			base.Etcd.ListenAddress = override.Etcd.ListenAddress
		}

		if override.Etcd != nil && override.Etcd.InitialToken != "" {
			base.Etcd.InitialToken = override.Etcd.InitialToken
		}

		if override.Etcd != nil && len(override.Etcd.Peers) > 0 {
			base.Etcd.Peers = append(base.Etcd.Peers, override.Etcd.Peers...)
		}
	}

	base.Users = append(base.Users, override.Users...)
	base.Extensions = append(base.Extensions, override.Extensions...)
	base.Containers = append(base.Containers, override.Containers...)
	base.Files = append(base.Files, override.Files...)
	base.Directories = append(base.Directories, override.Directories...)
	base.Symlinks = append(base.Symlinks, override.Symlinks...)
	base.Mounts = append(base.Mounts, override.Mounts...)
	base.Interfaces = append(base.Interfaces, override.Interfaces...)
	base.Services = append(base.Services, override.Services...)

	return base
}
