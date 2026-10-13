// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func beginStage(d *driver, stage string) Response {
	return d.handle(Request{Version: 1, Method: "begin", Stage: stage, Paths: []string{"good.go"}, Severities: severities, Categories: categories})
}

func TestDriverStagePlanAndUnknowns(t *testing.T) {
	for _, mode := range []string{"complete", "unknown", "basis", "check", "incomplete", "confirmed-with-unknown"} {
		t.Run(mode, func(t *testing.T) {
			d := &driver{policy: "Review affected behavior."}
			fixtures := behaviorFixtures()
			if mode == "unknown" || mode == "confirmed-with-unknown" {
				fixtures[2].Unknowns = []string{"unverified caller"}
			}
			if mode == "basis" {
				fixtures[2].BasisReviews[0].Verdict = "unresolved"
				fixtures[2].Checks[0].Verdict = "unresolved"
				fixtures[2].Checks[0].Refutation = nil
			}
			if mode == "check" {
				fixtures[2].Checks[0].Verdict = "unresolved"
				fixtures[2].Checks[0].Refutation = nil
			}
			if mode == "incomplete" {
				fixtures[0].Complete = false
			}
			if mode == "confirmed-with-unknown" {
				fixtures[2].Checks[0].Verdict = "confirmed"
				fixtures[2].Checks[0].Refutation = nil
				fixtures[2].Checks[0].Finding = &Comment{Path: "good.go", Content: "Value lost.", ExistingCode: "value", Severity: "medium", Category: "bug"}
			}
			for i, stage := range behaviorStages {
				begin := beginStage(d, stage)
				if begin.Error != "" || begin.Task == nil || begin.Report == nil || !strings.Contains(begin.Task.Input, stage) {
					t.Fatalf("begin=%+v", begin)
				}
				res := d.handle(Request{Version: 1, Method: "submit", Stage: stage, Arguments: reportArgs(fixtures[i])})
				if res.Accepted == nil || !*res.Accepted {
					t.Fatalf("submission=%+v", res)
				}
				want := mode == "complete" || (i < 2 && mode != "incomplete")
				if *res.Complete != want {
					t.Fatalf("completion=%+v", res)
				}
				if mode == "confirmed-with-unknown" && i == 2 && len(res.Comments) != 1 {
					t.Fatal("confirmed result lost when other work is incomplete")
				}
				if !want {
					if res.Feedback == "" {
						t.Fatal("missing incomplete reason")
					}
					break
				}
			}
		})
	}
}

func TestDriverRetainsValidEntriesAndRepairsMetadata(t *testing.T) {
	d := &driver{policy: "Review behavior."}
	fixtures := behaviorFixtures()
	for i, stage := range behaviorStages {
		if beginStage(d, stage).Error != "" {
			t.Fatal("could not begin")
		}
		args := reportArgs(fixtures[i])
		delete(args["report"].(map[string]any), "complete")
		bad := d.handle(Request{Version: 1, Method: "submit", Stage: stage, Arguments: args})
		if bad.Accepted == nil || *bad.Accepted || !strings.Contains(bad.Feedback, "report.complete") || bad.Report == nil {
			t.Fatalf("bad=%+v", bad)
		}
		fixed := d.handle(Request{Version: 1, Method: "submit", Stage: stage, Arguments: behaviorPatch(map[string]any{"complete": true})})
		if fixed.Accepted == nil || !*fixed.Accepted || !*fixed.Complete {
			t.Fatalf("fixed=%+v", fixed)
		}
	}
	if !reflect.DeepEqual(d.previous[2].Checks[0].Trace, fixtures[1].Checks[0].Trace) {
		t.Fatal("saved attempt changed")
	}
}

func TestDriverRejectsOutOfOrderRequests(t *testing.T) {
	for _, req := range []Request{{Version: 2, Method: "begin"}, {Version: 1, Method: "submit", Stage: "contracts"}, {Version: 1, Method: "begin", Stage: "adjudication"}, {Version: 1, Method: "other"}} {
		if res := (&driver{}).handle(req); res.Error == "" {
			t.Fatalf("accepted invalid request: %+v", req)
		}
	}
}

func TestServeFraming(t *testing.T) {
	request, _ := json.Marshal(Request{Version: 1, Method: "begin", Stage: "contracts", Paths: []string{"good.go"}, Severities: severities, Categories: categories})
	var output bytes.Buffer
	if err := serve(bytes.NewReader(append(request, '\n')), &output, "policy"); err != nil {
		t.Fatal(err)
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil || response.Task == nil {
		t.Fatalf("response=%s error=%v", output.String(), err)
	}
	if err := serve(strings.NewReader("broken\n"), &output, "policy"); err == nil {
		t.Fatal("accepted malformed protocol")
	}
}

func TestDriverUsesHostClassifications(t *testing.T) {
	oldSeverities, oldCategories := severities, categories
	t.Cleanup(func() { severities, categories = oldSeverities, oldCategories })
	d := &driver{policy: "policy"}
	fixtures := behaviorFixtures()
	for i, stage := range behaviorStages {
		res := d.handle(Request{Version: 1, Method: "begin", Stage: stage, Paths: []string{"good.go"}, Severities: []string{"notice"}, Categories: []string{"correctness"}})
		if res.Error != "" {
			t.Fatal(res.Error)
		}
		if i == 2 {
			fixtures[i].Checks[0].Verdict, fixtures[i].Checks[0].Refutation = "confirmed", nil
			fixtures[i].Checks[0].Finding = &Comment{Path: "good.go", Content: "Value lost", ExistingCode: "value", Severity: "notice", Category: "correctness"}
		}
		submitted := d.handle(Request{Version: 1, Method: "submit", Stage: stage, Arguments: reportArgs(fixtures[i])})
		if submitted.Accepted == nil || !*submitted.Accepted {
			t.Fatalf("host classification rejected: %+v", submitted)
		}
		if i == 2 && (len(submitted.Comments) != 1 || submitted.Comments[0].Severity != "notice") {
			t.Fatal("classification lost")
		}
	}
}
