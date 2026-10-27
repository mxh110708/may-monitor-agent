# Agent 配置参考

基线：0.1.0。以 `cmd/root.go` 和 `cmd/flags/flag.go` 为合同，不沿用上游首次引入版本作为 May 版本号。所有秘密字段均需私密保存。

## 优先级与常用字段

默认值 → 命令行参数 → 环境变量 → JSON 配置文件，后者覆盖前者。`AGENT_CONFIG_FILE` 也能指定文件路径；最终文件内容不是给命令行补默认值，而是最高优先级的覆盖。

| JSON 字段 | CLI | 环境变量 | 默认／单位 |
| --- | --- | --- | --- |
| `endpoint` | `--endpoint` / `-e` | `AGENT_ENDPOINT` | 必填，非回环必须 HTTPS，不含用户信息、查询或片段。 |
| `token` | `--token` / `-t` | `AGENT_TOKEN` | 指定节点 Token；推荐私密文件，避免参数暴露。 |
| `config_file` | `--config` | `AGENT_CONFIG_FILE` | JSON 文件完整路径。 |
| `interval` | `--interval` / `-i` | `AGENT_INTERVAL` | 3 秒，至少 1 秒。 |
| `max_retries` | `--max-retries` / `-r` | `AGENT_MAX_RETRIES` | 3，连接逻辑的重试参数。 |
| `reconnect_interval` | `--reconnect-interval` / `-c` | `AGENT_RECONNECT_INTERVAL` | 5 秒，正整数。 |
| `info_report_interval` | `--info-report-interval` | `AGENT_INFO_REPORT_INTERVAL` | 5 分钟，正整数。 |
| `disable_auto_update` | `--disable-auto-update` | `AGENT_DISABLE_AUTO_UPDATE` | true，保持关闭自动更新。 |
| `disable_web_ssh` | `--disable-web-ssh` | `AGENT_DISABLE_WEB_SSH` | true，关闭全部对应远程控制能力。 |
| `ignore_unsafe_cert` | `--ignore-unsafe-cert` / `-u` | `AGENT_IGNORE_UNSAFE_CERT` | false，正常证书校验。 |
| `prefer_ip_version` | `--prefer-ip-version` | `AGENT_PREFER_IP_VERSION` | 空／`4`／`6`，连接 Server 的地址族偏好。 |
| `disable_compression` | `--disable-compression` | `AGENT_DISABLE_COMPRESSION` | false，v2 传输压缩开关。 |

## 采集口径

| JSON 字段 | CLI | 环境变量 | 解释 |
| --- | --- | --- | --- |
| `include_nics` | `--include-nics` | `AGENT_INCLUDE_NICS` | 逗号分隔，仅统计所选网卡，支持模式匹配。 |
| `exclude_nics` | `--exclude-nics` | `AGENT_EXCLUDE_NICS` | 逗号分隔，排除网卡。 |
| `include_mountpoints` | `--include-mountpoint` | `AGENT_INCLUDE_MOUNTPOINTS` | 分号分隔；CLI 为单数 mountpoint。 |
| `month_rotate` | `--month-rotate` | `AGENT_MONTH_ROTATE` | 默认为 0，不启用月重置；不是云商账单核算。 |
| `memory_include_cache` | `--memory-include-cache` | `AGENT_MEMORY_INCLUDE_CACHE` | false；是否把缓存／缓冲纳入使用量。 |
| `memory_report_raw_used` | `--memory-exclude-bcf` | `AGENT_MEMORY_REPORT_RAW_USED` | false；保留现有技术字段与 CLI 拼写，勿按名称猜公式。 |
| `custom_dns` | `--custom-dns` | `AGENT_CUSTOM_DNS` | 空时使用系统 DNS，不启用内置公共 DNS 列表。 |
| `enable_gpu` | `--gpu` | `AGENT_ENABLE_GPU` | false，详细 GPU 监测另行启用。 |
| `custom_ipv4` / `custom_ipv6` | `--custom-ipv4` / `--custom-ipv6` | `AGENT_CUSTOM_IPV4` / `AGENT_CUSTOM_IPV6` | 手动报告地址；不配置主机路由。 |
| `get_ip_addr_from_nic` | `--get-ip-addr-from-nic` | `AGENT_GET_IP_ADDR_FROM_NIC` | false，从网卡取报告地址。 |

`auto_discovery_key`／`AGENT_AUTO_DISCOVERY_KEY`／`--auto-discovery` 为自动发现凭据，不是普通 Token 的别名。受管生产优先使用明确节点身份。`memory_mode_available` 已弃用；新配置不要使用。其他诊断与平台参数以 `--help` 为准，不代表全部平台已完成运行验收。

## 最小文件

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

保存后按[维护流程](OPERATIONS.zh-CN.md)重启并核对新报告。不要直接打印整个文件来排错。配置、服务环境或日志中出现令牌，都应按秘密处理，不提交仓库。
