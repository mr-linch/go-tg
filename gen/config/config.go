package config

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/expr-lang/expr"
	"gopkg.in/yaml.v3"

	"github.com/mr-linch/go-tg/gen/ir"
)

// EnumDef defines an enum via field references or an expr expression.
type EnumDef struct {
	Name    string   `yaml:"name"`
	Fields  []string `yaml:"fields,omitempty"`  // "TypeName.field_name" references
	Expr    string   `yaml:"expr,omitempty"`    // expr expression
	Exclude []string `yaml:"exclude,omitempty"` // values to exclude from expr/fields results
}

// Parser holds parser-related configuration.
type Parser struct {
	Enums []EnumDef `yaml:"enums"`
}

// FieldTypeRule defines an expr-based predicate for field type mapping.
type FieldTypeRule struct {
	Expr   string `yaml:"expr"`             // boolean expr predicate
	Type   string `yaml:"type"`             // Go type to use when matched
	Scalar bool   `yaml:"scalar,omitempty"` // if true, no pointer wrapping for optional fields
}

// EnumGenDef defines how to generate a standalone enum type.
type EnumGenDef struct {
	Name       string `yaml:"name"`                 // Enum name from parser.enums (e.g., "ChatType")
	Underlying string `yaml:"underlying,omitempty"` // Go underlying type (default: "int")
	Marshal    string `yaml:"marshal,omitempty"`    // "json", "text", "stringer", or "" (default: "text")
	Unknown    bool   `yaml:"unknown,omitempty"`    // If true, include Unknown sentinel at 0
}

// TypeMethodDef defines a method to generate on a type that returns an enum value.
type TypeMethodDef struct {
	Type   string `yaml:"type"`   // Type name (e.g., "Message")
	Method string `yaml:"method"` // Method name (e.g., "Type")
	Return string `yaml:"return"` // Return enum type name (e.g., "MessageType")
}

// VariantConstructorDef defines variant constructors for types with mutually exclusive optional fields.
// Example: InlineKeyboardButton has text (required) + one of: url, callback_data, web_app, etc.
type VariantConstructorDef struct {
	Type        string            `yaml:"type"`                   // Type name (e.g., "InlineKeyboardButton")
	BaseField   string            `yaml:"base_field"`             // Required field name (e.g., "text")
	IncludeBase bool              `yaml:"include_base,omitempty"` // Generate base constructor (e.g., NewKeyboardButton)
	Exclude     []string          `yaml:"exclude,omitempty"`      // Fields to exclude from variant generation
	Defaults    map[string]string `yaml:"defaults,omitempty"`     // Default values for bool/pointer fields (field_name -> "true" or "&Type{}")
}

// InterfaceUnionDef defines a marker interface union for types without a JSON discriminator.
// These are union types where variants implement a marker interface (e.g., ReplyMarkup, InputMessageContent).
type InterfaceUnionDef struct {
	Name     string   `yaml:"name"`     // Interface name (e.g., "ReplyMarkup")
	Variants []string `yaml:"variants"` // Variant type names (e.g., ["InlineKeyboardMarkup", ...])
}

// TypeGen holds type generation configuration.
type TypeGen struct {
	Exclude             []string                `yaml:"exclude"`
	NameOverrides       map[string]string       `yaml:"name_overrides"`
	TypeOverrides       map[string]string       `yaml:"type_overrides"`
	FieldTypeRules      []FieldTypeRule         `yaml:"field_type_rules"`
	Enums               []EnumGenDef            `yaml:"enums,omitempty"`
	TypeMethods         []TypeMethodDef         `yaml:"type_methods,omitempty"`
	VariantConstructors []VariantConstructorDef `yaml:"variant_constructors,omitempty"`
	InterfaceUnions     []InterfaceUnionDef     `yaml:"interface_unions,omitempty"`
}

// IsExcluded reports whether typeName should be skipped.
func (tg *TypeGen) IsExcluded(typeName string) bool {
	return slices.Contains(tg.Exclude, typeName)
}

// ParamTypeRule defines an expr-based predicate for method parameter type mapping.
type ParamTypeRule struct {
	Expr string `yaml:"expr"` // boolean expr predicate
	Type string `yaml:"type"` // Go type to use when matched
}

