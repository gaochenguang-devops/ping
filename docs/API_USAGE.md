# 接口调用说明

适用于没有 Web 访问入口、只能通过命令行、脚本或程序调用服务的场景。

## 基础信息

默认服务地址：

```text
http://localhost:8080
```

如果你是用自定义端口启动，例如：

```bash
./pingtool -port 8081
```

那么接口地址就是：

```text
http://localhost:8081
```

本文示例统一使用：

```text
http://localhost:8081
```

## 接口列表

- `GET /healthz`
- `POST /api/scan`
- `POST /api/scan-jobs`
- `GET /api/scan-jobs/{id}`
- `POST /api/scan-jobs/{id}/cancel`

## 请求字段

所有扫描接口都使用 JSON 请求体。

字段说明：

- `targets`: 多目标输入
- `cidr`: 网段输入
- `ports`: 端口输入
- `mode`: `ping` / `tcp` / `both`
- `count`: Ping 次数，范围 `1-4`
- `timeout_ms`: 超时毫秒，范围 `200-5000`
- `concurrency`: 并发数，范围 `1-256`
- `resolve_dns`: 是否先做 DNS 解析

说明：

- `mode = ping` 时，`ports` 可以为空
- `mode = tcp` 或 `mode = both` 时，`ports` 必填
- `targets` 和 `cidr` 至少要有一个非空

## 1. 健康检查

### curl

```bash
curl http://localhost:8081/healthz
```

### PowerShell

```powershell
Invoke-RestMethod http://localhost:8081/healthz | ConvertTo-Json
```

返回示例：

```json
{"status":"ok"}
```

## 2. 同步扫描

适合目标数量不大、希望一次请求直接拿到最终结果的场景。

### 2.1 只做 Ping

#### curl

```bash
curl -X POST http://localhost:8081/api/scan \
  -H "Content-Type: application/json" \
  -d '{
    "targets": "127.0.0.1\nbaidu.com",
    "mode": "ping",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 16,
    "resolve_dns": true
  }'
```

#### PowerShell

```powershell
$body = @{
  targets = "127.0.0.1`nbaidu.com"
  mode = "ping"
  count = 1
  timeout_ms = 1000
  concurrency = 16
  resolve_dns = $true
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/api/scan `
  -ContentType "application/json" `
  -Body $body | ConvertTo-Json -Depth 6
```

### 2.2 只做 TCP 端口探测

#### curl

```bash
curl -X POST http://localhost:8081/api/scan \
  -H "Content-Type: application/json" \
  -d '{
    "targets": "127.0.0.1 192.168.1.10",
    "ports": "22,80,443,8080-8082",
    "mode": "tcp",
    "timeout_ms": 1000,
    "concurrency": 32,
    "resolve_dns": false
  }'
```

#### PowerShell

```powershell
$body = @{
  targets = "127.0.0.1 192.168.1.10"
  ports = "22,80,443,8080-8082"
  mode = "tcp"
  timeout_ms = 1000
  concurrency = 32
  resolve_dns = $false
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/api/scan `
  -ContentType "application/json" `
  -Body $body | ConvertTo-Json -Depth 6
```

### 2.3 Ping + TCP

#### curl

```bash
curl -X POST http://localhost:8081/api/scan \
  -H "Content-Type: application/json" \
  -d '{
    "targets": "127.0.0.1",
    "ports": "22,80",
    "mode": "both",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 8,
    "resolve_dns": false
  }'
```

#### PowerShell

```powershell
$body = @{
  targets = "127.0.0.1"
  ports = "22,80"
  mode = "both"
  count = 1
  timeout_ms = 1000
  concurrency = 8
  resolve_dns = $false
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/api/scan `
  -ContentType "application/json" `
  -Body $body | ConvertTo-Json -Depth 6
```

### 2.4 使用网段和 IPv6

#### curl

```bash
curl -X POST http://localhost:8081/api/scan \
  -H "Content-Type: application/json" \
  -d '{
    "cidr": "192.168.1.0/30\n2001:db8::/126",
    "mode": "ping",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 32,
    "resolve_dns": false
  }'
```

## 3. 异步任务扫描

适合目标多、端口多、需要看进度、或者不希望一个请求阻塞太久的场景。

流程：

1. `POST /api/scan-jobs` 创建任务
2. 从返回中取到 `id`
3. 轮询 `GET /api/scan-jobs/{id}`
4. 状态变为 `done` 后读取结果
5. 如需取消，调用 `POST /api/scan-jobs/{id}/cancel`

### 3.1 创建任务

#### curl

```bash
curl -X POST http://localhost:8081/api/scan-jobs \
  -H "Content-Type: application/json" \
  -d '{
    "targets": "127.0.0.1 192.168.1.10",
    "ports": "22,80,443",
    "mode": "both",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 32,
    "resolve_dns": false
  }'
