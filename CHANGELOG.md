# Changelog

## 0.1.0 — 2026-10-27

Promote the accepted RC4 monitoring profile without functional changes. Preserve disabled remote control and auto-update, bounded probes and explicit credentials. See RELEASE-0.1.0.md.

## 0.1.0-rc.1 — 2026-10-24

Default to monitoring with remote control and automatic updates disabled. Bound server responses, redact sensitive transport errors, enforce update digests and strict update TLS. Publish May-named update assets from the owned repository.

Initial prerelease based on Komari. Security defaults remain enabled. Existing Komari, Deploy and Route installations are not modified by publishing this release. See SECURITY.md and the shared review for test coverage and limits.
