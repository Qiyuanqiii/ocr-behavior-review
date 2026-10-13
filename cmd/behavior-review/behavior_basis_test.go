// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package main

import (
	"reflect"
	"testing"
)

func TestContractProvenanceValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*BehaviorContract)
	}{
		{"unknown status", func(c *BehaviorContract) { c.BasisStatus = "proven" }},
		{"missing sources", func(c *BehaviorContract) { c.Sources = nil }},
		{"ungrounded explicit basis", func(c *BehaviorContract) { c.Sources = []BehaviorContractSource{} }},
		{"duplicate source", func(c *BehaviorContract) { c.Sources = append(c.Sources, c.Sources[0]) }},
		{"too many sources", func(c *BehaviorContract) { c.Sources = make([]BehaviorContractSource, 9) }},
		{"missing citation", func(c *BehaviorContract) { c.Sources[0].Citation = " " }},
		{"missing statement", func(c *BehaviorContract) { c.Sources[0].Statement = "" }},
		{"implementation as authority", func(c *BehaviorContract) { c.Sources[0].Kind = "implementation" }},
		{"missing observations", func(c *BehaviorContract) { c.Observations = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report := behaviorFixtures()[0]
			tc.edit(&report.Contracts[0])
			if _, _, err := parseBehaviorReport(reportArgs(report), "contracts", map[string]bool{"good.go": true}, nil); err == nil {
				t.Fatal("invalid contract provenance accepted")
			}
		})
	}
	for _, status := range []string{"explicit", "inferred", "disputed", "unknown"} {
		report := behaviorFixtures()[0]
		report.Contracts[0].BasisStatus = status
		report.Contracts[0].Observations = []string{}
		if status == "unknown" {
			report.Contracts[0].Sources = []BehaviorContractSource{}
			report.Unknowns = []string{"The selected file's expected return value needs a caller or interface source."}
		}
		if _, _, err := parseBehaviorReport(reportArgs(report), "contracts", map[string]bool{"good.go": true}, nil); err != nil {
			t.Fatalf("%s basis: %v", status, err)
		}
	}
}

func TestBasisReviewCoverageAndVerdicts(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*BehaviorReport)
	}{
		{"missing reviews", func(r *BehaviorReport) { r.BasisReviews = nil }},
		{"too many reviews", func(r *BehaviorReport) { r.BasisReviews = make([]BehaviorBasisReview, 17) }},
		{"unknown contract", func(r *BehaviorReport) { r.BasisReviews[0].ContractID = "other" }},
		{"duplicate review", func(r *BehaviorReport) { r.BasisReviews = append(r.BasisReviews, r.BasisReviews[0]) }},
		{"missing analysis", func(r *BehaviorReport) { r.BasisReviews[0].Analysis = "" }},
		{"missing assessments", func(r *BehaviorReport) { r.BasisReviews[0].SourceAssessments = nil }},
		{"dropped source", func(r *BehaviorReport) { r.BasisReviews[0].SourceAssessments = []BehaviorSourceAssessment{} }},
		{"unknown source", func(r *BehaviorReport) { r.BasisReviews[0].SourceAssessments[0].SourceID = "other" }},
		{"empty response", func(r *BehaviorReport) { r.BasisReviews[0].SourceAssessments[0].Analysis = " " }},
		{"duplicate response", func(r *BehaviorReport) {
			r.BasisReviews[0].SourceAssessments = append(r.BasisReviews[0].SourceAssessments, r.BasisReviews[0].SourceAssessments[0])
		}},
		{"unknown verdict", func(r *BehaviorReport) { r.BasisReviews[0].Verdict = "true" }},
		{"missing evidence", func(r *BehaviorReport) { r.BasisReviews[0].Evidence = nil }},
		{"blank evidence", func(r *BehaviorReport) { r.BasisReviews[0].Evidence = []string{" "} }},
		{"unevidenced support", func(r *BehaviorReport) { r.BasisReviews[0].Evidence = []string{} }},
		{"unsupported outcome refutation", func(r *BehaviorReport) { r.BasisReviews[0].Verdict = "unsupported" }},
		{"unresolved outcome refutation", func(r *BehaviorReport) { r.BasisReviews[0].Verdict = "unresolved" }},
		{"contradictory basis refutation", func(r *BehaviorReport) { r.Checks[0].Refutation.Target = "basis" }},
		{"unresolved confirmation", func(r *BehaviorReport) {
			r.BasisReviews[0].Verdict = "unresolved"
			r.Checks[0].Verdict = "confirmed"
			r.Checks[0].Refutation = nil
			r.Checks[0].Finding = &Comment{Path: "good.go", Content: "value lost", ExistingCode: "value"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reports := behaviorFixtures()
			tc.edit(&reports[2])
			if _, _, err := parseBehaviorReport(reportArgs(reports[2]), "adjudication", map[string]bool{"good.go": true}, reports[:2]); err == nil {
				t.Fatal("invalid basis adjudication accepted")
			}
		})
	}
	for i := 0; i < 2; i++ {
		reports := behaviorFixtures()
		reports[i].BasisReviews = reports[2].BasisReviews
		if _, _, err := parseBehaviorReport(reportArgs(reports[i]), behaviorStages[i], map[string]bool{"good.go": true}, reports[:i]); err == nil {
			t.Fatal("premature basis review accepted")
		}
	}
}

