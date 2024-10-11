package ignition

import (
	"encoding/json"
	"errors"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

var (
	ErrDuplicateFile      = errors.New("duplicate file")
	ErrDuplicateDirectory = errors.New("duplicate directory")
	ErrDuplicateUnit      = errors.New("duplicate unit")
	ErrDuplicateUser      = errors.New("duplicate user")
)

type GeneratorOpt func(*generator)

func WithBundledExtensions() GeneratorOpt {
	return func(o *generator) {
		o.BundledExtensions = true
	}
}

func Generate(cfg *config.ApplianceConfig, opts ...GeneratorOpt) ([]byte, error) {
	gen := new(generator)
	for _, opt := range opts {
		opt(gen)
	}

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
	BundledExtensions bool
	Users             []ignitionTypes.PasswdUser
	Files             []ignitionTypes.File
	Links             []ignitionTypes.Link
	Directories       []ignitionTypes.Directory
	Units             []ignitionTypes.Unit
}

func (g *generator) Ignition(cfg *config.ApplianceConfig) (*ignitionTypes.Config, error) {
	err := g.generate(cfg)
	if err != nil {
		return nil, err
	}

	err = g.validate()
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

func (g *generator) generate(cfg *config.ApplianceConfig) error {
	gens := []ignitionGenerator{
		generateUsers,
		generateContainers,
		generateExtensions,
		generateInterfaces,
		generateFiles,
		generateDirectories,
	}

	for _, gen := range gens {
		err := gen(cfg, g)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) validate() error {
	validators := []ignitionValidator{
		validateUsers,
		validateFiles,
		validateDirectories,
		validateUnits,
	}

	for _, validator := range validators {
		if err := validator(g); err != nil {
			return err
		}
	}

	return nil
}

type ignitionGenerator func(*config.ApplianceConfig, *generator) error

type ignitionValidator func(*generator) error

func toPtr[T any](v T) *T {
	return &v
}
