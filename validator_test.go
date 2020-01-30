package engine

import "testing"

type testCaseNewPolicy struct {
	name     string
	rules    []string
	hasError bool
}

func TestNewPolicy(t *testing.T) {
	cases := []testCaseNewPolicy{
		testCaseNewPolicy{
			name: "single rule",
			rules: []string{`
		violations[issue] {
			issue = "test"
		}
		`},
		},
		testCaseNewPolicy{
			name: "invalid syntax",
			rules: []string{`
			issue = "test issue")
		`},
			hasError: true,
		},
		testCaseNewPolicy{
			name: "multiple rules at once",
			rules: []string{`
			violations[issue] {
				issue = "test"
			}
			violations[issue] {
				issue = "test"
			}
		`},
			hasError: true,
		},
		testCaseNewPolicy{
			name: "invalid syntax",
			rules: []string{`
			issue = "test issue")
		`},
			hasError: true,
		},
		testCaseNewPolicy{
			name: "multiple rules at once",
			rules: []string{`

		`},
			hasError: true,
		},
		testCaseNewPolicy{
			name: "empty rule",
			rules: []string{`
			violations[issue] {
			}
		`},
			hasError: true,
		},
		testCaseNewPolicy{
			name: "no issue variable provided",
			rules: []string{`
			violations[issue] {
				x = 3
			}
		`},
			hasError: true,
		},
	}

	for _, c := range cases {
		_, err := NewPolicy(c.rules...)
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
	rule         string
	violationMsg string
	hasViolation bool
}

func TestOpaValidator(t *testing.T) {
	cases := []testCaseOpaValidator{
		testCaseOpaValidator{
			name: "rule has a string violation",
			rule: `
			violation[issue] {
				issue = "violation test"
			}`,
			violationMsg: "violation test",
			hasViolation: true,
		},
		testCaseOpaValidator{
			name: "rule has no violations",
			rule: `
			violation[issue] {
				1 == 2
				issue = "violation test"
			}`,
		},
		testCaseOpaValidator{
			name: "rule has an empty violations",
			rule: `
			violation[issue] {
				issue = ""
			}`,
			violationMsg: "",
			hasViolation: true,
		},
		testCaseOpaValidator{
			name: "rule has a bool violations",
			rule: `
			violation[issue] {
				issue = true
			}`,
			violationMsg: "true",
			hasViolation: true,
		},
	}

	for _, c := range cases {
		policy, _ := NewPolicy(c.rule)
		err := OpaValidator("{}", policy)
		if c.hasViolation {
			if err == nil {
				t.Errorf("[%s]: passed but should have been failed", c.name)
			}

			if err.Error() != c.violationMsg {
				t.Errorf("[%s]: expected error msg %s but got %s", c.name, c.violationMsg, err)
			}
		} else {
			if err != nil {
				t.Errorf("[%s]: %v", c.name, err)
			}
		}
	}
}