```

返回示例：

```json
{
  "id": "a1b2c3d4e5f6g7h8",
  "progress": {
    "total": 8,
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

### 3.2 查询任务

#### curl

```bash
curl http://localhost:8081/api/scan-jobs/a1b2c3d4e5f6g7h8
```

#### PowerShell

```powershell
Invoke-RestMethod http://localhost:8081/api/scan-jobs/a1b2c3d4e5f6g7h8 | ConvertTo-Json -Depth 6
```

### 3.3 PowerShell 异步完整示例

```powershell
$body = @{
  targets = "127.0.0.1 192.168.1.10"
  ports = "22,80,443"
  mode = "both"
  count = 1
  timeout_ms = 1000
  concurrency = 32
  resolve_dns = $false
} | ConvertTo-Json

$job = Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/api/scan-jobs `
  -ContentType "application/json" `
  -Body $body

do {
  Start-Sleep -Milliseconds 500
  $snapshot = Invoke-RestMethod "http://localhost:8081/api/scan-jobs/$($job.id)"
  $snapshot.progress | ConvertTo-Json
} while ($snapshot.progress.status -notin @("done", "error", "canceled"))

$snapshot.result | ConvertTo-Json -Depth 6
```

### 3.4 取消任务

#### curl

```bash
curl -X POST http://localhost:8081/api/scan-jobs/a1b2c3d4e5f6g7h8/cancel
```

#### PowerShell

```powershell
Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/api/scan-jobs/a1b2c3d4e5f6g7h8/cancel | ConvertTo-Json -Depth 6
```

## 4. Python 调用示例

需要：

```bash
pip install requests
```

### 4.1 同步扫描

```python
import requests

base_url = "http://localhost:8081"

payload = {
    "targets": "127.0.0.1\nbaidu.com",
    "ports": "80,443",
    "mode": "both",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 16,
    "resolve_dns": True,
}

resp = requests.post(f"{base_url}/api/scan", json=payload, timeout=120)
resp.raise_for_status()
data = resp.json()

print(data["summary"])
for item in data["results"]:
    print(item["target"], item["kind"], item["status"], item.get("endpoint"))
```

### 4.2 异步扫描

```python
import time
import requests

base_url = "http://localhost:8081"

payload = {
    "targets": "127.0.0.1 192.168.1.10",
    "ports": "22,80,443",
    "mode": "both",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 32,
    "resolve_dns": False,
}

job = requests.post(f"{base_url}/api/scan-jobs", json=payload, timeout=30).json()
job_id = job["id"]

while True:
    snapshot = requests.get(f"{base_url}/api/scan-jobs/{job_id}", timeout=30).json()
    progress = snapshot["progress"]
    print(progress["status"], progress["completed"], "/", progress["total"])

    if progress["status"] in {"done", "error", "canceled"}:
        break

    time.sleep(0.5)

if snapshot.get("result"):
    for item in snapshot["result"]["results"]:
        print(item["target"], item["kind"], item["status"], item.get("endpoint"))
```

## 5. 将结果保存为 JSON

### curl

```bash
curl -X POST http://localhost:8081/api/scan \
  -H "Content-Type: application/json" \
  -d '{
    "targets": "127.0.0.1",
    "ports": "80,443",
    "mode": "both",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 8,
    "resolve_dns": false
  }' > result.json
```

### PowerShell

```powershell
$body = @{
  targets = "127.0.0.1"
  ports = "80,443"
  mode = "both"
  count = 1
  timeout_ms = 1000
  concurrency = 8
  resolve_dns = $false
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri http://localhost:8081/api/scan `
  -ContentType "application/json" `
  -Body $body | ConvertTo-Json -Depth 6 | Set-Content result.json -Encoding UTF8
```

## 6. 将接口结果转成 CSV

如果没有 Web 页面，也可以直接把接口结果转成 CSV。

### Python 示例

```python
import csv
import requests

base_url = "http://localhost:8081"

payload = {
    "targets": "127.0.0.1",
    "ports": "22,80,443",
    "mode": "both",
    "count": 1,
    "timeout_ms": 1000,
    "concurrency": 8,
    "resolve_dns": False,
}

resp = requests.post(f"{base_url}/api/scan", json=payload, timeout=120)
resp.raise_for_status()
data = resp.json()

with open("result.csv", "w", newline="", encoding="utf-8-sig") as f:
    writer = csv.writer(f)
    writer.writerow([
        "目标", "类型", "端口/端点", "来源", "状态", "解析地址",
        "已收", "已发", "丢包率", "最小时延(ms)", "平均时延(ms)",
        "最大时延(ms)", "耗时(ms)", "说明",
    ])

    for item in data["results"]:
        writer.writerow([
            item.get("target", ""),
            item.get("kind", ""),
            item.get("endpoint") or item.get("port") or "ICMP",
            item.get("source", ""),
            item.get("status", ""),
            " | ".join(item.get("resolved_ips", [])),
            item.get("received", "") if item.get("kind") == "ping" else "",
            item.get("sent", "") if item.get("kind") == "ping" else "",
            item.get("loss_percent", "") if item.get("kind") == "ping" else "",
            item.get("min_latency_ms", ""),
            item.get("avg_latency_ms", ""),
            item.get("max_latency_ms", ""),
            item.get("duration_ms", ""),
            item.get("message", ""),
        ])
```

## 7. 常见调用场景

### 场景 1：只想批量测存活

用：

- `mode = ping`
- `ports = ""`

### 场景 2：目标禁 Ping，只关心端口

用：

- `mode = tcp`

### 场景 3：既想看存活，又想看端口开放

用：

- `mode = both`

### 场景 4：任务很大，不想阻塞一个请求

用：

- `POST /api/scan-jobs`
- `GET /api/scan-jobs/{id}`

## 8. 常见错误

### 400 Bad Request

一般是这些原因：

- `mode` 非法
- `ports` 格式错误
- `count` 超范围
- `timeout_ms` 超范围
- `concurrency` 超范围
- `targets` 和 `cidr` 同时为空

### 404

一般是任务 `id` 不存在。

### 500

一般是运行环境异常，例如：

- 系统没有 `ping`
- 系统 DNS 异常
- 内部执行失败

## 9. 建议

- 小任务优先用 `/api/scan`
- 大任务优先用 `/api/scan-jobs`
- 自动化脚本建议保存 JSON 原始结果
- 如果最终要给表格工具使用，建议自己按需要把 JSON 转成 CSV
