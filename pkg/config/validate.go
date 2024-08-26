package config

import "fmt"

func ValidateConfig(config *ApplianceConfig) error {
	if config.System == nil {
		return ErrNoSystemBlock
	}

	for _, validator := range validators {
		if err := validator(config); err != nil {
			return err
		}
	}

	return nil
}

var validators = []func(*ApplianceConfig) error{
	validateUsers,
	validateExtensions,
	validateContainers,
	validateFiles,
	validateDirectories,
	// validateMounts,
	validateInterfaces,
}

func validateUsers(config *ApplianceConfig) error {
	for i, user := range config.Users {
		if user.Username == "" {
			return fmt.Errorf("user[%d].username is required", i)
		}
	}

	return nil
}

func validateExtensions(config *ApplianceConfig) error {
	for i, extension := range config.Extensions {
		if extension.Name == "" {
			return fmt.Errorf("extension[%d].name is required", i)
		}

		if extension.Version == "" {
			return fmt.Errorf("extension[%d].version is required", i)
		}

		if extension.BakeryUrl == "" {
			return fmt.Errorf("extension[%d].bakery_url is required", i)
		}
	}

	return nil
}

func validateContainers(config *ApplianceConfig) error {
	for i, container := range config.Containers {
		if container.Name == "" {
			return fmt.Errorf("container[%d].name is required", i)
		}

		if container.Image == "" {
			return fmt.Errorf("container[%d].image is required", i)
		}
	}

	return nil
}

func validateFiles(config *ApplianceConfig) error {
	for i, file := range config.Files {
		if file.Path == "" {
			return fmt.Errorf("file[%d].path is required", i)
		}

		if file.Mode == "" {
			return fmt.Errorf("file[%d].mode is required", i)
		}
	}

	return nil
}

func validateDirectories(config *ApplianceConfig) error {
	for i, directory := range config.Directories {
		if directory.Path == "" {
			return fmt.Errorf("directory[%d].path is required", i)
		}

		if directory.Mode == "" {
			return fmt.Errorf("directory[%d].mode is required", i)
		}
	}

	return nil
}

// func validateMounts(config *ApplianceConfig) error {
// 	for i, mount := range config.Mounts {
// 		if mount.Source == "" {
// 			return fmt.Errorf("mount[%d].source is required", i)
// 		}

// 		if mount.Target == "" {
// 			return fmt.Errorf("mount[%d].target is required", i)
// 		}
// 	}

// 	return nil
// }

func validateInterfaces(config *ApplianceConfig) error {
	for i, iface := range config.Interfaces {
		if iface.Name == "" && iface.MACAddress == "" {
			return fmt.Errorf("interface[%d].name or interface[%d].mac_address is required", i, i)
		}

		if iface.MACAddress != "" && iface.Name != "" {
			return fmt.Errorf("interface[%d].name and interface[%d].mac_address are mutually exclusive", i, i)
		}

		if iface.Address != "" && iface.DHCP {
			return fmt.Errorf("interface[%d].address and interface[%d].dhcp are mutually exclusive", i, i)
		}

		if len(iface.VLANs) == 0 && iface.Address == "" && !iface.DHCP {
			return fmt.Errorf("interface[%d].address or interface[%d].dhcp is required", i, i)
		}

		if iface.Address != "" && iface.Gateway == "" {
			return fmt.Errorf("interface[%d].gateway is required", i)
		}

		if iface.Address != "" && iface.DNS == "" {
			return fmt.Errorf("interface[%d].dns is required", i)
		}

		seenNames := make(map[string]struct{})
		seenIDs := make(map[int]struct{})
		for j, vlan := range iface.VLANs {
			_, seenName := seenNames[vlan.Name]
			if seenName {
				return fmt.Errorf("interface[%d].vlan[%d].name is not unique", i, j)
			}

			seenNames[vlan.Name] = struct{}{}

			_, seenID := seenIDs[vlan.ID]
			if seenID {
				return fmt.Errorf("interface[%d].vlan[%d].id is not unique", i, j)
			}

			seenIDs[vlan.ID] = struct{}{}

			if err := validateVLAN(&vlan); err != nil {
				return fmt.Errorf("interface[%d].vlan[%d]: %w", i, j, err)
			}
		}
	}

	return nil
}

func validateVLAN(vlan *VLAN) error {
	if vlan.Address != "" && vlan.DHCP {
		return fmt.Errorf("vlan.address and vlan.dhcp are mutually exclusive")
	}

	if vlan.Address == "" && !vlan.DHCP {
		return fmt.Errorf("vlan.address or vlan.dhcp is required")
	}

	if vlan.Address != "" && vlan.Gateway == "" {
		return fmt.Errorf("vlan.gateway is required")
	}

	return nil
}
