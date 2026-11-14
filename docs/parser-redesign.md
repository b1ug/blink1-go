# Deterministic state parsing and shared text semantics

**Status: proposal for review; not implemented.** This document describes future
semantic changes. It does not change parser behavior, dependencies, device access,
or the existing JSON API. The grammar and breaking changes require review before
implementation.

## Problem and existing work

`ParseStateQuery` currently extracts a color, fade duration, and LED selector from
text and returns one `LightState`. It does not interpret action verbs, negation,
holding time, or the relationship between multiple instructions. Recognizing three
fields can therefore look like understanding a command when meaningful words were
ignored.

[PR #19](https://github.com/b1ug/blink1-go/pull/19) addresses a separate reliability
problem: deterministic duration selection, conflicting times, negative values and
overflow, zero values inside other fields, RGB/HSB whitespace, short titles, and
compact-state validation. Those fixes are existing work, not missing features of
this proposal. Their parser behavior is the baseline for the comparisons below.

The remaining issue is semantic accuracy. For example, `blink led 1 red 1s` can
produce a single red fade, and `do not turn led 1 red now` can still produce an
immediate red state. Adding more independent regular expressions does not establish
what the whole sentence means.

The recommendation is a small tokenizer followed by action-specific parsing,
complete input consumption, and explicit validation. Correct semantics take
priority over preserving accidental acceptance. Breaking changes should be
documented rather than hidden behind legacy modes or automatic parser fallback.

## Responsibilities and proposed inputs

The SDK should continue to describe one state through `LightState`. Its text entry
point can understand Set, Off, and Fade because these intents fit that type. Keep
existing function names and return signatures where practical; do not create
parallel Strict and Legacy APIs.

Higher-level effects belong to nb1. Its nine Action names remain `set`, `off`,
`fade`, `blink`, `breathe`, `pulse`, `alternate`, `orbit`, and `rainbow`. Text should
produce existing Action or Sequence values, then use the current compiler and
execution core. It must not implement another animation engine.

The following examples are proposed grammar, not currently available commands:

| Input | SDK state parser | nb1 text adapter |
| --- | --- | --- |
| `set red` | Immediate red, LED 0. | Set Action with existing defaults. |
| `off` | Immediate black, LED 0. | Off Action; this does not imply Job cancellation. |
| `fade to blue over 2s on led 1` | Blue, LED 1, two-second fade. | Fade Action with the same meaning. |
| `blink red 3 times` | Unsupported higher-level intent. | Three Blink cycles using documented timing defaults. |
| `set red hold 3s` | Unsupported: LightState has no hold field. | Serial Sequence with a three-second hold. |
| `set red hold 1s then fade to blue over 2s` | Unsupported multiple steps. | Two serial Steps. |

Set and Off may default to LED 0 and zero fade time. Fade requires an explicit
duration; colors are never invented. An explicit state form such as
`led=1 color=red fade=1s` may remain supported with exactly those three required
keys. Loose fragments such as `led 1 red 1s` would require rewriting into a command
or the explicit field form.

`over` describes a transition, `hold` describes dwell time, and `period` describes
one complete cycle where the Action accepts it. `fade ... in 2s` can mean a
transition when the Fade verb makes the role explicit. `in 2s set red` means a
delay and is outside the initial scope. An ambiguous request such as
`blink red for 3s` must not silently select a cycle count or a partial cycle.

Negation, unknown trailing content, repeated fields, and conflicting values produce
diagnostics instead of a partial command. The parser must not remove a negative
phrase and execute its opposite.

## Minimal implementation structure

Use tokens carrying their kind, original text, and byte range. Action-specific
functions read fields into small temporary values and produce the existing target
representation. Every meaningful token must be consumed; accepted connectors,
punctuation, and field names need an explicit grammar.

An independent AST is unnecessary for the initial scope. An Action is already a
field structure, and a simple `then` chain is a list of Steps. There is no immediate
need for generic nodes, visitors, tree rewriting, or an AST serialization contract.
A small composition tree can be reconsidered if nested repetition, conditions, or
parallel composition becomes a real requirement.

A small pure-Go SDK subpackage, provisionally `query`, could contain shared
tokenization, color literals, state parsing, and diagnostics. It must not import
the native USB backend or nb1. Existing SDK entry points would adapt its result
to `LightState`; nb1 would reuse the foundations and add Action/Sequence rules.
The package name and exported surface remain review decisions. Avoid building a
general language framework just to share a few parsing functions.

Preserve the SDK's Go 1.13 source floor. Native and WebAssembly reuse should happen
through pure business logic, without importing native HID code into the browser.
Parsing must not discover devices, start a daemon, acquire ownership, read local
configuration implicitly, or write patterns.

## Breaking behavior under consideration

These are proposed behavior changes, not claims that every previous choice was a
bug. Some fix misleading interpretation; others deliberately select a clearer
contract.

| Current behavior | Proposed change and caller impact |
| --- | --- |
| Set/Off require enough text to extract all three state fields. | `set red` and `off` gain explicit defaults and become valid. Callers cannot rely on those omissions always failing. |
| `ParseColor("please red")` extracts a matching color substring. | Color parsing consumes one complete color expression. Whole instructions belong to a command entry point. |
| Unknown words, negation, or an unsupported verb may be ignored. | Reject the unsupported intent or unexpected content, without returning an executable result. |
| Multiple colors or LEDs select a match by type priority or first occurrence. | Reject duplicate/conflicting state fields. The two colors of Alternate/Orbit remain legal because their grammar assigns distinct roles. |
| Numeric time overrides `now`; equivalent repeated times are accepted. | Reject repeated time fields and conflicts between immediate and nonzero transition timing. |
| `set ... for 3s` becomes a three-second fade. | Reject holding intent in the SDK state parser; nb1 can produce a host Sequence. Use `fade ... over 3s` for a transition. |
| Fractional milliseconds are truncated; writes can quantize or clamp. | Reject unrepresentable values before writing. Parser and relevant single-state write entry points need shared validation. |
| Out-of-range HSB channels are clipped, and some hues wrap. | Validate the chosen ranges. HSV spelling and valid decimal channels are extensions, not reasons to retain silent clipping. |
| `repeat 0` means unlimited; repeated declarations select the first. | Reject textual zero counts and duplicate declarations. Explicit `forever` can represent genuine SDK looping where supported; finite nb1 Jobs reject it. |
| Position aliases such as top/bottom imply fixed channel numbers. | Prefer numeric or first/second selectors; do not infer orientation. Single-target all/both maps to LED 0, while ordered dual-LED effects retain `[1,2]` or `[2,1]`. |
| `on` and `off` are also color aliases. | Prefer `white`/`black` as color values and `off`/`set white` as commands; caller wrappers must follow the same decision. |
| Errors are mainly text and may accompany a partially populated state. | Add typed diagnostics usable with `errors.As`. An error must not carry an executable partial result; callers should not match full message strings. |

Diagnostics should include a stable code, field, source range, and actionable hint.
Keep the original input for diagnostics and metadata; normalize supported ASCII
keywords locally instead of lowercasing the entire command. UTF-8 byte offsets
and display line/column information must agree for Unicode input.

### Color names and palette effects

The proposal selects nb1's existing frozen table of 148 opaque CSS names. A static
comparison with the SDK's 33 names finds 24 unchanged common names, four common
names with different RGB values, five SDK-only names, and 120 additional names.

| Name | Existing SDK RGB | Proposed RGB |
| --- | --- | --- |
| `green` | `#00ff00` | `#008000` |
| `lime` | `#008000` | `#00ff00` |
| `plum` | `#8e4585` | `#dda0dd` |
| `violet` | `#8000ff` | `#ee82ee` |

The SDK-only names are `apricot` (`#fbceb1`), `bronze` (`#cd7f32`), `mint`
(`#16982b`), `peach` (`#ffe5b4`), and `scarlet` (`#ff2400`). Use explicit Hex for
portable values; nb1 can save application-owned aliases when a color library is
available. The SDK must not implicitly read that library.

Approval would require checking named constants, forward and reverse lookup,
`GetColorNames`, and exported text together. `RainbowColors` references named
constants, so changing `ColorGreen` alone would also change its brightness.
Artistic palettes should have deliberate RGB anchors rather than accidental
dependence on a renamed standard color. Values already frozen as Hex do not need
rewriting solely because a name changes.

### Time representation and codecs

The existing SDK fade encoding uses 10ms units and a maximum of 655350ms. A 15ms
request must not quietly become 10ms. Set/Off may use zero; Fade must be positive
and exactly representable. nb1 keeps its narrower Action limits and its separate
Sequence budget; a single SDK state should not inherit a whole host Job's
24-hour budget.

`LightState.UnmarshalText` remains data decoding, not proof that a value can be
played. Explicit compact states and semicolon-separated SDK sequences can remain
supported. `ParseStateQuery` must validate a decoded compact state before treating
it as a playable request, and the final write boundary must validate again.
Keeping an explicit data format does not require fallback to loose text parsing.

## Cross-repository integration

SDK work alone cannot provide the complete nb1 experience. Follow-up nb1 changes
must account for these existing consumers:

- Color wrappers used by basic fade/blink commands and configuration validation
  must adopt complete color expressions and the agreed palette.
- The `act` command currently opens the device before parsing. Validation must
  precede USB access if invalid input is promised to have no device effects.
- `NormalizeQuery` currently lowercases the whole command. Preserve the original
  text and move keyword normalization into parsing.
- The text importer tries Title, State, then Repeat, ignoring unknown lines and
  accepting the first duplicate metadata value. Classify lines explicitly and
  report duplicates or unknown lines before submitting the file.
- The exporter emits numbered Shift/Maintain/Transition sentences. If the new
  grammar accepts only canonical state commands, the producer must change too.

For data valid and executable under the new contract, export followed by import
must preserve RGB, LED, duration, repetition, and title. Reject unplayable states
before exporting an executable file. A host hold cannot be disguised as a
same-color fade merely because the SDK file type lacks a hold field.

Blank lines and explicit single-line `//` comments can remain valid; `#abc` is a
color. Titles should retain Unicode and case. File parsing must not combine fields
from unrelated lines. Initial nb1 composition should cover only Set/Off/Fade with
hold and serial `then`; do not flatten an interval Orbit into an incompatible
serial plan.

## Acceptance and review decisions

Before implementation is considered complete:

- Define accepted examples and rejected counterexamples for every changed behavior,
  including negation, duplicates, unknown suffixes, missing fields, and defaults.
- Check the complete palette delta, numeric boundaries, Unicode locations,
  deterministic results, and bounded parsing without panics.
- Verify the Go 1.13 source floor and the current toolchain. Update obsolete
  expectations with explicit reasons instead of adding compatibility branches.
- Exercise export/import round trips and related configuration/command consumers.
  Verify invalid commands are rejected before device access.
- Compare text with equivalent JSON, and native with actual WASM, for the nb1
  integration. Start with software checks, then a small bounded hardware sample.

Review should decide the initial grammar, whether to adopt the complete CSS table
and its constant changes together, and the SDK/write-boundary scope. Implement SDK
semantics and necessary callers first, then nb1 Actions, followed by limited
Sequence authoring. No parser implementation, new release, or device qualification
is implied by this document-only draft.

## References

- [Reliability parser baseline](https://github.com/b1ug/blink1-go/blob/2e80d5a22c00d500c85b71222303c8f7dfd1fe7d/parser.go)
  and [regression tests](https://github.com/b1ug/blink1-go/blob/2e80d5a22c00d500c85b71222303c8f7dfd1fe7d/parser_reliability_test.go).
- [Existing palette](https://github.com/b1ug/blink1-go/blob/2e80d5a22c00d500c85b71222303c8f7dfd1fe7d/preset.go)
  and [fade encoding](https://github.com/b1ug/blink1-go/blob/2e80d5a22c00d500c85b71222303c8f7dfd1fe7d/util.go).
- [nb1 text import/export](https://github.com/b1ug/nb1/blob/112e224db3bcb7af823e4f403eec01f59a19d418/exchange/play.go),
  [frozen color table](https://github.com/b1ug/nb1/blob/112e224db3bcb7af823e4f403eec01f59a19d418/internal/colorinput/palette.go),
  and [Action/Sequence contract](https://github.com/b1ug/nb1/blob/112e224db3bcb7af823e4f403eec01f59a19d418/docs/api-v1.md).
