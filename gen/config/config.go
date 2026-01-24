package config

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/mr-linch/go-tg/gen/ir"
	"gopkg.in/yaml.v3"
)

//go:embed enums.yaml
var enumsYAML []byte

// EnumDef defines an enum via field references or an expr expression.
type EnumDef struct {
	Name   string   `yaml:"name"`
	Fields []string `yaml:"fields,omitempty"` // "TypeName.field_name" references
	Expr   string   `yaml:"expr,omitempty"`   // expr expression
}

// Config holds enum definitions.
type Config struct {
	Enums []EnumDef `yaml:"enums"`
}

// Load parses the embedded enums.yaml config.
func Load() (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(enumsYAML, &cfg); err != nil {
		return nil, fmt.Errorf("parse enums.yaml: %w", err)
	}
	return &cfg, nil
}

// ApplyEnums evaluates enum definitions against the parsed API
// and appends the resulting Enums to api.Enums.
func (c *Config) ApplyEnums(api *ir.API) error {
	env := &exprEnv{api: api}

	for _, def := range c.Enums {
		e := ir.Enum{
			Name:   def.Name,
			Fields: def.Fields,
		}

		switch {
		case def.Expr != "":
			values, err := evalExpr(def.Expr, env)
			if err != nil {
				return fmt.Errorf("enum %q: %w", def.Name, err)
			}
			e.Values = values

		case len(def.Fields) > 0:
			values := resolveFieldEnum(api, def.Fields[0])
			e.Values = values

		default:
			return fmt.Errorf("enum %q: must have either expr or fields", def.Name)
		}

		api.Enums = append(api.Enums, e)
	}

	return nil
}

// resolveFieldEnum finds the Enum values on the specified "TypeName.field_name" field.
func resolveFieldEnum(api *ir.API, ref string) []string {
	parts := strings.SplitN(ref, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	typeName, fieldName := parts[0], parts[1]

	for _, t := range api.Types {
		if t.Name == typeName {
			for _, f := range t.Fields {
				if f.Name == fieldName {
					return f.Enum
				}
			}
		}
	}
	return nil
}

// evalExpr evaluates an expr expression and returns the result as []string.
func evalExpr(code string, env *exprEnv) ([]string, error) {
	program, err := expr.Compile(code, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("run: %w", err)
	}

	return toStringSlice(output)
}

// toStringSlice converts an expression result to []string.
func toStringSlice(v any) ([]string, error) {
	switch val := v.(type) {
	case []string:
		return val, nil
	case []any:
		result := make([]string, len(val))
		for i, item := range val {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("element %d: expected string, got %T", i, item)
			}
			result[i] = s
		}
		return result, nil
	default:
		return nil, fmt.Errorf("expected []string, got %T", v)
	}
}

// exprEnv is the environment exposed to expr expressions.
type exprEnv struct {
	api *ir.API
}

// Fields returns fields of the named type.
func (e *exprEnv) Fields(typeName string) []ir.Field {
	for _, t := range e.api.Types {
		if t.Name == typeName {
			return t.Fields
		}
	}
	return nil
}

// Subtypes returns subtype names of a union type.
func (e *exprEnv) Subtypes(typeName string) []string {
	for _, t := range e.api.Types {
		if t.Name == typeName {
			return t.Subtypes
		}
	}
	return nil
}

// SubtypeConsts collects Const values of fieldName from all subtypes of unionName.
func (e *exprEnv) SubtypeConsts(unionName, fieldName string) []string {
	subtypes := e.Subtypes(unionName)
	if len(subtypes) == 0 {
		return nil
	}

	var values []string
	for _, stName := range subtypes {
		for _, t := range e.api.Types {
			if t.Name == stName {
				for _, f := range t.Fields {
					if f.Name == fieldName && f.Const != "" {
						values = append(values, f.Const)
					}
				}
				break
			}
		}
	}
	return values
}
