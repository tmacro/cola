package ignition

import (
	"encoding/json"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

func Generate(cfg *config.ApplianceConfig) ([]byte, error) {
	gen := new(generator)

	ignCfg, err := gen.Ignition(cfg)
	if err != nil {
		return nil, err
	}

	ignJson, err := json.MarshalIndent(&ignCfg, "", "  ")
	if err != nil {
		return nil, err
	}

	return ignJson, nil
}

type generator struct {
	Users []ignitionTypes.PasswdUser
	Files []ignitionTypes.File
	Links []ignitionTypes.Link
	Units []ignitionTypes.Unit
}

func (g *generator) Ignition(cfg *config.ApplianceConfig) (*ignitionTypes.Config, error) {
	return g.generate(cfg)
}

func (g *generator) generate(cfg *config.ApplianceConfig) (*ignitionTypes.Config, error) {
	gens := []ignitionGenerator{
		generateUsers,
		generateContainers,
		generateExtensions,
		generateInterfaces,
		generateFiles,
		generateDirectories,
		// generateLinks,
	}

	for _, gen := range gens {
		err := gen(cfg, g)
		if err != nil {
			return nil, err
		}
	}

	err := g.validate()
	if err != nil {
		return nil, err
	}

	ignCfg := defaultConfig

	ignCfg.Passwd.Users = g.Users
	ignCfg.Storage.Files = g.Files
	ignCfg.Storage.Links = g.Links
	ignCfg.Systemd.Units = g.Units

	return &ignCfg, nil
}

func (g *generator) validate() error {
	return nil
}

type ignitionGenerator func(*config.ApplianceConfig, *generator) error

func toPtr[T any](v T) *T {
	return &v
}
