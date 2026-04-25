# API 文档

本文档描述 PingTool 的 HTTP API 定义，包括接口路径、请求字段、响应结构、状态码和约束。

如果你更关心怎么在命令行、PowerShell 或 Python 里直接调用，请看：

- [API_USAGE.md](API_USAGE.md)

## 基本信息

- 协议：`HTTP`
- 请求体格式：`application/json`
- 响应格式：`application/json; charset=utf-8`
- 默认监听地址：`http://localhost:8080`

常见自定义地址示例：

- `http://localhost:8081`
- `http://127.0.0.1:8090`

## 设计说明

- 所有扫描类接口都返回 JSON
- `POST /api/scan` 是同步接口，等结果全部完成后直接返回
- `POST /api/scan-jobs` 是异步接口，先返回任务 ID，再通过轮询获取进度和结果
- 请求体启用了严格 JSON 解码，不允许未知字段
- 请求体大小上限为 `1 MiB`

## 通用对象

### ScanRequest

请求对象：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `targets` | `string` | 否 | 多目标输入，支持域名、IPv4、IPv6、范围 |
| `cidr` | `string` | 否 | 网段输入，支持 IPv4 / IPv6 CIDR |
| `ports` | `string` | 条件必填 | TCP 模式下必填，支持单端口和范围 |
| `mode` | `string` | 否 | `ping` / `tcp` / `both`，默认 `ping` |
| `count` | `int` | 否 | Ping 次数，范围 `1-4` |
| `timeout_ms` | `int` | 否 | 超时毫秒，范围 `200-5000` |
| `concurrency` | `int` | 否 | 并发数，范围 `1-256` |
| `resolve_dns` | `bool` | 否 | 是否先做 DNS 解析，默认 `true` |

约束：

- `targets` 和 `cidr` 至少要有一个非空
- `mode = ping` 时，`ports` 应为空
- `mode = tcp` 或 `mode = both` 时，`ports` 必填

### ScanSummary

汇总对象：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `total` | `int` | 结果总数 |
| `reachable` | `int` | 可达数量 |
| `unreachable` | `int` | 不可达数量 |
| `errors` | `int` | 错误数量 |
| `avg_latency_ms` | `number` | 平均时延，可能为空 |
| `elapsed_ms` | `int64` | 总耗时 |
| `started_at` | `string` | 开始时间，格式 `YYYY-MM-DD HH:MM:SS` |
| `finished_at` | `string` | 结束时间，格式 `YYYY-MM-DD HH:MM:SS` |

### ScanResult

单项结果对象：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `target` | `string` | 原始目标 |
| `source` | `string` | 目标来源，`manual` / `cidr` / 组合 |
| `kind` | `string` | 结果类型，`ping` / `tcp` |
| `endpoint` | `string` | TCP 端点，例如 `127.0.0.1:80` |
| `port` | `int` | TCP 端口 |
| `status` | `string` | `reachable` / `unreachable` / `error` |
| `reachable` | `bool` | 是否可达 |
| `resolved_ips` | `string[]` | 解析到的 IP 列表 |
| `sent` | `int` | Ping 已发包数 |
| `received` | `int` | Ping 已收包数 |
| `loss_percent` | `number` | 丢包率 |
| `min_latency_ms` | `number` | 最小时延 |
| `avg_latency_ms` | `number` | 平均时延 |
| `max_latency_ms` | `number` | 最大时延 |
| `duration_ms` | `int64` | 单项任务耗时 |
| `message` | `string` | 文本说明 |

说明：

- `kind = ping` 时，`sent` / `received` / `loss_percent` 有意义
- `kind = tcp` 时，`endpoint` / `port` 有意义
- 某些字段在当前结果类型下可能为空

### ScanResponse

同步扫描响应对象：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `summary` | `ScanSummary` | 汇总 |
| `results` | `ScanResult[]` | 结果列表 |

### ScanProgress

异步任务进度对象：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `total` | `int` | 总任务数 |
| `completed` | `int` | 已完成数 |
| `reachable` | `int` | 可达数 |
| `unreachable` | `int` | 不可达数 |
| `errors` | `int` | 错误数 |
| `percent` | `number` | 进度百分比 |
| `status` | `string` | `queued` / `running` / `done` / `error` / `canceled` |
| `started_at` | `string` | 开始时间 |
| `finished_at` | `string` | 结束时间 |
| `message` | `string` | 进度说明 |

### ScanJobSnapshot

异步任务快照对象：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | `string` | 任务 ID |
| `progress` | `ScanProgress` | 进度信息 |
| `result` | `ScanResponse` | 任务完成后的结果 |
| `error` | `string` | 错误信息 |

## 状态枚举

### status

- `reachable`
- `unreachable`
- `error`

### mode

- `ping`
- `tcp`
- `both`

### kind

- `ping`
- `tcp`

### progress.status

- `queued`
- `running`
- `done`
- `error`
- `canceled`

## 接口明细

---

## 1. 健康检查

### 请求

- 方法：`GET`
- 路径：`/healthz`

### 成功响应

状态码：`200 OK`

```json
{
  "status": "ok"
}
```

### 失败响应

如果使用了不支持的方法：

状态码：`405 Method Not Allowed`

```json
{
  "error": "method not allowed"
}
```

---

## 2. 同步扫描

### 请求

- 方法：`POST`
- 路径：`/api/scan`
- Content-Type：`application/json`

### 请求示例

```json
{
  "targets": "127.0.0.1\nbaidu.com",
  "cidr": "",
  "ports": "80,443",
  "mode": "both",
  "count": 1,
  "timeout_ms": 1000,
  "concurrency": 32,
  "resolve_dns": true
}
```

