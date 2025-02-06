package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/rs/zerolog/log"
	"github.com/zclconf/go-cty/cty"
)

var builtinVariables = map[string]cty.Value{
	VariableTypeString: cty.StringVal(VariableTypeString),
	VariableTypeBool:   cty.StringVal(VariableTypeBool),
	VariableTypeNumber: cty.StringVal(VariableTypeNumber),
}

func buildEvalContext(extraVars map[string]cty.Value) *hcl.EvalContext {
	variables := make(map[string]cty.Value)
	for k, v := range builtinVariables {
		variables[k] = v
	}

	if extraVars != nil {
		variables["var"] = cty.ObjectVal(extraVars)
	}

	return &hcl.EvalContext{
		Variables: variables,
	}
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

type VariablePartial struct {
	Variables []Variable `hcl:"variable,block"`
	Remain    hcl.Body   `hcl:",remain"`
}

type VariableFile struct {
	Body hcl.Body `hcl:",remain"`
}

func readVariableConfig(paths []string) (hcldec.ObjectSpec, map[string]cty.Value, error) {
	evalCtx := buildEvalContext(nil)

	variableSpecs := make(map[string]hcldec.Spec)
	defaultValues := make(map[string]cty.Value)

	for _, path := range paths {
		var partial VariablePartial

		err := hclsimple.DecodeFile(path, evalCtx, &partial)
		if err != nil {
			return nil, nil, ParseError{Err: err, Path: path}
		}

		for _, variable := range partial.Variables {
			if _, ok := variableSpecs[variable.Name]; ok {
				return nil, nil, fmt.Errorf("variable %q is defined more than once", variable.Name)
			}

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
				return nil, nil, fmt.Errorf("unsupported variable type: %s", variable.Type)
			}

			variableSpecs[variable.Name] = &hcldec.AttrSpec{
				Name:     variable.Name,
				Type:     varType,
				Required: false,
			}

			blockSpec := hcldec.ObjectSpec{
				"default": &hcldec.AttrSpec{
					Name:     "default",
					Type:     varType,
					Required: false,
				},
			}

			val, err := hcldec.Decode(variable.Remain, blockSpec, evalCtx)
			if err != nil {
				return nil, nil, err
			}

			v := val.GetAttr("default")
			if !v.IsNull() {
				defaultValues[variable.Name] = v
			}
		}
	}

	return hcldec.ObjectSpec(variableSpecs), defaultValues, nil
}

func readVariables(paths []string, spec hcldec.ObjectSpec, defaults map[string]cty.Value) (map[string]cty.Value, error) {
	variables := make(map[string]cty.Value)
	for _, valuePath := range paths {
		data, err := os.ReadFile(valuePath)
		if err != nil {
			return nil, err
		}

		var value VariableFile
		err = hclsimple.Decode("variable.hcl", data, nil, &value)
		if err != nil {
			return nil, ParseError{Err: err, Path: valuePath}
		}

		val, diags := hcldec.Decode(value.Body, spec, nil)
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

	missing := make([]string, 0)
	for k := range spec {
		_, ok := variables[k]
		if !ok {
			defaultVal, ok := defaults[k]
			if !ok {
				missing = append(missing, k)
				continue
			}

			variables[k] = defaultVal
			log.Debug().
				Str("name", k).
				Str("value", ctyValueToString(defaultVal)).
				Msg("Using default value for variable")

		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing values for required variables: %s", strings.Join(missing, ", "))
	}

	return variables, nil
}

func loadVariables(configPaths, valuePaths []string) (map[string]cty.Value, error) {
	variableSpec, defaultValues, err := readVariableConfig(configPaths)
	if err != nil {
		return nil, err
	}

	return readVariables(valuePaths, variableSpec, defaultValues)
}
