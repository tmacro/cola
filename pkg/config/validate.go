package config

import (
	"fmt"
	"slices"
	"strings"
)

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
	validateSystem,
	validateUsers,
	validateExtensions,
	validateContainers,
	validateFiles,
	validateDirectories,
	// validateMounts,
	validateInterfaces,
	validateServices,
	validateUpdate,
}

// etcd-lock    Reboot after first taking a distributed lock in etcd (reboot window applies)
// reboot       Reboot immediately after an update is applied (reboot window applies)
// off          Do not reboot after updates are applied
var validRebootStrategies = []string{"off", "reboot", "etcd-lock"}

// performance    Default. Operate at the maximum frequency
// ondemand       Dynamically scale frequency at 75% cpu load
// conservative   Dynamically scale frequency at 95% cpu load
// powersave      Operate at the minimum frequency
// userspace      Controlled by a userspace application via the `scaling_setspeed` file
var validPowerProfiles = []string{"performance", "ondemand", "conservative", "powersave", "userspace"}

func validateSystem(config *ApplianceConfig) error {
	if config.System.Updates != nil {
		valid := slices.Contains(validRebootStrategies, config.System.Updates.RebootStrategy)
		if !valid {
			return fmt.Errorf("system.updates.reboot_strategy must be one of: %s", strings.Join(validRebootStrategies, ", "))
		}
	}

	if config.System.PowerProfile != "" {
		valid := slices.Contains(validPowerProfiles, config.System.PowerProfile)
		if !valid {
			return fmt.Errorf("system.power-profile must be one of: %s", strings.Join(validPowerProfiles, ", "))
		}
	}

	return nil
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

		if container.Restart != "" && container.Restart != "always" && container.Restart != "no" {
			return fmt.Errorf("container[%d].restart must be 'always' or 'no'", i)
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

		if (iface.Address != "" || len(iface.Addresses) > 0) && iface.DHCP {
			return fmt.Errorf("interface[%d].address and interface[%d].dhcp are mutually exclusive", i, i)
		}

		if iface.Address != "" && len(iface.Addresses) > 0 {
			return fmt.Errorf("interface[%d].address and interface[%d].addresses are mutually exclusive", i, i)
		}

		if len(iface.VLANs) == 0 && iface.Address == "" && len(iface.Addresses) == 0 && !iface.DHCP {
			return fmt.Errorf("interface[%d].address, interface[%d].addresses, or interface[%d].dhcp is required", i, i, i)
		}

		if iface.Address != "" && iface.Gateway == "" {
			return fmt.Errorf("interface[%d].gateway is required", i)
		}

		// if iface.Address != "" && iface.DNS == "" {
		// 	return fmt.Errorf("interface[%d].dns is required", i)
		// }

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

func validateServices(config *ApplianceConfig) error {
	for i, service := range config.Services {
		if service.Name == "" {
			return fmt.Errorf("service[%d].name is required", i)
		}

		if service.Inline == "" && service.SourcePath == "" && len(service.DropIns) == 0 && !service.Enabled {
			return fmt.Errorf("service[%d] must have either inline, source_path, dropins, or be enabled", i)
		}

		for j, dropin := range service.DropIns {
			if dropin.Name == "" {
				return fmt.Errorf("service[%d].dropin[%d].name is required", i, j)
			}

			if dropin.Inline == "" && dropin.SourcePath == "" {
				return fmt.Errorf("service[%d].dropin[%d] must have either inline or source_path", i, j)
			}
		}
	}

	return nil
}

func validateUpdate(config *ApplianceConfig) error {
	if config.System == nil {
		return nil
	}

	if config.System.Updates == nil {
		return nil
	}

	if config.System.Updates.RebootStrategy == "" {
		return fmt.Errorf("updates.reboot_strategy is required")
	}

	strat := config.System.Updates.RebootStrategy
	if strat != "off" && strat != "reboot" && strat != "etcd-lock" {
		return fmt.Errorf("updates.reboot_strategy must be 'off', 'reboot', or 'etcd-lock'")
	}

	return nil
}

func validateEtcd(config *ApplianceConfig) error {
	if config.Etcd == nil {
		return nil
	}

	if config.Etcd.Name == "" {
		return fmt.Errorf("etcd.name is required")
	}

	if config.Etcd.InitialToken == "" {
		return fmt.Errorf("etcd.initial_token is required")
	}

	if config.Etcd.ListenAddress == "" {
		return fmt.Errorf("etcd.listen_address is required")
	}

	if len(config.Etcd.Peers) == 0 {
		return fmt.Errorf("etcd.peers is required")
	}

	for i, peer := range config.Etcd.Peers {
		if peer.Name == "" {
			return fmt.Errorf("etcd.peer[%d].name is required", i)
		}

		if peer.Address == "" {
			return fmt.Errorf("etcd.peer[%d].address is required", i)
		}

		if peer.Port == 0 {
			return fmt.Errorf("etcd.peer[%d].port is required", i)
		}
	}

	return nil
}