### 成功响应

状态码：`200 OK`

```json
{
  "summary": {
    "total": 3,
    "reachable": 2,
    "unreachable": 1,
    "errors": 0,
    "avg_latency_ms": 3,
    "elapsed_ms": 26,
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
      "duration_ms": 1,
      "message": "平均 1.00 ms"
    },
    {
      "target": "127.0.0.1",
      "source": "manual",
      "kind": "tcp",
      "endpoint": "127.0.0.1:80",
      "port": 80,
      "status": "unreachable",
      "reachable": false,
      "resolved_ips": ["127.0.0.1"],
      "duration_ms": 0,
      "message": "端口关闭"
    }
  ]
}
```

### 错误响应

#### 参数错误

状态码：`400 Bad Request`

```json
{
  "error": "TCP 端口探测需要输入要检测的端口"
}
```

#### 服务内部错误

状态码：`500 Internal Server Error`

```json
{
  "error": "internal error"
}
```

---

## 3. 创建异步扫描任务

### 请求

- 方法：`POST`
- 路径：`/api/scan-jobs`
- Content-Type：`application/json`

### 请求示例

```json
{
  "targets": "127.0.0.1 192.168.1.10",
  "ports": "22,80,443",
  "mode": "both",
  "count": 1,
  "timeout_ms": 1000,
  "concurrency": 32,
  "resolve_dns": false
}
```

### 成功响应

状态码：`202 Accepted`

```json
{
  "id": "f7d2d67c9f2d8f4a",
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

### 错误响应

状态码：`400 Bad Request` 或 `500 Internal Server Error`

```json
{
  "error": "error message"
}
```

---

## 4. 查询异步任务

### 请求

- 方法：`GET`
- 路径：`/api/scan-jobs/{id}`

### 成功响应

状态码：`200 OK`

运行中示例：

```json
{
  "id": "f7d2d67c9f2d8f4a",
  "progress": {
    "total": 10,
    "completed": 4,
    "reachable": 2,
    "unreachable": 2,
    "errors": 0,
    "percent": 40,
    "status": "running",
    "started_at": "2026-04-25 20:00:00",
    "message": "检测进行中"
  }
}
```

完成示例：

```json
{
  "id": "f7d2d67c9f2d8f4a",
  "progress": {
    "total": 10,
    "completed": 10,
    "reachable": 6,
    "unreachable": 4,
    "errors": 0,
    "percent": 100,
    "status": "done",
    "started_at": "2026-04-25 20:00:00",
    "finished_at": "2026-04-25 20:00:03",
    "message": "检测完成"
  },
  "result": {
    "summary": {
      "total": 10,
      "reachable": 6,
      "unreachable": 4,
      "errors": 0,
      "avg_latency_ms": 2,
      "elapsed_ms": 3000,
      "started_at": "2026-04-25 20:00:00",
      "finished_at": "2026-04-25 20:00:03"
    },
    "results": []
  }
}
```

### 错误响应

任务不存在：

状态码：`404 Not Found`

```json
{
  "error": "job not found"
}
```

---

## 5. 取消异步任务

### 请求

- 方法：`POST`
- 路径：`/api/scan-jobs/{id}/cancel`

### 成功响应

状态码：`200 OK`

```json
{
  "id": "f7d2d67c9f2d8f4a",
  "progress": {
    "status": "canceled",
    "message": "检测已取消"
  }
}
```

### 错误响应

任务不存在：

状态码：`404 Not Found`

```json
{
  "error": "job not found"
}
```

---

## 错误码与状态码

| 状态码 | 场景 |
| --- | --- |
| `200` | 查询成功、同步扫描成功、取消成功 |
| `202` | 异步任务创建成功 |
| `400` | 请求 JSON 非法、参数校验失败 |
| `404` | 任务不存在 |
| `405` | 请求方法不支持 |
| `500` | 服务内部错误 |

## 参数校验规则

### count

- 默认值：`1`
- 范围：`1-4`

### timeout_ms

- 默认值：`1000`
- 范围：`200-5000`

### concurrency

- 默认值：`32`
- 范围：`1-256`

### mode

- 默认值：`ping`
- 可选值：`ping` / `tcp` / `both`

### targets / cidr

- 至少一个非空
- 最大目标数：`4096`

### ports

- 最大端口数：`256`
- 端口范围：`1-65535`

### 最大任务数

- 最大任务数：`65536`

计算方式：

- `ping`：目标数
- `tcp`：目标数 `x` 端口数
- `both`：目标数 `x` (`1 + 端口数`)

## 输入格式说明

### targets

支持：

- 域名
- IPv4
- IPv6
- IPv4 范围
- IPv6 完整范围

分隔符支持：

- 换行
- 空格
- 逗号
- 分号

### cidr

支持：

- IPv4 CIDR
- IPv6 CIDR

### ports

支持：

- 单端口：`80`
- 多端口：`80,443`
- 范围：`8080-8082`

## 非目标能力

以下能力当前不是后端独立接口：

- CSV 导出

说明：

- Web 页面中的 CSV 导出是前端基于 `/api/scan` 或 `/api/scan-jobs/{id}` 的结果自行生成的
- 如果没有 Web，可自行把 JSON 结果转成 CSV

## 推荐使用方式

### 小规模任务

用：

- `POST /api/scan`

### 大规模任务

用：

- `POST /api/scan-jobs`
- `GET /api/scan-jobs/{id}`

### 自动化脚本

建议：

- 使用异步任务接口
- 保留原始 JSON 结果
- 由脚本自行决定是否转 CSV
