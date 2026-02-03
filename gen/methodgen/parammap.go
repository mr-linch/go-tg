package methodgen

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/mr-linch/go-tg/gen/config"
	"github.com/mr-linch/go-tg/gen/ir"
)

// paramRuleEnv is the environment exposed to param type rule expressions.
type paramRuleEnv struct {
	ir.Param
	MethodName string
}

// CompiledParamTypeRules holds pre-compiled expr programs for param type matching.
type CompiledParamTypeRules struct {
	rules    []config.ParamTypeRule
	programs []*vm.Program
	matched  []bool
}

// CompileParamTypeRules compiles all expr-based param type rules.
func CompileParamTypeRules(rules []config.ParamTypeRule) (*CompiledParamTypeRules, error) {
	c := &CompiledParamTypeRules{
		rules:   rules,
		matched: make([]bool, len(rules)),
	}
	for _, r := range rules {
		prog, err := expr.Compile(r.Expr, expr.Env(paramRuleEnv{}), expr.AsBool())
		if err != nil {
			return nil, fmt.Errorf("compile param rule %q: %w", r.Expr, err)
		}
		c.programs = append(c.programs, prog)
	}
	return c, nil
}

// Match evaluates rules in order and returns the first matching type.
func (c *CompiledParamTypeRules) Match(methodName string, p ir.Param) (goType string, ok bool) {
	if c == nil {
		return "", false
	}
	env := paramRuleEnv{Param: p, MethodName: methodName}
	for i, rule := range c.rules {
		result, err := expr.Run(c.programs[i], env)
		if err == nil {
			if b, matched := result.(bool); matched && b {
				c.matched[i] = true
				return rule.Type, true
			}
		}
	}
	return "", false
}

// Unmatched returns the expr strings of rules that never matched any param.
func (c *CompiledParamTypeRules) Unmatched() []string {
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
