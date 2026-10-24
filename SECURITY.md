# May Monitor agent security profile

Remote control and automatic updating default to disabled. The remote-control
switch also covers terminal sessions, files, startup configuration and version
switching. Operator opt-in expands the trust boundary and is outside approval
for the default monitoring profile.

Use HTTPS for non-loopback panel endpoints. Certificate verification is enabled
by default; the unsafe-certificate override is an explicit operator exception.
Update downloads always use certificate verification, independently of that
override, require a supported SHA-256 digest and reject unsafe redirects. The
default update repository is the owned May fork. A digest from the release API
is an integrity check, not an independent maintainer signature.

HTTP responses and decompressed WebSocket messages are bounded to 4 MiB.
Transport logs omit token-bearing request URLs, response bodies and command
text. Store configuration files containing tokens under private OS permissions.
Tokens identify a specific agent; do not share them between installations.

The old upstream installers are retained as `*.upstream-disabled` for review,
not as executable deployment entry points. Historical release workflows are
inactive under `.github/upstream-workflows/`. The manual security workflow
tests and compiles only; it does not install, publish or change a remote host.
Use Go 1.27.3 and the pinned module sums.

The workspace 2026-10-24 report contains baseline/fixed evidence. Linux tests
use temporary files and loopback listeners; raw ICMP requires an OS capability
and is skipped when unavailable. Windows cross-compilation is not a Windows
runtime, service, privilege or installation acceptance test.

## First May prerelease

`v0.1.0-rc.1` packages this reviewed monitoring profile. See the [shared security review](https://github.com/mxh110708/may-monitor-server/blob/v0.1.0-rc.1/docs/SECURITY-REVIEW-2026-10-24.zh-CN.md). Release manifests pin all three source commits. No automatic installation or production migration is authorized.
