// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package main

import (
	"strings"
	"testing"
)

func TestBasisFeedbackListsOriginalSourceChoices(t *testing.T) {
	for _, assessments := range [][]BehaviorSourceAssessment{
		{{SourceID: "good.go", Analysis: "A citation is not the original ID."}},
		{{SourceID: "src1", Analysis: "checked"}, {SourceID: "src1", Analysis: "repeated"}},
		{{SourceID: "src1", Analysis: ""}},
	} {
		reports := behaviorFixtures()
		reports[0].Contracts[0].Sources = append(reports[0].Contracts[0].Sources,
			BehaviorContractSource{ID: "test-1", Kind: "test", Citation: "good_test.go:1", Statement: "Get preserves its stored value."})
		reports[2].BasisReviews[0].SourceAssessments = assessments
		_, _, err := parseBehaviorReport(reportArgs(reports[2]), "adjudication", map[string]bool{"good.go": true}, reports[:2])
		if err == nil || !strings.Contains(err.Error(), "original source_ids: [src1 test-1]") || !strings.Contains(err.Error(), "source_id \""+assessments[0].SourceID+"\"") {
			t.Fatalf("source correction lacks the bad ID or original choices: %v", err)
		}
	}
}

func TestBasisFeedbackCollectsAllOmissions(t *testing.T) {
	reports := behaviorFixtures()
	contracts := map[string]BehaviorContract{}
	var reviews []BehaviorBasisReview
	for _, id := range []string{"c2", "c1", "c4", "c3"} {
		contract := reports[0].Contracts[0]
		contract.ID = id
		contract.Sources = append([]BehaviorContractSource(nil), contract.Sources...)
		contract.Sources = append(contract.Sources,
			BehaviorContractSource{ID: id + "-test", Kind: "test", Citation: "good_test.go:1", Statement: "Get preserves its stored value."})
		contracts[id] = contract
		if id == "c1" || id == "c2" {
			review := reports[2].BasisReviews[0]
			review.ContractID = id
			reviews = append(reviews, review)
		}
	}
	verdicts, err := validateBehaviorBasisReviews(reviews, contracts)
	if err == nil || verdicts != nil {
		t.Fatalf("incomplete basis accepted: verdicts=%v error=%v", verdicts, err)
	}
	for _, missing := range []string{"c1-test", "c2-test", "c3", "c4"} {
		if !strings.Contains(err.Error(), missing) {
			t.Errorf("correction feedback omits %s: %v", missing, err)
		}
	}
	if !strings.Contains(err.Error(), "contract_ids [c3 c4]") {
		t.Errorf("missing contracts need stable IDs in feedback: %v", err)
	}
}
