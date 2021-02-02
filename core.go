package core

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

// Parse constructs OPA policy from string
func Parse(content, ruleQuery string) (Policy, error) {
	// validate module
	module, err := ast.ParseModule("", content)
	if err != nil {
		return Policy{}, err
	}

	if module == nil {
		return Policy{}, fmt.Errorf("Failed to parse module: empty content")
	}

	var valid bool
	for _, rule := range module.Rules {
		if rule.Head.Name == ast.Var(ruleQuery) {
			valid = true
			break
		}
	}

	if !valid {
		return Policy{}, fmt.Errorf("rule `%s` is not found", ruleQuery)
	}

	policy := Policy{
		module: module,
		pkg:    strings.Split(module.Package.String(), "package ")[1],
	}

	return policy, nil
}

// Eval validates data against given policy
// returns error if there're any violations found
func (p Policy) Eval(data interface{}, query string) error {
	rego := rego.New(
		rego.Query(fmt.Sprintf("data.%s.%s", p.pkg, query)),
		rego.ParsedModule(p.module),
		rego.Input(data),
	)

	// Run evaluation.
	rs, err := rego.Eval(context.Background())
	if err != nil {
		return err
	}
	for _, r := range rs {
		for _, expr := range r.Expressions {
			switch reflect.TypeOf(expr.Value).Kind() {
			// FIXME: support more formats
			case reflect.Slice:
				s := expr.Value.([]interface{})
				// FIXME: return multiple violations if found
				for i := 0; i < len(s); i++ {
					err := NoValidError{
						Details: s[i],
					}
					return err
				}
			case reflect.Map:
				s := expr.Value.(map[string]interface{})
				err := NoValidError{
					Details: s,
				}
				return err
			}
		}
	}

	return nil
}
