package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	"github.com/zclconf/go-cty/cty/function"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type Variable struct {
	Name        string
	Type        cty.Type
	Description string
	Sensitive   bool

	DefaultExpr hcl.Expression
	HasDefault  bool

	From hcl.Range
}

type Source string

// Source indicates where an input value came from.
const (
	SourceCLI     Source = "cli"
	SourceVarFile Source = "varfile"
	SourceDefault Source = "default"
)

// InputValue is a value provided by the user (cli/env/varfile).
type InputValue struct {
	Value  cty.Value
	Source Source
}

// LoadVariables parses variable blocks from an HCL file.
func LoadVariables(filepaths []string) (map[string]Variable, error) {
	// func LoadVariables(configPaths, valuePaths []string) (*hcl.EvalContext, error) {
	variables := make(map[string]Variable)
	for _, cPath := range filepaths {
		fileVars, diags, err := ParseVariableDeclarations(cPath)
		if diags.HasErrors() {
			return nil, diags
		}

		if err != nil {
			return nil, err
		}

		for k, v := range fileVars {
			_, ok := variables[k]
			if ok {
				return nil, fmt.Errorf("duplicate variable declaration: %s", k)
			}
			variables[k] = v
		}
	}

	return variables, nil
}

// ParseVarDeclsFromFile parses variable blocks from an HCL file.
//
// Supported syntax:
//
//	variable "name" {
//	  type        = string
//	  description = "desc"
//	  default     = "x"
//	  sensitive   = true
//	}
//
// type can be primitives or calls like list(string), object({ ... }), tuple([ ... ]).
func ParseVariableDeclarations(filename string) (map[string]Variable, hcl.Diagnostics, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags, fmt.Errorf("parse HCL: %s", filename)
	}

	body := f.Body

	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "system"},
			{Type: "etcd"},
			{Type: "user", LabelNames: []string{"username"}},
			{Type: "extension", LabelNames: []string{"name"}},
			{Type: "container", LabelNames: []string{"name"}},
			{Type: "file", LabelNames: []string{"path"}},
			{Type: "directory", LabelNames: []string{"path"}},
			{Type: "symlink", LabelNames: []string{"path"}},
			{Type: "mount", LabelNames: []string{"path"}},
			{Type: "interface", LabelNames: []string{"name"}},
			{Type: "service", LabelNames: []string{"path"}},
			{Type: "variable", LabelNames: []string{"name"}},
		},
	}

	content, moreDiags := body.Content(schema)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags, fmt.Errorf("decode variable blocks: %s", filename)
	}

	out := map[string]Variable{}

	for _, block := range content.Blocks {
		if block.Type != "variable" {
			continue
		}

		name := block.Labels[0]
		decl := Variable{
			Name: name,
			From: block.DefRange,
		}

		innerSchema := &hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "type", Required: false},
				{Name: "description", Required: false},
				{Name: "default", Required: false},
				{Name: "sensitive", Required: false},
			},
		}
		inner, innerDiags := block.Body.Content(innerSchema)
		diags = append(diags, innerDiags...)
		if innerDiags.HasErrors() {
			continue
		}

		// type (required unless default exists; if neither, error)
		if attr, ok := inner.Attributes["type"]; ok {
			ty, tdiags := ParseTypeExpr(attr.Expr)
			diags = append(diags, tdiags...)
			if !tdiags.HasErrors() {
				decl.Type = ty
			}
		}

		// description (string)
		if attr, ok := inner.Attributes["description"]; ok {
			v, vdiags := attr.Expr.Value(&hcl.EvalContext{})
			diags = append(diags, vdiags...)
			if !vdiags.HasErrors() {
				if v.Type() != cty.String {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid variable description",
						Detail:   "description must be a string",
						Subject:  &attr.Range,
					})
				} else {
					decl.Description = v.AsString()
				}
			}
		}

		// sensitive (bool)
		if attr, ok := inner.Attributes["sensitive"]; ok {
			v, vdiags := attr.Expr.Value(&hcl.EvalContext{})
			diags = append(diags, vdiags...)
			if !vdiags.HasErrors() {
				if v.Type() != cty.Bool {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid variable sensitive flag",
						Detail:   "sensitive must be a bool",
						Subject:  &attr.Range,
					})
				} else {
					decl.Sensitive = v.True()
				}
			}
		}

		// default (expression)
		if attr, ok := inner.Attributes["default"]; ok {
			decl.DefaultExpr = attr.Expr
			decl.HasDefault = true
		}

		// If type missing: infer from default if present, else error.
		if decl.Type == (cty.Type{}) {
			if decl.HasDefault {
				v, vdiags := decl.DefaultExpr.Value(&hcl.EvalContext{})
				diags = append(diags, vdiags...)
				if !vdiags.HasErrors() {
					decl.Type = v.Type()
				}
			} else {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Missing variable type",
					Detail:   fmt.Sprintf(`variable %q must declare "type" or set a "default"`, decl.Name),
					Subject:  &decl.From,
				})
			}
		}

		// Duplicate var declarations
		if _, exists := out[name]; exists {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate variable declaration",
				Detail:   fmt.Sprintf("variable %q declared more than once", name),
				Subject:  &decl.From,
			})
			continue
		}

		out[name] = decl
	}

	if diags.HasErrors() {
		return out, diags, errors.New("variable declaration errors")
	}

	return out, diags, nil
}