// This fixture annotates f2 DEL-7 / CHK-07 from the 2026-09-22 full-flow pilot.
// Sources were checked at 2b5800e8e6204cf2cabef02639e4f6a0cf8e3890. The identity
// premise remains disputed; this scripted record cannot measure model quality.
func caseAliasBasisReports() []BehaviorReport {
	reports := behaviorFixtures()
	c := &reports[0].Contracts[0]
	c.Path = "internal/session/delete.go"
	c.Behavior = "A readable session for the same physical repository remains deletable through a case-variant path on a case-insensitive filesystem."
	c.BasisStatus = "disputed"
	c.Basis = "Physical identity is an inference from the guard's purpose and related path tests; the helper documentation's examples cover only separators and uncleaned paths."
	c.Relation = "session rm -> DeleteSession -> sameRepoPath -> refusal or removal"
	c.Sources = []BehaviorContractSource{
		{ID: "purpose", Kind: "interface", Citation: "internal/session/delete.go:19-27 ErrSessionOtherRepo", Statement: "ErrSessionOtherRepo reports that the session file found for an id records a different repository than the one being operated on."},
		{ID: "examples", Kind: "interface", Citation: "internal/session/delete.go:120-122 sameRepoPath", Statement: "a trailing separator or an uncleaned path does not read as a different repository"},
		{ID: "related-test", Kind: "test", Citation: "internal/pathutil/path_test.go:144-166 TestWithinBase_CaseInsensitiveVariant", Statement: "WithinBase(repo, variant) must hold after os.SameFile confirms the case-variant directory."},
	}
	c.Observations = []string{
		"internal/session/delete.go:123-125 compares filepath.Clean(a) == filepath.Clean(b) with case preserved.",
		"internal/viewer/handler.go:195 also compares CWD strings; this demonstrates an existing implementation convention.",
	}
	check := &reports[1].Checks[0]
	check.Scenario = "A readable record names C:/Repo; --repo C:/repo reaches the same repository and session store on a case-insensitive filesystem."
	check.Expected = "Under the inferred physical-identity obligation, remove the owned session."
	check.Outcome = "The record is found but the case-sensitive guard returns ErrSessionOtherRepo; the file remains."
	check.Trace = []BehaviorTraceStep{{ID: "guard", Code: "internal/session/delete.go:91-95,123-125 sameRepoPath", Value: "Clean preserves case; comparison is false; removal is skipped"}}
	check.Analysis = "The observable refusal is supported; its contract violation depends on the physical-identity premise."
	check.Evidence = []string{"internal/session/delete.go:91-95,123-125"}
	check.Counterevidence = "The documented examples are narrower, and selecting the original spelling permits deletion."
	check.Verdict = "candidate"
	review := &reports[2].BasisReviews[0]
	review.SourceAssessments = []BehaviorSourceAssessment{
		{SourceID: "purpose", Analysis: "The guard concerns another repository, but does not explicitly define physical identity."},
		{SourceID: "examples", Analysis: "The documented examples do not settle case aliases."},
		{SourceID: "related-test", Analysis: "This test requires physical identity in path containment; applicability to session ownership remains an inference."},
	}
	review.Analysis = "The original textual-comparison contract mixed mechanism with obligation; the cited sources leave this policy disputed."
	review.Evidence = []string{"internal/session/delete.go:19-27,120-125", "internal/pathutil/path_test.go:144-166"}
	review.Verdict = "unresolved"
	reports[2].Checks[0] = *check
	reports[2].Checks[0].Scenario, reports[2].Checks[0].Expected, reports[2].Checks[0].Outcome, reports[2].Checks[0].Trace = "", "", "", nil
	reports[2].Checks[0].Verdict = "unresolved"
	return reports
}

