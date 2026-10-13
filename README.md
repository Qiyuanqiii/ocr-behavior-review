# Behavior Review

[中文说明](README.zh-CN.md)

An opt-in review policy for [OpenCodeReview](https://github.com/Qiyuanqiii/open-code-review).
It traces sourced behavior contracts, searches for counterexamples in the affected
code, and checks the evidence before publishing findings.

## Requirements

Use the host integration on
[`codex/behavior-review-plugin`](https://github.com/Qiyuanqiii/open-code-review/tree/codex/behavior-review-plugin),
based on `codex/plugins-tui-latest`. Check that `ocr review --help` lists
`--review-plugin`. Earlier hosts only support rule packs and MCP tools and will
reject this package's `review` contribution.

To build the compatible host from source with Go installed:

```sh
git clone --branch codex/behavior-review-plugin --single-branch https://github.com/Qiyuanqiii/open-code-review.git ocr-behavior-host
cd ocr-behavior-host
go build -o ocr ./cmd/opencodereview
```

On Windows, use `-o ocr.exe`. Run this binary directly or put its directory first
on `PATH` for the commands below.

The plugin uses OCR's configured provider and model. Installation makes no model
requests. A review uses the normal OCR tools, file selection, project rules,
sessions, comment processing and filtering.

## Install

From a checkout of this repository:

```sh
git clone https://github.com/Qiyuanqiii/ocr-behavior-review.git
ocr plugin install ./ocr-behavior-review/plugin
```

Or install the release through its marketplace:

```sh
ocr plugin market add behavior https://raw.githubusercontent.com/Qiyuanqiii/ocr-behavior-review/main/marketplace.json
ocr plugin install behavior-review@behavior
```

Keep the package installed without `--enable`. Review policies require an explicit
selection on each invocation, including when the plugin's saved state is enabled.

## Run once

```sh
# Inspect selected files without calling the model
ocr review --review-plugin behavior-review --preview

# Review local changes
ocr review --review-plugin behavior-review

# Review a branch and save the full stage records
ocr review --from main --to HEAD --review-plugin behavior-review --format json --output behavior-review.json
```

In the TUI, choose the normal scope and background, then enter:

```text
/scope range main HEAD
/preview behavior-review
/review behavior-review
```

The next `/review` uses the ordinary review flow. Preview does not select the
policy for a later command. A new invocation is required to rerun this policy;
`--resume` is currently unavailable for plugin reviews.

## Review method

1. **Contracts:** identify consequential behavior in changed files, the sources
   that establish it, and its dependencies on callers, data, state and ordering.
2. **Counterexamples:** challenge the premises with cases derived from the current
   code and its relationships, prioritizing consequence and weak evidence. This
   stage runs even when there are no initial candidates.
3. **Adjudication:** check each original source and attempt. Confirm, refute,
   narrow or leave the claim unresolved using applicable evidence.

The host retains accepted entries, restores saved links, reports specific
corrections and preserves stage inputs during context compression. Confirmed
findings pass through OCR's normal publication and filtering. The JSON output's
`behavior_review` array also retains rejected attempts, unknowns, pending work and
completion status. Required unresolved work produces incomplete coverage and a
nonzero exit, while available findings remain in the output.

This policy always requests three stages. `--effort` does not remove stages;
model reasoning settings remain independent. OCR's existing tool-round, timeout,
context and user-selected token limits still apply. No extra token cap is imposed;
`--max-tokens-budget 0` means unlimited total tokens. Completion describes the work
performed, and does not certify that the reviewed code has no defects.

## Development

The package is `plugin/ocr-plugin.json` plus `plugin/behavior-review.md`. The host
owns the `behavior-v1` protocol, validation and execution. Change the policy here;
change protocol behavior and its regression tests in the host repository.

```sh
python scripts/package.py
python -m unittest discover -s tests
```

Packaging produces a ZIP with the manifest at its root and a marketplace catalog
with the archive's SHA-256. Copy the generated catalog to `marketplace.json` when
publishing the matching release asset. Use new version tags for changed packages.

Prompt changes should remain general principles. Historical misses belong in
regression cases; evaluate discovery improvements on PRs not used to tune the
policy, with corrected code as false-positive controls. Packaging tests establish
installation and execution compatibility. They do not measure discovery quality.

## License

Apache-2.0. The policy and host mechanism were developed from the OCR behavior
review prototype. See [NOTICE](NOTICE).