func LoadVarFiles(filenames []string) (map[string]InputValue, error) {
	values := make(map[string]InputValue)
	for _, vPath := range filenames {
		fileValues, diags, err := LoadVarFile(vPath)
		if err != nil {
			return nil, err
		}

		if diags.HasErrors() {
			return nil, diags
		}

		for k, v := range fileValues {
			ev, ok := values[k]
			if ok {
				return nil, fmt.Errorf("duplicate value definition: %s=%v", k, ev)
			}
			values[k] = v
		}
	}

	return values, nil
}

func MergeValues(valueMaps ...map[string]InputValue) map[string]InputValue {
	merged := map[string]InputValue{}
	for _, vm := range valueMaps {
		for k, v := range vm {
			merged[k] = v
		}
	}

	return merged
}

// LoadVarFile reads a tfvars-like HCL file consisting of top-level attributes:
//
//	datacenter =  "oak01"
//	subnets = ["a", "b"]
//
// Values are evaluated with an empty context (no var references).
func LoadVarFile(filename string) (map[string]InputValue, hcl.Diagnostics, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		return nil, diags, fmt.Errorf("parse var file: %s", filename)
	}

	// We accept any top-level attributes; ignore blocks.
	body := f.Body

	// Best effort: if the body is syntax, we can list attrs directly.
	var attrs hclsyntax.Attributes
	switch b := body.(type) {
	case *hclsyntax.Body:
		attrs = b.Attributes
	default:
		// Fall back to Content with empty schema: it will only return known attrs; so instead just try PartialContent
		// for any attributes by using the syntax body above. Most configs are hclsyntax, so this is OK.
	}

	out := map[string]InputValue{}
	if attrs == nil {
		return out, diags, nil
	}

	for name, attr := range attrs {
		v, vdiags := attr.Expr.Value(&hcl.EvalContext{})
		diags = append(diags, vdiags...)
		if vdiags.HasErrors() {
			continue
		}
		out[name] = InputValue{
			Value:  v,
			Source: SourceVarFile,
		}
	}

	if diags.HasErrors() {
		for _, err := range diags.Errs() {
			fmt.Println(err)
		}
		return out, diags, errors.New("var file errors")
	}
	return out, diags, nil
}

var NilInputValue = InputValue{}

func EvalVariableExpr(s string) (key string, value InputValue, diags hcl.Diagnostics, err error) {
	before, after, ok := strings.Cut(s, "=")
	if !ok {
		return "", NilInputValue, nil, fmt.Errorf("expected key=value, got %q", s)
	}
	key = strings.TrimSpace(before)
	raw := strings.TrimSpace(after)
	if key == "" {
		return "", NilInputValue, nil, fmt.Errorf("empty key in %q", s)
	}

	expr, pdiags := hclsyntax.ParseExpression([]byte(raw), "<var>", hcl.Pos{Line: 1, Column: 1})
	diags = append(diags, pdiags...)
	if pdiags.HasErrors() {
		return "", NilInputValue, diags, fmt.Errorf("parse expression for %q", key)
	}

	v, vdiags := expr.Value(&hcl.EvalContext{})
	diags = append(diags, vdiags...)
	if vdiags.HasErrors() {
		return "", NilInputValue, diags, fmt.Errorf("eval expression for %q", key)
	}

	return key, InputValue{Value: v, Source: SourceCLI}, diags, nil
}

func coerceValue(v cty.Value, want cty.Type, context string) (cty.Value, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// cty.NilVal shouldn't happen in practice; treat as error.
	if v == cty.NilVal {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid value",
			Detail:   fmt.Sprintf("%s: value is nil", context),
		})
		return cty.NilVal, diags
	}

	cv, err := convert.Convert(v, want)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid variable value",
			Detail:   fmt.Sprintf("%s: cannot convert %s to %s: %v", context, v.Type().FriendlyName(), want.FriendlyName(), err),
		})
		return cty.NilVal, diags
	}
	return cv, diags
}

