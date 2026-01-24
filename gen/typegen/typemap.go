package typegen

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
)

// primitiveGoType maps IR primitive type names to Go type strings.
var primitiveGoType = map[ir.PrimitiveType]string{
	ir.TypeInteger:   "int",
	ir.TypeInteger64: "int64",
	ir.TypeFloat:     "float64",
	ir.TypeString:    "string",
	ir.TypeBoolean:   "bool",
	ir.TypeTrue:      "bool",
}

// scalarTypes are Go types that don't need a pointer for optional fields.
var scalarTypes = map[string]bool{
	"int":      true,
	"int64":    true,
	"float64":  true,
	"string":   true,
	"bool":     true,
	"UnixTime": true,
}

// fieldRuleEnv is the environment exposed to field type rule expressions.
type fieldRuleEnv struct {
	ir.Field
	TypeName string
}

// CompiledFieldTypeRules holds pre-compiled expr programs for field type matching.
type CompiledFieldTypeRules struct {
	rules    []config.FieldTypeRule
	programs []*vm.Program
	matched  []bool
}

// CompileFieldTypeRules compiles all expr-based field type rules.
func CompileFieldTypeRules(rules []config.FieldTypeRule) (*CompiledFieldTypeRules, error) {
	c := &CompiledFieldTypeRules{
		rules:   rules,
		matched: make([]bool, len(rules)),
	}
	for _, r := range rules {
		prog, err := expr.Compile(r.Expr, expr.Env(fieldRuleEnv{}), expr.AsBool())
		if err != nil {
			return nil, fmt.Errorf("compile field rule %q: %w", r.Expr, err)
		}
		c.programs = append(c.programs, prog)
	}
	return c, nil
}

// Match evaluates rules in order and returns the first matching type and scalar flag.
func (c *CompiledFieldTypeRules) Match(typeName string, f ir.Field) (goType string, scalar bool, ok bool) {
	if c == nil {
		return "", false, false
	}
	env := fieldRuleEnv{Field: f, TypeName: typeName}
	for i, rule := range c.rules {
		result, err := expr.Run(c.programs[i], env)
		if err == nil {
			if b, matched := result.(bool); matched && b {
				c.matched[i] = true
				return rule.Type, rule.Scalar, true
			}
		}
	}
	return "", false, false
}

// Unmatched returns the expr strings of rules that never matched any field.
func (c *CompiledFieldTypeRules) Unmatched() []string {
	if c == nil {
		return nil
	}
	var result []string
	for i, rule := range c.rules {
		if !c.matched[i] {
			result = append(result, rule.Expr)
		}
	}
	return result
}

// resolveGoType resolves a field's TypeExpr to a Go type string.
func resolveGoType(typeName string, f ir.Field, cfg *config.TypeGen, rules *CompiledFieldTypeRules, usedTypeOverrides map[string]bool) string {
	key := typeName + "." + f.Name
	if override, ok := cfg.TypeOverrides[key]; ok {
		usedTypeOverrides[key] = true
		return override
	}

	if matched, scalar, ok := rules.Match(typeName, f); ok {
		if scalar {
			return matched
		}
		return applyOptionalAndArray(matched, f.TypeExpr.Array, f.Optional, false)
	}

	if len(f.TypeExpr.Types) == 0 {
		return "any"
	}

	baseType := resolveBaseType(f.TypeExpr.Types[0].Type)
	scalar := isScalar(baseType)

	return applyOptionalAndArray(baseType, f.TypeExpr.Array, f.Optional, scalar)
}

// resolveBaseType maps a single type name to its Go equivalent.
func resolveBaseType(name string) string {
	if goType, ok := primitiveGoType[ir.PrimitiveType(name)]; ok {
		return goType
	}
	return normalizeTypeName(name)
}

// isScalar reports whether the Go type is a scalar (no pointer needed for optional).
func isScalar(goType string) bool {
	return scalarTypes[goType]
}

// applyOptionalAndArray wraps the base type with array brackets and/or pointer.
func applyOptionalAndArray(base string, arrayDepth int, optional bool, scalar bool) string {
	if arrayDepth > 0 {
		result := base
		for range arrayDepth {
			result = "[]" + result
		}
		return result
	}
	if optional && !scalar {
		return "*" + base
	}
	return base
}
