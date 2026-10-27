# Agent workflows

The active `security.yml` workflow is manually dispatched and has read-only repository permissions. It verifies module integrity, runs Go tests/race checks and vulnerability analysis, and cross-compiles Windows. It does not publish, install or deploy.

See [development and acceptance limits](../../docs/OPERATIONS.zh-CN.md) and [SECURITY.md](../../SECURITY.md). Historical upstream release/install automation is stored under `.github/upstream-workflows/` with disabled extensions. Do not treat those files as supported May release instructions or restore automatic upstream following without review.
