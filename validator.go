package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

func moduleTemplate() *ast.Module {
	template := `
package magalix.Validator

# returns containers specs from the controller's full spec
containers = input.spec.template.spec.containers {
	s := {"StatefulSet" , "DaemonSet", "Deployment", "Job"}
    s[_] = input.kind
} else = input.spec.containers {
	input.kind == "Pod"
} else = input.spec.jobTemplate.spec.template.spec.containers {
	input.kind == "CronJob"
}

controller = input

`
	return ast.MustParseModule(template)
}

// Policy parsed policy
type Policy *ast.Module

// NewPolicy constructs OPA policy from the given rule bodies
func NewPolicy(rules ...string) (Policy, error) {
	module := moduleTemplate()
	for _, rule := range rules {
		parsedRule, err := ast.ParseRule(rule)
		if err != nil {
			return nil, fmt.Errorf("Invalid rule syntax %w", err)
		}
		parsedRule.Module = module
		module.Rules = append(module.Rules, parsedRule)
	}

	// validate module
	c := ast.NewCompiler()
	mods := map[string]*ast.Module{
		"": module,
	}
	if c.Compile(mods); c.Failed() {
		return nil, c.Errors
	}

	return module, nil
}

// OpaValidator validates data against given policy
// returns error if there're any violations found
func OpaValidator(data string, policy Policy) error {
	// Build inputs from data
	var input interface{}
	dataDecoder := json.NewDecoder(bytes.NewBufferString(data))
	dataDecoder.UseNumber()
	if err := dataDecoder.Decode(&input); err != nil {
		return err
	}

	rego := rego.New(
		rego.Query("data.magalix.Validator.violation"),
		rego.ParsedModule(policy),
		rego.Input(input),
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
					return fmt.Errorf("%v", s[i])
				}
			}
		}
	}

	return nil
}
