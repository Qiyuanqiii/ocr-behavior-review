// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 alibaba/open-code-review Contributors

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"unicode/utf8"
)

type Comment struct {
	Path           string `json:"path"`
	Content        string `json:"content"`
	ExistingCode   string `json:"existing_code,omitempty"`
	SuggestionCode string `json:"suggestion_code,omitempty"`
	StartLine      int    `json:"start_line"`
	EndLine        int    `json:"end_line"`
	Thinking       string `json:"thinking,omitempty"`
	Category       string `json:"category,omitempty"`
	Severity       string `json:"severity,omitempty"`
}

type Completion struct {
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type Request struct {
	Version    int            `json:"version"`
	Method     string         `json:"method"`
	Stage      string         `json:"stage"`
	Paths      []string       `json:"paths,omitempty"`
	Severities []string       `json:"severities,omitempty"`
	Categories []string       `json:"categories,omitempty"`
	Arguments  map[string]any `json:"arguments,omitempty"`
}

type Task struct {
	Instructions string     `json:"instructions"`
	Input        string     `json:"input"`
	Completion   Completion `json:"completion"`
}

type Response struct {
	Version  int             `json:"version"`
	Task     *Task           `json:"task,omitempty"`
	Accepted *bool           `json:"accepted,omitempty"`
	Complete *bool           `json:"complete,omitempty"`
	Feedback string          `json:"feedback,omitempty"`
	Report   *BehaviorReport `json:"report,omitempty"`
	Comments []Comment       `json:"comments,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// Each process handles one file group, sequentially. OCR supplies its current
// classification vocabulary before any schema is generated or report validated.
var severities = []string{"critical", "high", "medium", "low"}
var categories = []string{"bug", "security", "performance", "maintainability", "test", "style", "documentation", "other"}

func commentClassifications() ([]string, []string) { return severities, categories }

type driver struct {
	policy   string
	previous []BehaviorReport
	progress *behaviorProgress
	closed   bool
}

func (d *driver) handle(req Request) Response {
	res := Response{Version: 1}
	if req.Version != 1 {
		res.Error = "unsupported review protocol version"
		return res
	}
	switch req.Method {
	case "begin":
		next := len(d.previous)
		if next >= len(behaviorStages) || req.Stage != behaviorStages[next] || (d.progress != nil && !d.closed) {
			res.Error = "stage does not follow the completed behavior stage"
			return res
		}
		if len(req.Paths) == 0 || len(req.Severities) == 0 || len(req.Categories) == 0 {
			res.Error = "begin requires selected paths and comment classifications"
			return res
		}
		severities, categories = slices.Clone(req.Severities), slices.Clone(req.Categories)
		paths := map[string]bool{}
		for _, path := range req.Paths {
			paths[path] = true
		}
		d.progress = newBehaviorProgress(req.Stage, paths, d.previous)
		d.closed = false
		previous, _ := json.Marshal(d.previous)
		res.Task = &Task{Instructions: d.policy, Input: "Required stage: " + req.Stage + "\nPrevious stage records (claims to verify):\n" + string(previous), Completion: behaviorCompletion(req.Stage, d.previous)}
		report := d.progress.snapshot()
		res.Report = &report
	case "submit":
		if d.progress == nil || d.closed || req.Stage != d.progress.stage {
			res.Error = "submit requires the active stage"
			return res
		}
		report, comments, err := d.progress.submit(req.Arguments)
		res.Report = &report
		accepted := err == nil
		res.Accepted = &accepted
		if err != nil {
			res.Feedback = err.Error()
			return res
		}
		complete := report.Complete
		res.Comments = comments
		if !complete {
			res.Feedback = "required stage declared incomplete"
		}
		if req.Stage == "adjudication" {
			if reason := finalIncompleteReason(report); reason != "" {
				complete, res.Feedback = false, reason
			}
		}
		res.Complete = &complete
		d.closed = true
		if complete {
			d.previous = append(d.previous, report)
		}
	default:
		res.Error = "unsupported review method"
	}
	return res
}

func finalIncompleteReason(report BehaviorReport) string {
	for _, review := range report.BasisReviews {
		if review.Verdict == "unresolved" {
			return "behavior review has an unresolved contract basis"
		}
	}
	if len(report.Unknowns) > 0 {
		return "behavior review has unresolved required work"
	}
	for _, check := range report.Checks {
		if check.Verdict == "unresolved" {
			return "behavior review has unresolved counterexamples"
		}
	}
	return ""
}

func serve(in io.Reader, out io.Writer, policy string) error {
	d := &driver{policy: policy}
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 4<<20)
	encoder := json.NewEncoder(out)
	for scanner.Scan() {
		var req Request
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			return fmt.Errorf("decode request: %w", err)
		}
		if err := encoder.Encode(d.handle(req)); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func run() error {
	policyPath := flag.String("policy", "", "path to the packaged review policy")
	flag.Parse()
	file, err := os.Open(*policyPath)
	if err != nil {
		return err
	}
	defer file.Close()
	policy, err := io.ReadAll(io.LimitReader(file, 64*1024+1))
	if err != nil {
		return err
	}
	if len(policy) > 64*1024 || !utf8.Valid(policy) || strings.TrimSpace(string(policy)) == "" {
		return fmt.Errorf("policy must contain nonempty UTF-8 text up to 64 KiB")
	}
	return serve(os.Stdin, os.Stdout, string(policy))
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
