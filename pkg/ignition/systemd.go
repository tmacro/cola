package ignition

import (
	"fmt"
	"os"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

func validateUnits(g *generator) error {
	if !keysAreUnique(g.Units, func(u ignitionTypes.Unit) string { return u.Name }) {
		return ErrDuplicateUnit
	}

	for _, unit := range g.Units {
		if unit.Contents == nil && len(unit.Dropins) == 0 && unit.Enabled == nil {
			return fmt.Errorf("unit %s has no contents, dropins, or enabled flag", unit.Name)
		}

		if !keysAreUnique(unit.Dropins, func(d ignitionTypes.Dropin) string { return d.Name }) {
			return ErrDuplicateDropin
		}

		for _, dropin := range unit.Dropins {
			if dropin.Contents == nil {
				return fmt.Errorf("dropin %s has no contents", dropin.Name)
			}
		}
	}

	return nil
}

func generateServices(cfg *config.ApplianceConfig, g *generator) error {
	fmt.Printf("%+v\n", cfg.Services)
	for _, service := range cfg.Services {
		unit := ignitionTypes.Unit{
			Name: service.Name,
		}

		if service.Enabled {
			unit.Enabled = toPtr(true)
		}

		if service.Inline != "" {
			unit.Contents = toPtr(service.Inline)
		} else if service.SourcePath != "" {
			contents, err := os.ReadFile(service.SourcePath)
			if err != nil {
				return fmt.Errorf("failed to read service file %s: %v", service.SourcePath, err)
			}

			unit.Contents = toPtr(string(contents))
		}

		for _, dropin := range service.DropIns {
			dropinUnit := ignitionTypes.Dropin{
				Name: dropin.Name,
			}

			if dropin.Inline != "" {
				dropinUnit.Contents = toPtr(dropin.Inline)
			} else if dropin.SourcePath != "" {
				contents, err := os.ReadFile(dropin.SourcePath)
				if err != nil {
					return fmt.Errorf("failed to read dropin file %s: %v", dropin.SourcePath, err)
				}

				dropinUnit.Contents = toPtr(string(contents))
			}

			unit.Dropins = append(unit.Dropins, dropinUnit)
		}

		g.Units = append(g.Units, unit)
	}

	return nil
}
