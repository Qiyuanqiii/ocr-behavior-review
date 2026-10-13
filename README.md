# Behavior Review

[中文说明](README.zh-CN.md)

An opt-in review workflow for [OpenCodeReview](https://github.com/Qiyuanqiii/open-code-review).
The plugin traces sourced behavior contracts, searches for counterexamples, and
adjudicates findings against the evidence. Its executable owns the three stages,
report schemas, saved references, corrections and completion criteria.

## Requirements

Use the host integration on
[`codex/behavior-review-plugin`](https://github.com/Qiyuanqiii/open-code-review/tree/codex/behavior-review-plugin),
based on `codex/plugins-tui-latest`. The host must support review protocol
`stdio-v1`. This interface runs other plugins with their own stage plans and
validators. OCR supplies the configured model, tools, file/rule selection,
sessions, comment processing and filtering.

To build the compatible host with Go installed:

```sh
git clone --branch codex/behavior-review-plugin --single-branch https://github.com/Qiyuanqiii/open-code-review.git ocr-behavior-host
cd ocr-behavior-host
go build -o ocr ./cmd/opencodereview
```

On Windows, use `-o ocr.exe`. Run that binary directly or put its directory first
on `PATH`. The earlier v0.1.0 host with built-in `behavior-v1` requires an upgrade.

## Install

```sh
ocr plugin market add behavior https://raw.githubusercontent.com/Qiyuanqiii/ocr-behavior-review/main/marketplace.json
ocr plugin install behavior-review@behavior
```

For an existing marketplace installation, update the host first, then run:

```sh
ocr plugin market refresh behavior
ocr plugin update behavior-review
```

The release ZIP includes executables for Windows, Linux and macOS on amd64 and
arm64. Installation copies files without executing them or calling a model.
Explicitly starting a review runs the executable with your local permissions.

## Run once

```sh
ocr review --review-plugin behavior-review --preview
ocr review --review-plugin behavior-review
ocr review --from main --to HEAD --review-plugin behavior-review --format json --output behavior-review.json
```

In the TUI, choose the normal scope and background, then enter
`/preview behavior-review` or `/review behavior-review`. Preview makes no model
calls. The next plain `/review` uses the ordinary flow. Installing or globally
enabling the plugin does not select its workflow for future reviews.

Plugin reviews require a fresh invocation; `--resume` is currently unavailable.

## Review method

1. **Contracts:** identify consequential behavior, its sources and the premises
   connecting callers, data, state and execution order.
2. **Counterexamples:** derive attempts from the affected code and its
   relationships. Prioritize consequential premises with weak evidence and
   continue even when there are no initial candidates.
3. **Adjudication:** assess the original sources and preserved attempts. Confirm,
   refute, narrow or leave claims unresolved using applicable evidence.

The plugin retains valid partial entries, restores saved references and gives
specific correction feedback. Required unknowns keep the review incomplete.
OCR preserves task inputs during context compression, publishes accepted findings
through its normal filter, and records incomplete coverage with a nonzero exit.

The JSON `review_plugin` array contains stage status and the plugin's report under
`report`, including contracts, checks, rejected attempts, unknowns and pending work.
This replaces v0.1.0's `behavior_review` field. Completion describes assigned work
performed; it does not certify absence of defects.

`--effort` does not remove declared stages. Model reasoning remains an independent
setting. Existing tool-round, timeout, context and user-selected token limits
apply. There is no additional total token cap; `--max-tokens-budget 0` is unlimited.

## Development

- `plugin/behavior-review.md`: general judgment principles.
- `cmd/behavior-review/`: stage schemas, state, validation and stdio driver.
- `plugin/ocr-plugin.json`: stage plan and executable declaration.

The executable uses the Go standard library and has no dependency on OCR's
internal packages. Build and package with Go and Python 3:

```sh
make check test
python scripts/package.py
ocr plugin install ./dist/behavior-review-0.2.0.zip
```

Packaging builds six executables and a ZIP with a matching SHA-256 marketplace
catalog. Publish a new version for changed packages. The host's
[protocol documentation](https://github.com/Qiyuanqiii/open-code-review/blob/codex/behavior-review-plugin/docs/review-workflow-plugins.md)
describes the language-independent interface.

Keep prompt changes general. Historical misses belong in regression tests;
evaluate discovery gains on PRs not used to tune the policy, with corrected code
as false-positive controls. Integration tests measure execution compatibility;
discovery quality needs separate evaluation.

Apache-2.0. See [NOTICE](NOTICE) for the original implementation attribution.
