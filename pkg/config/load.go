package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/rs/zerolog/log"
	"github.com/zclconf/go-cty/cty"
)

var (
	ErrNoSystemBlock = fmt.Errorf("system block is required")
)

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

func getFilesInDir(path, ext string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0)

	for _, file := range files {
		fp := filepath.Join(path, file.Name())

		if file.IsDir() {
			resolved, err := resolveFilePath(fp, ext)
			if err != nil {
				return nil, err
			}

			paths = append(paths, resolved...)
		} else if filepath.Ext(fp) == ext {
			paths = append(paths, fp)
		}
	}

	return paths, nil
}

func resolveFilePath(path, ext string) ([]string, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		dirPaths, err := getFilesInDir(path, ext)
		if err != nil {
			return nil, err
		}
		return dirPaths, nil
	}

	return []string{path}, nil
}

func resolveFilePaths(paths []string, ext string) ([]string, error) {
	filePaths := make([]string, 0)
	for _, path := range paths {
		paths, err := resolveFilePath(path, ext)
		if err != nil {
			return nil, err
		}

		filePaths = append(filePaths, paths...)
	}

	return filePaths, nil
}

type VariablePartial struct {
	Variables []Variable `hcl:"variable,block"`
	Remain    hcl.Body   `hcl:",remain"`
}

type VariableFile struct {
	Body hcl.Body `hcl:",remain"`
}

func readVariableConfig(paths []string) ([]Variable, error) {
	vars := make([]Variable, 0)

	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			VariableTypeString: cty.StringVal(VariableTypeString),
			VariableTypeBool:   cty.StringVal(VariableTypeBool),
			VariableTypeNumber: cty.StringVal(VariableTypeNumber),
		},
	}

	for _, path := range paths {
		var partial VariablePartial
		err := hclsimple.DecodeFile(path, evalCtx, &partial)
		if err != nil {
			return nil, ParseError{Err: err, Path: path}
		}

		vars = append(vars, partial.Variables...)
	}

	return vars, nil
}

func ctyValueToString(v cty.Value) string {
	switch v.Type() {
	case cty.String:
		return v.AsString()
	case cty.Number:
		return v.AsBigFloat().String()
	case cty.Bool:
		return fmt.Sprintf("%t", v.True())
	default:
		return "<unknown>"
	}
}

func loadVariables(paths, values []string) (map[string]cty.Value, error) {
	vars, err := readVariableConfig(paths)
	if err != nil {
		return nil, err
	}

	variableSpec := hcldec.ObjectSpec{}
	for _, variable := range vars {
		var varType cty.Type
		switch variable.Type {
		case VariableTypeString:
			log.Debug().Str("name", variable.Name).Str("type", "string").Msg("Discovered variable")
			varType = cty.String
		case VariableTypeNumber:
			log.Debug().Str("name", variable.Name).Str("type", "number").Msg("Discovered variable")
			varType = cty.Number
		case VariableTypeBool:
			log.Debug().Str("name", variable.Name).Str("type", "boolean").Msg("Discovered variable")
			varType = cty.Bool
		default:
			return nil, fmt.Errorf("unsupported variable type: %s", variable.Type)
		}

		variableSpec[variable.Name] = &hcldec.AttrSpec{
			Name:     variable.Name,
			Type:     varType,
			Required: false,
		}
	}

	variables := make(map[string]cty.Value)
	for _, valuePath := range values {
		data, err := os.ReadFile(valuePath)
		if err != nil {
			return nil, err
		}

		var value VariableFile
		err = hclsimple.Decode("variable.hcl", data, nil, &value)
		if err != nil {
			return nil, ParseError{Err: err, Path: valuePath}
		}

		val, diags := hcldec.Decode(value.Body, variableSpec, nil)
		if diags.HasErrors() {
			return nil, diags
		}

		for k, v := range val.AsValueMap() {
			if v.IsNull() {
				continue
			}

			log.Debug().
				Str("name", k).
				Str("value", ctyValueToString(v)).
				Str("file", filepath.Base(valuePath)).
				Msg("Loaded value for variable")

			_, ok := variables[k]
			if ok {
				return nil, fmt.Errorf("variable %s already defined", k)
			}

			variables[k] = v
		}
	}

	for k := range variableSpec {
		_, ok := variables[k]
		if !ok {
			return nil, fmt.Errorf("variable %s requires a value", k)
		}
	}

	return variables, nil
}

func readConfigFile(path string, evalCtx *hcl.EvalContext) (*ApplianceConfig, error) {
	var config ApplianceConfig
	err := hclsimple.DecodeFile(path, evalCtx, &config)
	if err != nil {
		return nil, ParseError{Err: err, Path: path}
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

func ReadConfig(paths, values []string) (*ApplianceConfig, error) {
	configPaths, err := resolveFilePaths(paths, ".hcl")
	if err != nil {
		return nil, err
	}

	valuePaths, err := resolveFilePaths(values, ".cvars")
	if err != nil {
		return nil, err
	}

	variables, err := loadVariables(configPaths, valuePaths)
	if err != nil {
		return nil, err
	}

	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var":              cty.ObjectVal(variables),
			VariableTypeString: cty.StringVal(VariableTypeString),
			VariableTypeBool:   cty.StringVal(VariableTypeBool),
			VariableTypeNumber: cty.StringVal(VariableTypeNumber),
		},
	}

	var merged *ApplianceConfig
	for _, cPath := range configPaths {
		config, err := readConfigFile(cPath, evalCtx)
		if err != nil {
			return nil, err
		}

		merged = MergeConfigs(merged, config)
	}

	if merged.System == nil {
		return nil, ErrNoSystemBlock
	}

	return merged, nil
}
