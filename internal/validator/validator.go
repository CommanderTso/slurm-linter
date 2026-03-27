package validator

import (
	"github.com/CommanderTso/slurm-linter/internal/diagnostic"
	"github.com/CommanderTso/slurm-linter/internal/rules"
)

// Validator runs a set of rules against a LintInput and collects all diagnostics.
type Validator struct {
	rules []rules.Rule
}

// New creates a Validator with the provided rules.
func New(r ...rules.Rule) *Validator {
	return &Validator{rules: r}
}

// Run executes all rules and returns the combined diagnostics.
func (v *Validator) Run(input *rules.LintInput) []diagnostic.Diagnostic {
	var all []diagnostic.Diagnostic
	for _, rule := range v.rules {
		all = append(all, rule.Check(input)...)
	}
	return all
}

// DefaultRules returns the standard set of rules for production use.
func DefaultRules() []rules.Rule {
	return []rules.Rule{
		rules.RequiredParams{},
		rules.TypeValidation{},
		rules.TopologyRule{},
		rules.CrossRefRule{},
	}
}
