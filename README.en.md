# May Monitor Agent

The host collector for May Monitor: CPU, memory, disk and network telemetry, report transport and configured probes. It is not the administration UI, a VPS deployment tool or a proxy client.

[简体中文](readme.md) · [Releases](https://github.com/mxh110708/may-monitor-agent/releases/latest) · [Operations](docs/OPERATIONS.zh-CN.md) · [Configuration](docs/CONFIGURATION.zh-CN.md)

The documentation baseline is **0.1.0**, paired with [May Monitor Server 0.1.0](https://github.com/mxh110708/may-monitor-server). Linux amd64 is the runtime acceptance target. Linux arm64 and Windows amd64 builds do not establish native service/permission acceptance.

Remote control and auto-update are disabled by default. The remote-control switch covers terminals, commands, files, startup configuration and version switching. Neither capability is needed for ordinary monitoring.

## Setup

Prefer [Deploy](https://github.com/mxh110708/may-vps-deploy) for managed installation. Verify the SSH host, HTTPS origin, Server UUID and this node's token before writing a protected configuration and starting a dedicated service.

For manual operation, verify the exact release artifact and checksum, create a dedicated account and a private config file:

```json
{
  "endpoint": "https://monitor.example.com",
  "token": "<THIS_NODE_AGENT_TOKEN>",
  "interval": 3,
  "disable_auto_update": true,
  "disable_web_ssh": true,
  "ignore_unsafe_cert": false
}
```

```bash
./may-monitor-agent --config '/etc/may-monitor-agent/config.json'
```

Non-loopback endpoints require HTTPS. Keep normal certificate verification. Tokens belong to one node and are not interchangeable with administrator passwords, root passwords or scoped read credentials. Do not put them in URLs, shared command history or screenshots.

## Precedence and validation

Configuration precedence, lowest to highest: **defaults → CLI → environment → JSON file**. A JSON value overrides a conflicting CLI flag. Inspect all sources when diagnosing an unexpected setting.

Default reporting is every three seconds, reconnection every five seconds and inventory every five minutes. The [configuration reference](docs/CONFIGURATION.zh-CN.md) lists exact field names and units. Verify the actual version, node/Server identity, advancing report time and stored history after startup; a running process is insufficient.

Change endpoints/tokens through a scoped maintenance operation with a snapshot and new credential verification. A revoked old token cannot be restored by copying its old config. Agent maintenance does not authorize changes to Server, tunnels, proxies or SSH.

Use Go 1.27.3 and locked module sums for development. Tests, simulated agents and cross-compilation have distinct evidence limits. See [SECURITY.md](SECURITY.md), [CHANGELOG](CHANGELOG.md) and [release notes](RELEASE-0.1.0.md).

Derived from [Komari Agent](https://github.com/komari-monitor/komari-agent), with retained module/protocol identifiers and [license](LICENSE). Historical upstream installers and publishing workflows remain inactive references.
