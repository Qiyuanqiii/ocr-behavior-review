// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func behaviorFixtures() []BehaviorReport {
	contract := BehaviorContract{
		ID: "c1", Path: "good.go", Behavior: "return a value", Basis: "caller expects a value", BasisStatus: "explicit", Relation: "caller -> value",
		Sources:      []BehaviorContractSource{{ID: "src1", Kind: "interface", Citation: "good.go:1 Get", Statement: "Get returns the stored value."}},
		Observations: []string{"good.go:1 return value"},
	}
	basis := BehaviorBasisReview{
		ContractID: "c1", Verdict: "supported", Analysis: "The interface requires preserving the stored value, including zero.",
		SourceAssessments: []BehaviorSourceAssessment{{SourceID: "src1", Analysis: "Get's return contract applies to the selected function."}},
		Evidence:          []string{"good.go:1 Get returns the stored value"},
	}
	check := BehaviorCheck{ID: "t1", ContractID: "c1", Scenario: "value is zero", Expected: "return zero to caller", Outcome: "caller receives zero", Trace: []BehaviorTraceStep{{ID: "s1", Code: "good.go:1 return value", Value: "return 0"}}, Analysis: "caller accepts zero", Evidence: []string{"good.go:1 return value"}, Counterevidence: "zero is explicitly accepted", Verdict: "refuted"}
	decision := check
	decision.Scenario, decision.Expected, decision.Outcome, decision.Trace = "", "", "", nil
	decision.Refutation = &BehaviorRefutation{Target: "outcome", Reason: "returned zero satisfies the caller's expectation"}
	return []BehaviorReport{
		{Stage: "contracts", Complete: true, Contracts: []BehaviorContract{contract}, Unknowns: []string{}, Limitations: []string{}},
		{Stage: "counterexamples", Complete: true, Checks: []BehaviorCheck{check}, Unknowns: []string{}, Limitations: []string{}},
		{Stage: "adjudication", Complete: true, BasisReviews: []BehaviorBasisReview{basis}, Checks: []BehaviorCheck{decision}, Unknowns: []string{}, Limitations: []string{}},
	}
}

func reportArgs(report BehaviorReport) map[string]any {
	data, _ := json.Marshal(report)
	var value any
	_ = json.Unmarshal(data, &value)
	return map[string]any{"report": value}
}

