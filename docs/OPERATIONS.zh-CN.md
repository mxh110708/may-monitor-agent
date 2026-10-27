# Agent 安装、维护与构建

基线：0.1.0。先阅读[首页](../readme.md)和[参数](CONFIGURATION.zh-CN.md)。本手册以 Debian 12／13 amd64 的受管运行范围为主；Windows 交叉编译不代表服务安装已验收。

## 受管安装

1. 在 Server 创建节点，取得该节点专用 Token；确认 HTTPS 和安装 UUID。
2. 在 Deploy 选择正确 SSH 实例，核对指纹和组件归属。未知或不完整布局不能当成已受管实例。
3. 审阅固定 Agent 资产、摘要、账户、私密配置和独立服务。默认关闭远程控制和自动更新。
4. 写入前先通过 Server 的专用凭据核验确认 Client UUID 与 Server UUID，不能以匿名主页可访问替代。
5. 启动后观察连续新报告和历史。系统信息／采样间隔是不同字段，首次信息等待不应误判为零负载。

## 手动运行与权限

校验发行版和架构后，把程序放在专用账户不可修改的目录，配置放在私有目录，令牌文件只允许需要的账户读取。systemd 工作目录要稳定，以保留本地统计状态；不要从临时下载目录长期运行。

命令结构：

```bash
./may-monitor-agent --version
./may-monitor-agent --help
./may-monitor-agent --config '/etc/may-monitor-agent/config.json'
```

普通采集使用必要最小权限。ICMP 探测可能需要额外系统能力，不能为了一个探测失败就把所有服务改成 root 或放开所有远程控制。测试因权限跳过 ICMP 不等于已经完成 ICMP 验收。

## 核对数据与重连

确认节点 UUID、地址和名称的对应，检查报告时间向前推进、网卡选择、挂载点选择、内存口径与 Server 历史。接口网卡过滤／月流量重置会影响数字，变更前记录当前口径。

HTTPS 失败先检查 DNS、证书、系统时间、地址族与反向代理 WS；Token 错误先查节点归属与撤销状态。HTTP 回退与 WS 状态有各自租约，在线不等于报告新鲜。

`prefer_ip_version` 只影响连接 Server 优先地址族，不保证所有探测改用该地址族。IPv6 有地址不等于有可用路由，VPS 的 ATT IPv6 池配置属于服务商网络资料，Agent 不负责申请或替换租约。

## 变更和恢复

先保存本组件程序、配置、统计状态和实际启停状态。修改端点或 Token 前核验目标身份；不借同版本升级绕过配置维护。更新后重新检查实际版本、Token、报告及历史。

连接中断或提交确认丢失时核对原事务，不能重新安装来覆盖。恢复旧配置前确认旧 Token 仍有效，若已撤销则需显式修复凭据。停止／卸载只处理当前 Agent，不动 Server、Tunnel、代理和 SSH。

## 开发验证

从本仓库固定提交和 `go.sum` 构建，Go 版本为 1.27.3：

```bash
go mod verify
go test -count=1 -timeout=5m ./...
go test -race -count=1 ./server ./update ./cmd ./ws
go build -trimpath -o may-monitor-agent .
```

本地普通构建没有自动设置发行版本／来源元数据，不能冒充正式附件。正式构建需记录源码、工具链、架构、大小和摘要，并运行固定漏洞检查。Windows／ARM 构建与对应平台实机验收分开记录。

`fake_agent.py` 只用于隔离服务的合成采样和协议夹具，不证明真实主机采集、长期占用或生产连接。只在自己创建的测试环境使用，测试后清理节点与私密状态，不能让它冒充真实在线节点。
