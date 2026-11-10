# Fixed SavePattern payload validation

This is a CI-only wrapper for the exact source of [Draft PR #18](https://github.com/b1ug/blink1-go/pull/18), commit `5dfb297f2958c8be0e4b0107aa47eaaca2b656f8`. The wrapper changes the workflow and adds this guide. It does not change SDK code, tests, README, dependencies, device reports, or the PR branch.

The workflow is prepared locally. No hosted execution is implied by this document. The wrapper must be reviewed before anyone pushes or dispatches it.

## Two source identities

The workflow comes from the reviewed wrapper commit on the dedicated branch `ci/savepattern-fixed-5dfb297-20261110`. Every job explicitly checks out the complete payload commit `5dfb297f2958c8be0e4b0107aa47eaaca2b656f8`, verifies `HEAD`, rejects a dirty worktree, and records both the workflow event commit and checked-out payload. There is no input for an arbitrary ref or test command.

The payload keeps its existing GID **v0.0.1** dependency. It is not the current integrated SDK b3/GID37 candidate, and its results must not fill the separate b3/GID37 matrix gap.

The fixed checkout means that the working tree inside a job contains the payload's old workflow file. GitHub has already read the wrapper workflow to construct that job; source verification must use the recorded wrapper event commit and payload commit separately.

## Matrix and execution scope

The matrix and Action revisions follow the successful [SDK #17 Build run](https://github.com/b1ug/blink1-go/actions/runs/42654372823):

| Jobs | Runner | Go | Checks |
| --- | --- | --- | --- |
| 2 | Ubuntu 24.04 amd64 | 1.26.x, 1.27.x | Native cgo build, race tests, vet |
| 2 | Ubuntu 24.04 arm64 | 1.26.x, 1.27.x | Native cgo build, race tests, vet |
| 2 | macOS 15 arm64 | 1.26.x, 1.27.x | Native cgo build, race tests, vet |
| 2 | macOS 15 Intel | 1.26.x, 1.27.x | Native cgo build, race tests, vet |
| 2 | Windows 2022 amd64 | 1.26.x, 1.27.x | Native cgo build, race tests, vet |
| 1 | Ubuntu 22.04 amd64 | 1.13.15 | Actual source-floor build and tests |

Windows runs the tests instead of skipping them. The legacy job does not claim race coverage. Each job records the actual Go/OS/architecture/cgo/compiler and module selection, and verifies downloaded modules. `GOWORK=off`, `GOTOOLCHAIN=local`, and `-mod=readonly` preserve the payload's selected modules. Optional device-test gates remain disabled. No USB device or camera is required or authorized by this workflow.

All jobs have a 15-minute bound and tests have a 2-minute bound. The token has only `contents: read`; checkout does not persist credentials. There are no release, deployment, artifact-upload, package-publishing, or repository-write steps.

The old PR workflow uses retired `ubuntu-20.04` and `macos-12` hosted images. The fixed wrapper uses the supported labels already exercised by SDK #17. This repairs the validation path, not the SavePattern implementation. Retired labels do not by themselves establish the sole cause of all old queued jobs. [Ubuntu retirement](https://github.blog/changelog/2025-01-15-github-actions-ubuntu-20-runner-image-brownout-dates-and-other-breaking-changes/), [macOS retirement](https://github.blog/changelog/2024-08-19-notice-of-upcoming-deprecations-and-breaking-changes-in-github-actions-runners/).

## Trigger only after review

The initial supported trigger is a deliberate push of the **reviewed wrapper commit** to the one allowlisted wrapper branch. That push starts this 11-job workflow. It does not update `fix/save-pattern-errors-5dfb297` or PR #18.

```sh
# Remote mutation and Actions execution: review before running.
git push origin <reviewed-wrapper-commit>:refs/heads/ci/savepattern-fixed-5dfb297-20261110
```

The placeholder must be replaced with the reviewed, frozen 40-character wrapper commit. Do not push a mutable local branch without checking its commit. Do not use force push. If the remote branch already exists, inspect its exact head and any live run first; do not assume a timeout means the push failed.

The YAML also has `workflow_dispatch` with no inputs, but a repository/branch condition prevents jobs outside the exact wrapper branch. GitHub requires the workflow to be available on the default branch for manual dispatch. This wrapper does not change the default branch, so it does **not** promise a new Run workflow button or successful `gh workflow run` registration. The reviewed branch-push path does not depend on that default-branch prerequisite. [Manual workflow requirements](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow).

Do not dispatch again after the push merely because jobs remain queued. Observe the same run ID and retain its source identity. Repeating jobs after a platform error requires a separate decision; this document does not authorize an automatic retry loop.

## Accepting hosted evidence

A completed qualification requires one successful workflow with all **11 expected jobs** present, each runner's actual toolchain and dependency selection recorded, the exact payload checkout verified, and the applicable build/test/race/vet steps successful. Inspect unexpected skips and logs instead of trusting only the aggregate green label.

The workflow `head_sha` describes the wrapper; it is not the payload SHA. The checkout/log proof must independently bind the payload. An archived local equivalence manifest should confirm every payload file except the workflow is byte-identical in the wrapper branch; this extra guide is the only added file.

Results from several failed or partial attempts may be listed separately, but must not be combined into a fictional single 11/11 pass. A success belongs to this SavePattern payload/GID v0.0.1 combination. It does not qualify a later SDK patch or the b3/GID37 integration.

These checks prove software behavior with fake handles. They do not prove real USB operation, Flash persistence, recovery after power loss, MK3 compatibility, an old operating-system runtime, or installed service behavior. A timeout after a save may mean it was already applied; the payload still forbids automatic save retries and does not turn nil into persistence proof.
