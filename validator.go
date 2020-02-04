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

// Policy contains policy and metedata
type Policy struct {
	name    string
	query   string
	content *ast.Module
}

// NewPolicy constructs OPA policy from the given rule bodies
func NewPolicy(name, query string, rules []string) (Policy, error) {
	// generate an empty module with package
	template := fmt.Sprintf(`
package %s

`, name)
	module := ast.MustParseModule(template)

	for _, rule := range rules {
		parsedRule, err := ast.ParseRule(rule)
		if err != nil {
			return Policy{}, fmt.Errorf("Invalid rule syntax %w", err)
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
		return Policy{}, c.Errors
	}

	return Policy{
		name:    name,
		query:   query,
		content: module,
	}, nil
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
		rego.Query(fmt.Sprintf("data.%s.%s", policy.name, policy.query)),
		rego.ParsedModule(policy.content),
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
