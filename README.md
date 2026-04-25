# PingTool

一个基于 Go 的网络检测工具，提供 Web 界面和 HTTP API，支持批量目标检测、网段展开、IPv4/IPv6、TCP 端口探测、任务进度、结果筛选和 CSV 导出。

## 功能概览

- 支持 `Ping`、`TCP`、`Ping + TCP` 三种检测模式
- 支持多目标批量输入
- 支持 IPv4 / IPv6 单地址
- 支持 IPv4 / IPv6 CIDR 网段展开
- 支持 IPv4 范围
  - `192.168.1.1-10`
  - `192.168.1.1-192.168.1.100`
- 支持 IPv6 完整范围
  - `2001:db8::1-2001:db8::10`
- 支持端口列表和范围
  - `80`
  - `80,443,8080-8082`
- 支持并发扫描和实时进度
- 支持结果按关键字、状态、来源、类型筛选
- 支持导出当前筛选结果为 CSV
- 支持 Windows / Linux / macOS

## 界面能力

前端页面提供这些能力：

- 检测模式切换
- 多目标输入
- 网段输入
- 端口输入
- Ping 次数、超时、并发数配置
- DNS 解析开关
- 扫描进度条和统计卡片
- 结果汇总
- 结果表格
- CSV 导出

导出 CSV 时，`已收` 和 `已发` 会分成两列，避免 Excel 把 `1/1` 自动识别成日期。

## 运行环境

- Go `1.25.0` 或更高版本
- 目标系统需要存在可执行的 `ping` 命令
  - Windows：系统自带
  - Linux：通常来自 `iputils-ping`
  - macOS：系统自带

## 快速开始

### 1. 直接运行

```bash
go run .
```

默认监听：

```text
http://localhost:8080
```

### 2. 先编译再运行

```bash
go build -o pingtool.exe .
./pingtool.exe
```

Windows 下：

```powershell
go build -o pingtool.exe .
.\pingtool.exe
```

## 监听地址与端口配置

项目支持通过命令行参数或环境变量修改监听地址。

### 命令行参数

- `-port`
- `-host`
- `-addr`

示例：

```powershell
.\pingtool.exe -port 8090
.\pingtool.exe -host 0.0.0.0 -port 8090
.\pingtool.exe -addr 127.0.0.1:8090
```

```bash
./pingtool -port 8090
./pingtool -host 0.0.0.0 -port 8090
./pingtool -addr 127.0.0.1:8090
```

### 环境变量

- `PINGTOOL_ADDR`
- `PINGTOOL_HOST`
- `PINGTOOL_PORT`
- `PINGTOOL_OPEN_BROWSER`

示例：

```powershell
$env:PINGTOOL_PORT="8090"
$env:PINGTOOL_OPEN_BROWSER="false"
.\pingtool.exe
```

```bash
export PINGTOOL_PORT=8090
export PINGTOOL_OPEN_BROWSER=false
./pingtool
```

### 优先级

命令行参数优先级高于环境变量：

1. `-addr`
2. `-host` / `-port`
3. `PINGTOOL_ADDR`
4. `PINGTOOL_HOST` / `PINGTOOL_PORT`
5. 默认值 `:8080`

## 使用说明

### 目标输入

`targets` 支持：

- 域名：`baidu.com`
- IPv4：`192.168.1.10`
- IPv6：`2001:db8::10`
- IPv4 短范围：`192.168.1.1-10`
- IPv4 完整范围：`192.168.1.1-192.168.1.100`
- IPv6 完整范围：`2001:db8::1-2001:db8::10`

多个目标可使用：

- 换行
- 空格
- 逗号
- 分号

### 网段输入

`cidr` 支持：

- IPv4 CIDR：`192.168.1.0/24`
- IPv6 CIDR：`2001:db8::/126`

### 端口输入

`ports` 支持：

- 单端口：`22`
- 多端口：`22,80,443`
- 范围：`8000-8010`

多个端口也支持空格、换行、逗号、分号混合分隔。

### 检测模式

- `ping`：只做 Ping
- `tcp`：只做 TCP 端口探测
- `both`：对每个目标同时生成 Ping 任务和 TCP 任务

## 参数限制

后端当前限制如下：

- `count`: `1-4`
- `timeout_ms`: `200-5000`
- `concurrency`: `1-256`
- 最大目标数：`4096`
- 最大端口数：`256`
- 最大任务数：`65536`

任务数计算方式：

- `ping`：目标数
- `tcp`：目标数 `x` 端口数
- `both`：目标数 `x` (`1 + 端口数`)

## HTTP API

如果当前环境没有 Web 访问入口，直接看接口调用文档：

- [docs/API.md](docs/API.md)
- [docs/API_USAGE.md](docs/API_USAGE.md)

### 健康检查

`GET /healthz`

响应示例：

```json
{
  "status": "ok"
}
```

### 同步扫描

`POST /api/scan`

请求示例：

