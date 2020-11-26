package core

import (
	"testing"
)

type testCaseParsePolicy struct {
	name     string
	content  string
	hasError bool
}

func TestParse(t *testing.T) {
	cases := []testCaseParsePolicy{
		{
			name: "single rule",
			content: `
		package core
		violations[issue] {
			issue = "test"
		}`},
		{
			name: "multiple rules at once",
			content: `
			package core
			violations[issue] {
				issue = "test"
			}
			violations[issue] {
				issue = "test"
			}
		`,
			hasError: false,
		},
		{
			name: "invalid syntax",
			content: `
			package core
			issue = "test issue")
		`,
			hasError: true,
		},
		{
			name: "invalid syntax",
			content: `
			package core
			issue = "test issue")
		`,
			hasError: true,
		},
		{
			name:     "empty content",
			content:  "",
			hasError: true,
		},
		{
			name: "no issue variable",
			content: `
			package core
			violations[issue] {
				x = 3
			}
		`,
			hasError: true,
		},
		{
			name: "policy without package",
			content: `
			violations[issue] {
				x = 3
			}
		`,
			hasError: true,
		},
		{
			name: "policy with runtime error for multiple specs or containers",
			content: `
			package magalix.advisor.image_pull

			violations[result] {
				not controller_spec.imagePullPolicy
				result = {
					"issue": true
				}
			}



			# controller_container acts as an iterator to get containers from the template
			controller_spec = input.spec.template.spec.containers[_] {
				contains_kind(input.kind, {"StatefulSet" , "DaemonSet", "Deployment", "Job"})
			} else = input.spec {
				input.kind == "Pod"
			} else = input.spec.jobTemplate.spec.template.spec {
				input.kind == "CronJob"
			}

			contains_kind(kind, kinds) {
			  kinds[_] = kind
			}
			`,
			hasError: true,
		},
	}

	for _, c := range cases {
		_, err := Parse(c.content)
		if c.hasError {
			if err == nil {
				t.Errorf("[%s]: passed but should have been failed", c.name)
			}
		} else {
			if err != nil {
				t.Errorf("[%s]: %v", c.name, err)
			}
		}
	}
}

type testCaseEval struct {
	name         string
	content      string
	violationMsg string
	hasViolation bool
}

func TestEval(t *testing.T) {
	cases := []testCaseEval{
		{
			name: "rule has no violations",
			content: `
			package core
			violations[issue] {
				1 == 2
				issue = "violation test"
			}`,
		},
		{
			name: "rule has an empty violations",
			content: `
			package core
			violations[issue] {
				issue = ""
			}`,
			violationMsg: "\"\"",
			hasViolation: true,
		},
		{
			name: "rule has a bool violations",
			content: `
			package core
			violations[issue] {
				issue = true
			}`,
			violationMsg: "true",
			hasViolation: true,
		},
	}

	for _, c := range cases {
		policy, err := Parse(c.content)
		err = policy.Eval("{}", "violations")

		if c.hasViolation {
			if err == nil {
				t.Errorf("[%s]: passed but should have been failed", c.name)
			} else if err.Error() != c.violationMsg {
				t.Errorf("[%s]: expected error msg '%s' but got %s", c.name, c.violationMsg, err)
			}
		} else {
			if err != nil {
				t.Errorf("[%s]: %v", c.name, err)
			}
		}
	}
}