// ResolveVars merges input values with declarations, applies defaults, and coerces to declared types.
func ResolveVars(decls map[string]Variable, inputs map[string]InputValue) (resolved map[string]cty.Value, diags hcl.Diagnostics, err error) {
	resolved = map[string]cty.Value{}

	// Helpful: detect unknown inputs (values provided but no decl).
	for name := range inputs {
		if _, ok := decls[name]; !ok {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Undeclared variable",
				Detail:   fmt.Sprintf("A value was provided for %q but no variable %q is declared.", name, name),
			})
		}
	}

	// Resolve declared vars.
	for name, decl := range decls {
		if in, ok := inputs[name]; ok {
			cv, cdiags := coerceValue(in.Value, decl.Type, fmt.Sprintf("variable %q from %s", name, in.Source))
			diags = append(diags, cdiags...)
			if !cdiags.HasErrors() {
				resolved[name] = cv
			}
			continue
		}

		if decl.HasDefault {
			dv, ddiags := decl.DefaultExpr.Value(&hcl.EvalContext{})
			diags = append(diags, ddiags...)
			if ddiags.HasErrors() {
				continue
			}
			cv, cdiags := coerceValue(dv, decl.Type, fmt.Sprintf("default for variable %q", name))
			diags = append(diags, cdiags...)
			if !cdiags.HasErrors() {
				resolved[name] = cv
			}
			continue
		}

		// Required
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Missing required variable",
			Detail:   fmt.Sprintf("No value for required variable %q was provided, and it has no default.", name),
			Subject:  &decl.From,
		})
	}

	if diags.HasErrors() {
		return resolved, diags, errors.New("variable resolution errors")
	}
	return resolved, diags, nil
}

// ParseTypeExpr converts an HCL expression into a cty.Type.
//
// Supported forms:
//
//	string | number | bool | any
//	list(T) | set(T) | map(T)
//	object({ a = string, b = list(number) })
//	tuple([string, number])
func ParseTypeExpr(expr hcl.Expression) (cty.Type, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// Keywords like: string, number, bool, any
	if kw := hcl.ExprAsKeyword(expr); kw != "" {
		switch kw {
		case "string":
			return cty.String, diags
		case "number":
			return cty.Number, diags
		case "bool":
			return cty.Bool, diags
		case "any":
			return cty.DynamicPseudoType, diags
		default:
			diags = append(diags, diagType(expr, fmt.Sprintf("unknown primitive type %q", kw)))
			return cty.DynamicPseudoType, diags
		}
	}

	// Function-style type constructors: list(string), map(number), object({...}), tuple([...]), optional(T)
	if call, ok := expr.(*hclsyntax.FunctionCallExpr); ok {
		name := call.Name
		args := call.Args

		oneArg := func() (hcl.Expression, bool) {
			if len(args) != 1 {
				diags = append(diags, diagType(expr, fmt.Sprintf("%s() expects exactly 1 argument", name)))
				return nil, false
			}
			return args[0], true
		}

		switch name {
		case "list":
			a, ok := oneArg()
			if !ok {
				return cty.DynamicPseudoType, diags
			}
			elem, d := ParseTypeExpr(a)
			diags = append(diags, d...)
			if d.HasErrors() {
				return cty.DynamicPseudoType, diags
			}
			return cty.List(elem), diags

		case "set":
			a, ok := oneArg()
			if !ok {
				return cty.DynamicPseudoType, diags
			}
			elem, d := ParseTypeExpr(a)
			diags = append(diags, d...)
			if d.HasErrors() {
				return cty.DynamicPseudoType, diags
			}
			return cty.Set(elem), diags

		case "map":
			a, ok := oneArg()
			if !ok {
				return cty.DynamicPseudoType, diags
			}
			elem, d := ParseTypeExpr(a)
			diags = append(diags, d...)
			if d.HasErrors() {
				return cty.DynamicPseudoType, diags
			}
			return cty.Map(elem), diags

		case "tuple":
			a, ok := oneArg()
			if !ok {
				return cty.DynamicPseudoType, diags
			}
			// tuple([T1, T2, ...])
			tup, ok := a.(*hclsyntax.TupleConsExpr)
			if !ok {
				diags = append(diags, diagType(a, "tuple() expects a tuple literal like tuple([string, number])"))
				return cty.DynamicPseudoType, diags
			}
			elems := make([]cty.Type, 0, len(tup.Exprs))
			for _, te := range tup.Exprs {
				et, d := ParseTypeExpr(te)
				diags = append(diags, d...)
				if d.HasErrors() {
					return cty.DynamicPseudoType, diags
				}
				elems = append(elems, et)
			}
			return cty.Tuple(elems), diags

		case "object":
			a, ok := oneArg()
			if !ok {
				return cty.DynamicPseudoType, diags
			}
			obj, ok := a.(*hclsyntax.ObjectConsExpr)
			if !ok {
				diags = append(diags, diagType(a, "object() expects an object literal like object({ a = string })"))
				return cty.DynamicPseudoType, diags
			}

			attrTypes := map[string]cty.Type{}
			for _, item := range obj.Items {
				// Keys should be simple identifiers or string literals.
				key, kdiag := objectKey(item.KeyExpr)
				diags = append(diags, kdiag...)
				if kdiag.HasErrors() {
					return cty.DynamicPseudoType, diags
				}

				t, d := ParseTypeExpr(item.ValueExpr)
				diags = append(diags, d...)
				if d.HasErrors() {
					return cty.DynamicPseudoType, diags
				}
				attrTypes[key] = t
			}
			return cty.Object(attrTypes), diags

		default:
			diags = append(diags, diagType(expr, fmt.Sprintf("unknown type constructor %q", name)))
			return cty.DynamicPseudoType, diags
		}
	}

	diags = append(diags, diagType(expr, "unsupported type expression (use primitives or list/map/set/object/tuple)"))
	return cty.DynamicPseudoType, diags
}