```json
{
  "targets": "127.0.0.1\nbaidu.com",
  "cidr": "192.168.1.0/30",
  "ports": "80,443",
  "mode": "both",
  "count": 1,
  "timeout_ms": 1000,
  "concurrency": 32,
  "resolve_dns": true
}
```

响应结构：

```json
{
  "summary": {
    "total": 0,
    "reachable": 0,
    "unreachable": 0,
    "errors": 0,
    "avg_latency_ms": 0,
    "elapsed_ms": 0,
    "started_at": "2026-04-25 20:00:00",
    "finished_at": "2026-04-25 20:00:01"
  },
  "results": [
    {
      "target": "127.0.0.1",
      "source": "manual",
      "kind": "ping",
      "status": "reachable",
      "reachable": true,
      "resolved_ips": ["127.0.0.1"],
      "sent": 1,
      "received": 1,
      "loss_percent": 0,
      "min_latency_ms": 1,
      "avg_latency_ms": 1,
      "max_latency_ms": 1,
      "duration_ms": 2,
      "message": "平均 1.00 ms"
    }
  ]
}
```

### 异步任务扫描

#### 创建任务

`POST /api/scan-jobs`

返回任务快照：

```json
{
  "id": "job-id",
  "progress": {
    "total": 0,
    "completed": 0,
    "reachable": 0,
    "unreachable": 0,
    "errors": 0,
    "percent": 0,
    "status": "queued",
    "message": "任务已创建，等待开始"
  }
}
```

#### 查询任务

`GET /api/scan-jobs/{id}`

#### 取消任务

`POST /api/scan-jobs/{id}/cancel`

## 结果字段说明

### summary

- `total`: 结果总数
- `reachable`: 可达数量
- `unreachable`: 不可达数量
- `errors`: 错误数量
- `avg_latency_ms`: 平均时延
- `elapsed_ms`: 总耗时

### result

- `target`: 原始目标
- `source`: 目标来源，`manual` / `cidr`
- `kind`: 结果类型，`ping` / `tcp`
- `endpoint`: TCP 端点，如 `127.0.0.1:80`
- `port`: TCP 端口
- `status`: `reachable` / `unreachable` / `error`
- `resolved_ips`: DNS 解析结果
- `sent` / `received`: Ping 发包和收包
- `loss_percent`: 丢包率
- `min_latency_ms` / `avg_latency_ms` / `max_latency_ms`: 时延数据
- `duration_ms`: 单项任务耗时
- `message`: 文本说明

## 跨平台说明

项目已经按平台区分 `ping` 调用方式：

- Windows：使用 `ping -n -w`
- Linux：使用 `ping -c -i -W`
- macOS：使用 `ping -c`

TCP 端口探测基于 Go 标准库 `net.Dialer`，天然跨平台。

交叉编译请看：

- [docs/CROSS_COMPILE.md](docs/CROSS_COMPILE.md)

## 项目结构

```text
.
├── main.go
├── README.md
├── docs/
│   └── CROSS_COMPILE.md
└── internal/
    ├── app/
    │   ├── app.go
    │   ├── config.go
    │   └── config_test.go
    ├── scan/
    │   ├── types.go
    │   ├── service.go
    │   ├── targets.go
    │   ├── ports.go
    │   ├── ping.go
    │   ├── tcp.go
    │   ├── parser.go
    │   └── *_test.go
    └── web/
        ├── handlers.go
        ├── jobs.go
        └── templates/
            └── index.html
```

### 目录职责

- `internal/app`
  - 启动入口
  - 监听地址解析
  - 打开浏览器
- `internal/scan`
  - 请求归一化
  - 目标展开
  - 端口展开
  - Ping 探测
  - TCP 探测
  - 结果汇总
- `internal/web`
  - HTTP 路由
  - 页面渲染
  - 异步任务管理

## 常用命令

### 运行测试

```bash
go test ./...
```

### 构建

```bash
go build -o pingtool.exe .
```

### 本地启动

```bash
go run .
```

### 指定端口启动

```bash
go run . -port 8081
```

## 常见问题

### 1. 为什么 Ping 失败但 TCP 正常

可能是目标禁用了 ICMP 回显，但端口仍然开放。这种场景建议切换到 `tcp` 或 `both` 模式。

### 2. 为什么 CSV 打开后格式不对

当前导出的 CSV 已对 `已收` / `已发` 做了拆列处理，优先使用 Excel 或 WPS 的“按逗号导入”方式打开。

### 3. 为什么启动时报端口占用

换一个端口启动：

```bash
go run . -port 8090
```

### 4. 为什么提示未找到 ping 命令

目标系统缺少 `ping`，或者该命令不在 `PATH` 中。安装系统自带的 `ping` 工具后再试。

## 适合的使用场景

- 局域网主机巡检
- 多网段存活探测
- 服务器端口快速探测
- IPv4 / IPv6 连通性验证
- 交付前的批量网络自检
