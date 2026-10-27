# May Monitor Agent 0.1.0-rc.4

Reviewed monitoring-only prerelease containing the M4–M9 connection, telemetry and maintenance work. Keep explicit credentials, installation identity confirmation, bounded reconnects, sensitive-error redaction and default-off remote control and automatic updates.

Retire the startup bridge that scans and removes the old Komari update-motd hook. May no longer cleans files owned by the retired product. Current warning ownership checks and monitoring behavior remain intact.

Linux amd64 receives isolated runtime verification; arm64 and Windows amd64 are cross-compiled. Publishing does not install, replace or restart any production Agent.