func TestCaseAliasBasisReplay(t *testing.T) {
	reports := caseAliasBasisReports()
	paths := map[string]bool{"internal/session/delete.go": true}
	var previous []BehaviorReport
	for _, input := range reports {
		report, comments, err := parseBehaviorReport(reportArgs(input), input.Stage, paths, previous)
		if err != nil || len(comments) != 0 {
			t.Fatalf("%s: comments=%v err=%v", input.Stage, comments, err)
		}
		previous = append(previous, report)
	}
	if !reflect.DeepEqual(previous[0].Contracts, reports[0].Contracts) || !reflect.DeepEqual(previous[2].Checks[0].Trace, reports[1].Checks[0].Trace) {
		t.Fatal("source or trace record changed during replay")
	}
	// The old basis-refutation shape cannot settle a recorded source conflict.
	reports[2].Checks[0].Verdict = "refuted"
	reports[2].Checks[0].Refutation = &BehaviorRefutation{Target: "basis", Reason: "The implementation uses a textual comparison."}
	if _, _, err := parseBehaviorReport(reportArgs(reports[2]), "adjudication", paths, previous[:2]); err == nil {
		t.Fatal("unresolved sources became a definitive basis refutation")
	}
	// Even a declared conclusion cannot silently drop the competing test source.
	reports[2].BasisReviews[0].Verdict = "unsupported"
	reports[2].BasisReviews[0].SourceAssessments = reports[2].BasisReviews[0].SourceAssessments[:2]
	if _, _, err := parseBehaviorReport(reportArgs(reports[2]), "adjudication", paths, previous[:2]); err == nil {
		t.Fatal("competing source disappeared from adjudication")
	}
}

func TestBasisCanBeResolvedWithEvidence(t *testing.T) {
	for _, verdict := range []string{"supported", "unsupported"} {
		t.Run(verdict, func(t *testing.T) {
			reports := behaviorFixtures()
			reports[0].Contracts[0].BasisStatus = "unknown"
			reports[0].Contracts[0].Sources = []BehaviorContractSource{}
			review := &reports[2].BasisReviews[0]
			review.SourceAssessments = []BehaviorSourceAssessment{}
			review.Verdict = verdict
			if verdict == "unsupported" {
				review.Analysis = "The previously unavailable interface defines this output as an identifier, defeating the hypothesized stored-value obligation."
				review.Evidence = []string{"api.go:5 Get returns an identifier, not the stored value."}
				reports[2].Checks[0].Refutation = &BehaviorRefutation{Target: "basis", Reason: "The interface specifies an identifier."}
			}
			got, comments, err := parseBehaviorReport(reportArgs(reports[2]), "adjudication", map[string]bool{"good.go": true}, reports[:2])
			if err != nil || len(comments) != 0 || got.BasisReviews[0].Verdict != verdict {
				t.Fatalf("evidence-backed resolution rejected: %v", err)
			}
			if reports[0].Contracts[0].BasisStatus != "unknown" || len(reports[0].Contracts[0].Sources) != 0 {
				t.Fatal("resolution rewrote original provenance")
			}
		})
	}
}