func objectKey(expr hcl.Expression) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	// { foo = ... } -> KeyExpr is *hclsyntax.ObjectConsKeyExpr wrapping traversal
	if kexpr, ok := expr.(*hclsyntax.ObjectConsKeyExpr); ok {
		// Try keyword identifier
		if kw := hcl.ExprAsKeyword(kexpr.Wrapped); kw != "" {
			return kw, diags
		}
		// Or string literal
		v, vdiags := kexpr.Wrapped.Value(&hcl.EvalContext{})
		diags = append(diags, vdiags...)
		if vdiags.HasErrors() {
			return "", diags
		}
		if v.Type() != cty.String {
			diags = append(diags, diagType(expr, "object key must be an identifier or string"))
			return "", diags
		}
		return v.AsString(), diags
	}

	// Fallback: evaluate and require string
	v, vdiags := expr.Value(&hcl.EvalContext{})
	diags = append(diags, vdiags...)
	if vdiags.HasErrors() {
		return "", diags
	}
	if v.Type() != cty.String {
		diags = append(diags, diagType(expr, "object key must be an identifier or string"))
		return "", diags
	}
	return v.AsString(), diags
}

func diagType(expr hcl.Expression, detail string) *hcl.Diagnostic {
	rng := expr.Range()
	return &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid type expression",
		Detail:   detail,
		Subject:  &rng,
	}
}

// MakeEvalContext returns an EvalContext with var.<name> available.
func MakeEvalContext(variables map[string]cty.Value) *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(variables),
		},
		Functions: map[string]function.Function{
			"to_json": function.New(&toJsonSpec),
		},
	}
}

var toJsonSpec = function.Spec{
	Description: "Encode parameter as a json string",
	Params: []function.Parameter{
		{Type: cty.DynamicPseudoType, AllowNull: true, AllowDynamicType: true},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: toJSONImpl,
}

func toJSONImpl(args []cty.Value, retType cty.Type) (cty.Value, error) {
	if len(args) != 1 {
		return cty.NilVal, fmt.Errorf("jsonencode expects exactly 1 argument, got %d", len(args))
	}
	if !retType.Equals(cty.String) {
		return cty.NilVal, fmt.Errorf("jsonencode return type must be string, got %s", retType.FriendlyName())
	}

	v := args[0]

	// If the value isn't known yet (depends on unknown vars), propagate unknown.
	if !v.IsKnown() {
		return cty.UnknownVal(cty.String), nil
	}

	// If the value is null, encode as JSON null.
	if v.IsNull() {
		return cty.StringVal("null"), nil
	}

	// Marshal to JSON bytes using cty's JSON marshaler (preserves types properly).
	b, err := ctyjson.Marshal(v, v.Type())
	if err != nil {
		return cty.NilVal, fmt.Errorf("jsonencode: marshal failed: %w", err)
	}

	// Optional: ensure it's valid JSON (ctyjson.Marshal should already be, but harmless)
	if !json.Valid(b) {
		return cty.NilVal, fmt.Errorf("jsonencode: produced invalid json")
	}

	return cty.StringVal(string(b)), nil
}
