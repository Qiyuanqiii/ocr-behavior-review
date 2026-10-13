// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package main

import (
	"testing"
)

func classifiedBehaviorDecision(severity, category string) BehaviorReport {
	r := behaviorFixtures()[2]
	r.Checks[0].Verdict, r.Checks[0].Refutation = "confirmed", nil
	r.Checks[0].Finding = &Comment{Path: "good.go", Content: "The returned value differs from the stored value.", ExistingCode: "value", Severity: severity, Category: category}
	return r
}

func TestBehaviorRejectsLossyClassification(t *testing.T) {
	for _, pair := range [][2]string{{"moderate", "bug"}, {"medium", "correctness"}, {"", "bug"}, {"medium", ""}} {
		t.Run(pair[0]+"/"+pair[1], func(t *testing.T) {
			reports := behaviorFixtures()
			r := classifiedBehaviorDecision(pair[0], pair[1])
			if _, _, err := parseBehaviorReport(reportArgs(r), "adjudication", map[string]bool{"good.go": true}, reports[:2]); err == nil {
				t.Fatal("classification requiring a fallback was accepted")
			}
		})
	}
}
