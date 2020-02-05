package core

import "testing"

type testCaseParsePolicy struct {
	name     string
	content  string
	hasError bool
}

func TestNewPolicy(t *testing.T) {
	cases := []testCaseParsePolicy{
		testCaseParsePolicy{
			name: "single rule",
			content: `
		package core
		violations[issue] {
			issue = "test"
		}`},
		testCaseParsePolicy{
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
		testCaseParsePolicy{
			name: "invalid syntax",
			content: `
			package core
			issue = "test issue")
		`,
			hasError: true,
		},
		testCaseParsePolicy{
			name: "invalid syntax",
			content: `
			package core
			issue = "test issue")
		`,
			hasError: true,
		},
		testCaseParsePolicy{
			name:     "empty content",
			content:  "",
			hasError: true,
		},
		testCaseParsePolicy{
			name: "no issue variable provided",
			content: `
			violations[issue] {
				x = 3
			}
		`,
			hasError: true,
		},
	}

	for _, c := range cases {
		_, err := Parse("magalix.advisor", c.content)
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

type testCaseOpaValidator struct {
	name         string
	content      string
	violationMsg string
	hasViolation bool
}

func TestOpaValidator(t *testing.T) {
	cases := []testCaseOpaValidator{
		testCaseOpaValidator{
			name: "rule has a string violation",
			content: `
			package core
			violations[issue] {
				issue = "violation test"
			}`,
			violationMsg: "violation test",
			hasViolation: true,
		},
		testCaseOpaValidator{
			name: "rule has no violations",
			content: `
			package core
			violations[issue] {
				1 == 2
				issue = "violation test"
			}`,
		},
		testCaseOpaValidator{
			name: "rule has an empty violations",
			content: `
			package core
			violations[issue] {
				issue = ""
			}`,
			violationMsg: "",
			hasViolation: true,
		},
		testCaseOpaValidator{
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
		policy, _ := Parse("magalix.advisor", c.content)
		err := policy.Eval("{}", "violations")
		if c.hasViolation {
			if err == nil {
				t.Errorf("[%s]: passed but should have been failed", c.name)
			} else if err.Error() != c.violationMsg {
				t.Errorf("[%s]: expected error msg %s but got %s", c.name, c.violationMsg, err)
			}
		} else {
			if err != nil {
				t.Errorf("[%s]: %v", c.name, err)
			}
		}
	}
}
