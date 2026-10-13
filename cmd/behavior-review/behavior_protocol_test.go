// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package main

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestBehaviorCompletionDeclaration(t *testing.T) {
	for stageIndex, stage := range behaviorStages {
		for _, value := range []any{nil, "missing", "true", 1, false, true} {
			reports := behaviorFixtures()
			args := reportArgs(reports[stageIndex])
			submitted := args["report"].(map[string]any)
			delete(submitted, "stage")
			submitted["complete"] = value
			if value == "missing" {
				delete(submitted, "complete")
			}
			got, comments, err := parseBehaviorReport(args, stage, map[string]bool{"good.go": true}, reports[:stageIndex])
			complete, valid := value.(bool)
			if !valid {
				if err == nil || !strings.Contains(strings.ToLower(err.Error()), "complete") || len(comments) != 0 {
					t.Fatalf("%s complete=%v accepted or unnamed: %v, %+v", stage, value, err, comments)
				}
			} else if err != nil || got.Complete != complete || got.Stage != stage {
				t.Fatalf("%s complete=%v: report=%+v error=%v", stage, value, got, err)
			}
		}
	}
}

func TestBehaviorReferenceFeedback(t *testing.T) {
	reports := behaviorFixtures()
	second := reports[1].Checks[0]
	second.ID, second.ContractID = "t2", "renamed-two"
	reports[1].Checks[0].ContractID = "renamed-one"
	reports[1].Checks = append(reports[1].Checks, second)
	_, comments, err := parseBehaviorReport(reportArgs(reports[1]), "counterexamples", map[string]bool{"good.go": true}, reports[:1])
	if err == nil || len(comments) != 0 {
		t.Fatalf("invented references accepted: %v %+v", err, comments)
	}
	for _, detail := range []string{"t1", "renamed-one", "t2", "renamed-two", "valid contract_ids: [c1]"} {
		if !strings.Contains(err.Error(), detail) {
			t.Errorf("repair feedback omits %q: %v", detail, err)
		}
	}
	reports = behaviorFixtures()
	reports[2].Checks = nil
	_, _, err = parseBehaviorReport(reportArgs(reports[2]), "adjudication", map[string]bool{"good.go": true}, reports[:2])
	if err == nil || !strings.Contains(err.Error(), "contract_ids [c1] and prior check ids [t1]") {
		t.Fatalf("coverage repair needs the missing IDs: %v", err)
	}
}

func TestBehaviorHostRestoresDecisionLinks(t *testing.T) {
	reports := behaviorFixtures()
	contract := reports[0].Contracts[0]
	contract.ID = "c2"
	reports[0].Contracts = append(reports[0].Contracts, contract)
	check := reports[1].Checks[0]
	check.ID, check.ContractID, check.Scenario = "t2", "c2", "a different stored value"
	reports[1].Checks = append(reports[1].Checks, check)
	basis := reports[2].BasisReviews[0]
	basis.ContractID = "c2"
	reports[2].BasisReviews = append(reports[2].BasisReviews, basis)
	decision := reports[2].Checks[0]
	decision.ID = "t2"
	// Reorder the decisions: their positions cannot establish their identity.
	reports[2].Checks = append([]BehaviorCheck{decision}, reports[2].Checks...)
	args := reportArgs(reports[2])
	submitted := args["report"].(map[string]any)
	delete(submitted, "stage")
	for _, value := range submitted["checks"].([]any) {
		delete(value.(map[string]any), "contract_id")
	}
	before, _ := json.Marshal(reports[:2])
	got, _, err := parseBehaviorReport(args, "adjudication", map[string]bool{"good.go": true}, reports[:2])
	if err != nil || got.Checks[0].ContractID != "c2" || got.Checks[0].Scenario != check.Scenario || got.Checks[1].ContractID != "c1" {
		t.Fatalf("saved links were not restored by ID: report=%+v error=%v", got, err)
	}
	after, _ := json.Marshal(reports[:2])
	if string(before) != string(after) {
		t.Fatal("parsing changed the saved records")
	}
	// A supplied conflicting link must be rejected even when it names a real contract.
	submitted["checks"].([]any)[0].(map[string]any)["contract_id"] = "c1"
	if _, comments, err := parseBehaviorReport(args, "adjudication", map[string]bool{"good.go": true}, reports[:2]); err == nil || !strings.Contains(err.Error(), `belongs to contract_id "c2"`) || len(comments) != 0 {
		t.Fatalf("rebound decision accepted: %v %+v", err, comments)
	}
}

func TestBehaviorSchemaUsesSavedReferences(t *testing.T) {
	reports := behaviorFixtures()
	for i, stage := range behaviorStages {
		completion := behaviorCompletion(stage, reports[:i])
		reportSchema := completion.Parameters["properties"].(map[string]any)["report"].(map[string]any)
		if required := reportSchema["required"].([]string); required == nil || len(required) != 0 {
			t.Fatal("the report schema must allow entry-only or metadata-only patches")
		}
		props := reportSchema["properties"].(map[string]any)
		if _, present := props["pending"]; present {
			t.Fatal("the model cannot set host-owned pending work")
		}
		if _, present := props["stage"]; present {
			t.Fatal("the model is asked to repeat the host-owned stage")
		}
		if i == 0 {
			continue
		}
		checkSchema := props["checks"].(map[string]any)["items"].(map[string]any)
		check := checkSchema["properties"].(map[string]any)
		if stage == "counterexamples" {
			if !reflect.DeepEqual(check["contract_id"].(map[string]any)["enum"], []string{"c1"}) {
				t.Fatal("counterexample contract choices do not match saved records")
			}
		} else {
			if !slices.Contains(checkSchema["required"].([]string), "refutation") || !reflect.DeepEqual(check["refutation"].(map[string]any)["type"], []string{"object", "null"}) {
				t.Fatal("adjudication must request a refutation explicitly, with null for other verdicts")
			}
			if _, present := check["contract_id"]; present {
				t.Fatal("adjudication is asked to repeat the saved contract link")
			}
			if !reflect.DeepEqual(check["id"].(map[string]any)["enum"], []string{"t1"}) {
				t.Fatal("decision choices do not match saved checks")
			}
			basis := props["basis_reviews"].(map[string]any)["items"].(map[string]any)["properties"].(map[string]any)
			if !reflect.DeepEqual(basis["contract_id"].(map[string]any)["enum"], []string{"c1"}) {
				t.Fatal("basis choices do not match saved contracts")
			}
		}
	}
}