func TestBehaviorReportCoverage(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stage int
		edit  func(*BehaviorReport)
	}{
		{"missing path", 0, func(r *BehaviorReport) { r.Contracts = nil }},
		{"invented path", 0, func(r *BehaviorReport) { r.Contracts[0].Path = "other.go" }},
		{"missing basis", 0, func(r *BehaviorReport) { r.Contracts[0].Basis = " " }},
		{"empty observation", 0, func(r *BehaviorReport) { r.Contracts[0].Observations = []string{" "} }},
		{"duplicate contract", 0, func(r *BehaviorReport) { r.Contracts = append(r.Contracts, r.Contracts[0]) }},
		{"missing unknowns", 0, func(r *BehaviorReport) { r.Unknowns = nil }},
		{"missing limitations", 0, func(r *BehaviorReport) { r.Limitations = nil }},
		{"wrong stage", 1, func(r *BehaviorReport) { r.Stage = "contracts" }},
		{"missing attempt", 1, func(r *BehaviorReport) { r.Checks = nil }},
		{"unknown contract", 1, func(r *BehaviorReport) { r.Checks[0].ContractID = "missing" }},
		{"no challenge", 1, func(r *BehaviorReport) { r.Checks[0].Counterevidence = "" }},
		{"no expectation", 1, func(r *BehaviorReport) { r.Checks[0].Expected = "" }},
		{"no outcome", 1, func(r *BehaviorReport) { r.Checks[0].Outcome = "" }},
		{"no trace", 1, func(r *BehaviorReport) { r.Checks[0].Trace = nil }},
		{"uncited trace", 1, func(r *BehaviorReport) { r.Checks[0].Trace[0].Code = "" }},
		{"duplicate step", 1, func(r *BehaviorReport) { r.Checks[0].Trace = append(r.Checks[0].Trace, r.Checks[0].Trace[0]) }},
		{"oversized trace", 1, func(r *BehaviorReport) { r.Checks[0].Trace = make([]BehaviorTraceStep, 9) }},
		{"duplicate check", 1, func(r *BehaviorReport) { r.Checks = append(r.Checks, r.Checks[0]) }},
		{"premature confirmation", 1, func(r *BehaviorReport) { r.Checks[0].Verdict = "confirmed" }},
		{"dropped attempt", 2, func(r *BehaviorReport) { r.Checks[0].ID = "replacement" }},
		{"rewritten scenario", 2, func(r *BehaviorReport) { r.Checks[0].Scenario = "value is one" }},
		{"rewritten expectation", 2, func(r *BehaviorReport) { r.Checks[0].Expected = "return one" }},
		{"rewritten outcome", 2, func(r *BehaviorReport) { r.Checks[0].Outcome = "return one" }},
		{"rewritten trace", 2, func(r *BehaviorReport) {
			r.Checks[0].Trace = []BehaviorTraceStep{{ID: "s1", Code: "other code", Value: "1"}}
		}},
		{"unbound refutation", 2, func(r *BehaviorReport) { r.Checks[0].Refutation = nil }},
		{"unsupported target", 2, func(r *BehaviorReport) { r.Checks[0].Refutation.Target = "severity" }},
		{"no rebuttal", 2, func(r *BehaviorReport) { r.Checks[0].Refutation.Reason = " " }},
		{"unknown step", 2, func(r *BehaviorReport) {
			r.Checks[0].Refutation = &BehaviorRefutation{Target: "step", StepID: "s2", Reason: "different computation"}
		}},
		{"step on other target", 2, func(r *BehaviorReport) { r.Checks[0].Refutation.StepID = "s1" }},
		{"unresolved with refutation", 2, func(r *BehaviorReport) { r.Checks[0].Verdict = "unresolved" }},
		{"undecided candidate", 2, func(r *BehaviorReport) { r.Checks[0].Verdict = "candidate" }},
		{"missing finding", 2, func(r *BehaviorReport) { r.Checks[0].Verdict = "confirmed" }},
		{"refuted finding", 2, func(r *BehaviorReport) { r.Checks[0].Finding = &Comment{} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reports := behaviorFixtures()
			report := reports[tc.stage]
			tc.edit(&report)
			if _, _, err := parseBehaviorReport(reportArgs(report), behaviorStages[tc.stage], map[string]bool{"good.go": true}, reports[:tc.stage]); err == nil {
				t.Fatal("invalid coverage accepted")
			}
		})
	}
}

func TestAdjudicationPreservesAttempt(t *testing.T) {
	for _, target := range []string{"basis", "input", "step", "outcome", "change"} {
		t.Run(target, func(t *testing.T) {
			reports := behaviorFixtures()
			reports[2].Checks[0].Refutation.Target = target
			if target == "basis" {
				reports[2].BasisReviews[0].Verdict = "unsupported"
			}
			if target == "step" {
				reports[2].Checks[0].Refutation.StepID = "s1"
			}
			got, _, err := parseBehaviorReport(reportArgs(reports[2]), "adjudication", map[string]bool{"good.go": true}, reports[:2])
			if err != nil {
				t.Fatal(err)
			}
			before, after := reports[1].Checks[0], got.Checks[0]
			if after.Scenario != before.Scenario || after.Expected != before.Expected || after.Outcome != before.Outcome || !reflect.DeepEqual(after.Trace, before.Trace) {
				t.Fatalf("attempt changed during adjudication: before=%+v after=%+v", before, after)
			}
			got.Checks[0].Trace[0].Value = "edited"
			if reports[1].Checks[0].Trace[0].Value != "return 0" {
				t.Fatal("adjudication aliases the preceding trace")
			}
		})
	}
}

func parseBehaviorReport(args map[string]any, stage string, paths map[string]bool, previous []BehaviorReport) (BehaviorReport, []Comment, error) {
	return newBehaviorProgress(stage, paths, previous).submit(args)
}

const behaviorPrompt = "Apply the selected behavior review policy to the assigned stage."
