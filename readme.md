# May Monitor Agent

May Monitor 的主机采集端：采集 CPU、内存、磁盘和流量，向指定 Server 回传指标，并执行已配置的网络探测。它不是 Deploy 控制端、Web 后台或代理客户端。

[English](README.en.md) · [发行版](https://github.com/mxh110708/may-monitor-agent/releases/latest) · [安装与维护](docs/OPERATIONS.zh-CN.md) · [参数](docs/CONFIGURATION.zh-CN.md) · [安全边界](SECURITY.md)

## 当前范围

文档基线为 **0.1.0**，配套 [Server 0.1.0](https://github.com/mxh110708/may-monitor-server)。Linux amd64 是运行验收目标；Linux arm64 和 Windows amd64 附件属于交叉编译范围，不能视为对应平台服务安装及权限均已验收。

远程控制与自动更新**默认关闭**，前者同时限制终端、命令、文件、启动配置和版本切换。正常采集和探测不要求开启这些功能；不要复制旧上游示例把它们改成 false。

## 推荐安装

通过 [May VPS Deploy](https://github.com/mxh110708/may-vps-deploy) 的实例维护安装固定 Agent。先核对 SSH 主机、HTTPS Server 安装 UUID 及节点 Token，再建立专用用户、私有配置和独立 systemd 服务。

手动运维先从同次 Release 取得程序和 SHA-256，创建专用账户与私密配置。示例字段全部为占位符，不可直接当作凭据使用：

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

将配置保存在私有目录并限制文件权限，然后运行已经验证的程序：

```bash
./may-monitor-agent --config '/etc/may-monitor-agent/config.json'
```

Token 绑定指定节点，不能与其他安装共享，也不是管理员密码、只读凭据或 VPS root 密码。非回环端点要求 HTTPS，默认校验证书。不要把 Token 放进公开命令历史、URL、截图或问题日志。

## 配置优先级

从低到高为：**默认值 → 命令行 → 环境变量 → JSON 文件**。与许多 CLI 相反，命令行不是最高优先级；若 JSON 将某项写成 false，额外加一个 CLI true 不会修复文件内的覆盖。具体字段、默认值、单位和诊断命令见[参数参考](docs/CONFIGURATION.zh-CN.md)。

## 验收与维护

启动后核对实际版本、节点 UUID、Server UUID、连接状态、报告时间持续推进和历史持久化。网页能打开、进程存在或单次上报不能代替持续验收。探测需检查对应目标和协议；TCP 成功不证明 UDP 或真实业务质量。

修改端点或 Token 是配置维护，不是版本升级；先验证新凭据归属，再快照、写入、重启和核对新报告。旧 Token 已撤销时，恢复旧文件不能恢复认证。停止 Agent 不应操作 Server、Tunnel、代理或 SSH。

## 开发、来源与反馈

使用 Go 1.27.3 和 `go.sum`，先运行 `go mod verify` 及相关测试；详细开发流程见[运维与构建](docs/OPERATIONS.zh-CN.md)。默认安全工作流仅测试与编译，不自动发行或部署。`fake_agent.py` 是测试夹具，不用于生产指标或长测替代。

本项目源自 [Komari Agent](https://github.com/komari-monitor/komari-agent)，保留上游模块／协议技术名、版权和 [LICENSE](LICENSE)。上游安装脚本与发行工作流是停用的历史参考。反馈附版本、系统、时间和脱敏错误，不上传 Token 或完整启动配置。
