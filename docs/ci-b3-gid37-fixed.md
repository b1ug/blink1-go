# Fixed SDK b3 and GID37 validation

This CI-only wrapper validates the exact SDK payload `b3f2586cc73073a7a0432889ea14a18c13247f46` with its existing GID dependency `v0.0.3-0.20261109043911-37d7cfde72ed`. It changes only `.github/workflows/build.yml` and adds this guide. SDK source, tests, README, module files, reports and all other payload Git blobs and modes remain unchanged. A local freeze compares every such file; Windows filesystem permission equivalence is not implied by Git mode identity.

This is a locally prepared candidate. It has not been pushed, dispatched, retried, or run on hosted Actions. Root must independently review the frozen wrapper before a deliberate push. The checked-out source in every job is the literal full b3 SHA; there is no ref input, dynamic payload selection, or code change to make a matrix pass.

## Two identities, one dependency combination

The workflow comes from the reviewed wrapper commit on `ci/b3-gid37-fixed-b3f2586-20261110`. GitHub's workflow `head_sha` is that wrapper commit. Every job separately logs the workflow event/file commit and proves that its SDK checkout is exact b3 and clean. Do not call the wrapper commit the tested SDK source.

Every job records the selected module graph and confirms GID's exact version, no replacement anywhere, the original `go.mod`/`go.sum` Git object identities, the selected module zip and go.mod checksums, and `go mod verify`. Changes to module files or an unexpected working tree fail the job. GID's selected checksums are:

```text
h1:yxpXHZnzwCPfGfEfxYTeTJ6Pz72M1grVJ3EllxQEa38=
h1:b9pd+IB3F4qVXRRUOWgux7TWqxceIEGo/mhoBcOp9RM=
```

`go mod download -json` provides checksum fields on the actual Go 1.13 floor. The workflow does not assume that older `go list -m` exposes newer checksum fields. [Go 1.13 download implementation](https://github.com/golang/go/blob/go1.13.15/src/cmd/go/internal/modcmd/download.go), [module verification reference](https://go.dev/ref/mod#go-mod-verify).

## Eleven native jobs

| Hosted runner | Expected native target | Go line | Checks |
| --- | --- | --- | --- |
| `ubuntu-24.04` | linux/amd64, X64 runner | 1.26.x and 1.27.x | CGO1 build, race tests, vet |
| `ubuntu-24.04-arm` | linux/arm64, ARM64 runner | 1.26.x and 1.27.x | CGO1 build, race tests, vet |
| `macos-15` | darwin/arm64, ARM64 runner | 1.26.x and 1.27.x | CGO1 build, race tests, vet |
| `macos-15-intel` | darwin/amd64, X64 runner | 1.26.x and 1.27.x | CGO1 build, race tests, vet |
| `windows-2022` | windows/amd64, X64 runner | 1.26.x and 1.27.x | CGO1 build, race tests, vet |
| `ubuntu-22.04` | linux/amd64, X64 runner | actual 1.13.15 | CGO1 build and source-floor tests |

The five modern runner pairs make ten jobs; the Linux source-floor job is the eleventh. The floor job does not claim race/vet or an old operating-system runtime. Each job records actual Go version, GOOS/GOARCH, CGO, compiler and runner architecture, with explicit native-target checks. Supported standard labels and architecture are documented by [GitHub's hosted runner reference](https://docs.github.com/en/actions/reference/runners/github-hosted-runners) and [default variables reference](https://docs.github.com/en/actions/reference/workflows-and-actions/variables).

Tests run with `-count=1`, a two-minute timeout and JSON events. A small Go 1.13-compatible standard-library log summarizer is created under `RUNNER_TEMP`, outside the payload. It records passed nodes, top-level tests, failed/skipped nodes, passed/skipped packages, command exit and wrapper/payload/GID identities. Duplicate terminal nodes, empty success evidence, failures, skips, failed test exit or a failed log pipeline reject qualification. Compiler text outside JSON is counted and remains visible for review. These counts are evidence to inspect, not a substitute for the actual test exit or job conclusion.

No hardware discovery/path gate is enabled. There is no USB device, camera, physical persistence, service, package publishing, release or artifact-upload step. The token has only `contents: read`; checkout does not persist credentials and setup does not restore Go caches. Each job has a fifteen-minute bound. The Actions runner's normal setup and teardown are separate from SDK operation qualification.

## Existing runs remain separate

The old integrated run [42671837084](https://github.com/b1ug/blink1-go/actions/runs/42671837084) used wrapper d7 with b3/GID37 source identity. Its three attempts provide only seven different successful combinations out of eleven; they are not one successful matrix. Missing/platform-blocked or queued jobs are not automatically source-test failures, and the wrapper does not retry them.

The successful [8b wrapper run 42746189614](https://github.com/b1ug/blink1-go/actions/runs/42746189614) proves eleven jobs for `5dfb297f2958c8be0e4b0107aa47eaaca2b656f8` plus GID v0.0.1. That result does not cover b3/GID37. Draft [PR #18](https://github.com/b1ug/blink1-go/pull/18), its source and merge decision remain unchanged by this wrapper.

## Planned trigger after independent review

The initial supported trigger is one normal push of the exact, reviewed wrapper SHA to the single allowlisted branch. No force push or PR update is needed. Before pushing, inspect the exact remote branch head and existing live runs; a timeout does not prove that a push failed. This guide does not authorize a push, dispatch or retry.

`workflow_dispatch` has no inputs and the same repository/branch job guard. GitHub requires a workflow on the default branch for manual dispatch; this candidate does not change that branch and does not promise a Run workflow button. [Manual workflow requirements](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow).

Do not dispatch a second run because the push-created run is queued. Do not combine successful jobs from different attempts, payloads or dependency versions into a fictional pass. Qualification requires one exact-wrapper/exact-b3/GID37 run with all eleven expected jobs and complete primary logs. An organization self-hosted runner listing may be permission-restricted; that does not establish whether a standard hosted label can be scheduled.

Software CI does not qualify real USB, MK3 hardware, Flash/power-loss behavior, old OS support, optical timing or installed services. Current SDK/GID deployment and independent fixes remain separate.