// MethodGen holds method generation configuration.
// ConstructorVariant defines an alternative constructor for a method.
type ConstructorVariant struct {
	Suffix         string   `yaml:"suffix"`          // e.g., "Inline" -> NewEditMessageTextInlineCall
	RequiredParams []string `yaml:"required_params"` // param names to use as required
	ReturnType     string   `yaml:"return_type"`     // override return type for this variant (empty = use method default)
}

type MethodGen struct {
	ParamTypeRules      []ParamTypeRule                 `yaml:"param_type_rules"`
	ParamTypeOverrides  map[string]string               `yaml:"param_type_overrides"`  // "methodName.param_name" -> "GoType"
	ReturnTypeOverrides map[string]string               `yaml:"return_type_overrides"` // "methodName" -> "GoType"
	StringerTypes       []string                        `yaml:"stringer_types"`        // types that use request.Stringer()
	ConstructorVariants map[string][]ConstructorVariant `yaml:"constructor_variants"`  // method -> variants
}

// Config holds the unified go-tg-gen configuration.
type Config struct {
	Parser    Parser    `yaml:"parser"`
	TypeGen   TypeGen   `yaml:"typegen"`
	MethodGen MethodGen `yaml:"methodgen"`
}

// LoadFile loads configuration from the given YAML file path.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// ApplyEnums evaluates enum definitions against the parsed API
// and appends the resulting Enums to api.Enums.
func (c *Config) ApplyEnums(api *ir.API) error {
	env := &exprEnv{api: api}

	for _, def := range c.Parser.Enums {
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
			values, err := resolveFieldEnums(api, def.Fields)
			if err != nil {
				return fmt.Errorf("enum %q: %w", def.Name, err)
			}
			e.Values = values

		default:
			return fmt.Errorf("enum %q: must have either expr or fields", def.Name)
		}

		if len(def.Exclude) > 0 {
			excludeSet := make(map[string]struct{}, len(def.Exclude))
			for _, ex := range def.Exclude {
				excludeSet[ex] = struct{}{}
			}

			matched := make(map[string]struct{})
			e.Values = slices.DeleteFunc(e.Values, func(v string) bool {
				if _, ok := excludeSet[v]; ok {
					matched[v] = struct{}{}
					return true
				}
				return false
			})

			if len(matched) != len(excludeSet) {
				var unmatched []string
				for _, ex := range def.Exclude {
					if _, ok := matched[ex]; !ok {
						unmatched = append(unmatched, ex)
					}
				}
				slog.Warn("enum exclude entries did not match any values",
					"enum", def.Name,
					"unmatched", unmatched,
				)
			}
		}

		if len(e.Values) == 0 {
			return fmt.Errorf("enum %q: produced no values", def.Name)
		}

		api.Enums = append(api.Enums, e)
	}

	return nil
}

// resolveFieldEnums merges Enum values from all specified field references.
// It collects unique values while preserving order from each field.
func resolveFieldEnums(api *ir.API, refs []string) ([]string, error) {
	seen := make(map[string]bool)
	var result []string

	for _, ref := range refs {
		values, err := resolveFieldEnum(api, ref)
		if err != nil {
			return nil, err
		}
		for _, v := range values {
			if !seen[v] {
				seen[v] = true
				result = append(result, v)
			}
		}
	}
	return result, nil
}

// resolveFieldEnum finds the Enum values on the specified "TypeName.field_name" field.
func resolveFieldEnum(api *ir.API, ref string) ([]string, error) {
	parts := strings.SplitN(ref, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid field reference %q: expected TypeName.field_name", ref)
	}
	typeName, fieldName := parts[0], parts[1]

	for _, t := range api.Types {
		if t.Name == typeName {
			for _, f := range t.Fields {
				if f.Name == fieldName {
					return f.Enum, nil
				}
			}
			return nil, fmt.Errorf("field %q not found in type %q", fieldName, typeName)
		}
	}
	return nil, fmt.Errorf("type %q not found in API", typeName)
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

// ParamEnumValues returns enum values from a method parameter's <em> tags.
func (e *exprEnv) ParamEnumValues(methodName, paramName string) []string {
	for _, m := range e.api.Methods {
		if m.Name == methodName {
			for _, p := range m.Params {
				if p.Name == paramName {
					return p.Enum
				}
			}
		}
	}
	return nil
}
