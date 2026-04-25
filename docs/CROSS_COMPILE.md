# 交叉编译说明

本文档说明如何为本项目生成 Windows、Linux、macOS 的可执行文件。

## 适用前提

- Go `1.25.0` 或更高版本
- 当前项目是纯 Go 项目，默认不依赖 CGO
- 目标机器运行时需要有可用的 `ping` 命令

建议在交叉编译时显式设置：

```text
CGO_ENABLED=0
```

这样产物更稳定，也更适合跨平台分发。

## 支持的常见目标

- Windows
  - `windows/amd64`
  - `windows/arm64`
- Linux
  - `linux/amd64`
  - `linux/arm64`
- macOS
  - `darwin/amd64`
  - `darwin/arm64`

## 输出目录建议

建议统一输出到 `dist/`：

```text
dist/
├── pingtool-windows-amd64.exe
├── pingtool-windows-arm64.exe
├── pingtool-linux-amd64
├── pingtool-linux-arm64
├── pingtool-darwin-amd64
└── pingtool-darwin-arm64
```

## PowerShell 示例

先创建输出目录：

```powershell
New-Item -ItemType Directory -Force dist | Out-Null
```

### Windows amd64

```powershell
$env:CGO_ENABLED="0"
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -o .\dist\pingtool-windows-amd64.exe .
```

### Windows arm64

```powershell
$env:CGO_ENABLED="0"
$env:GOOS="windows"
$env:GOARCH="arm64"
go build -o .\dist\pingtool-windows-arm64.exe .
```

### Linux amd64

```powershell
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o .\dist\pingtool-linux-amd64 .
```

### Linux arm64

```powershell
$env:CGO_ENABLED="0"
$env:GOOS="linux"
$env:GOARCH="arm64"
go build -o .\dist\pingtool-linux-arm64 .
```

### macOS amd64

```powershell
$env:CGO_ENABLED="0"
$env:GOOS="darwin"
$env:GOARCH="amd64"
go build -o .\dist\pingtool-darwin-amd64 .
```

### macOS arm64

```powershell
$env:CGO_ENABLED="0"
$env:GOOS="darwin"
$env:GOARCH="arm64"
go build -o .\dist\pingtool-darwin-arm64 .
```

### 编译结束后清理环境变量

```powershell
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
```

## Bash 示例

先创建输出目录：

```bash
mkdir -p dist
```

### Windows amd64

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/pingtool-windows-amd64.exe .
```

### Windows arm64

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o dist/pingtool-windows-arm64.exe .
```

### Linux amd64

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/pingtool-linux-amd64 .
```

### Linux arm64

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/pingtool-linux-arm64 .
```

### macOS amd64

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o dist/pingtool-darwin-amd64 .
```

### macOS arm64

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o dist/pingtool-darwin-arm64 .
```

## 一次构建全部常见目标

### PowerShell

```powershell
New-Item -ItemType Directory -Force dist | Out-Null

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Output = ".\\dist\\pingtool-windows-amd64.exe" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Output = ".\\dist\\pingtool-windows-arm64.exe" },
    @{ GOOS = "linux";   GOARCH = "amd64"; Output = ".\\dist\\pingtool-linux-amd64" },
    @{ GOOS = "linux";   GOARCH = "arm64"; Output = ".\\dist\\pingtool-linux-arm64" },
    @{ GOOS = "darwin";  GOARCH = "amd64"; Output = ".\\dist\\pingtool-darwin-amd64" },
    @{ GOOS = "darwin";  GOARCH = "arm64"; Output = ".\\dist\\pingtool-darwin-arm64" }
)

foreach ($target in $targets) {
    $env:CGO_ENABLED = "0"
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    go build -o $target.Output .
}

Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
```

### Bash

```bash
mkdir -p dist

for target in \
  "windows amd64 .exe" \
  "windows arm64 .exe" \
  "linux amd64 ''" \
  "linux arm64 ''" \
  "darwin amd64 ''" \
  "darwin arm64 ''"
do
  set -- $target
  GOOS="$1"
  GOARCH="$2"
  SUFFIX="$3"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
    go build -o "dist/pingtool-${GOOS}-${GOARCH}${SUFFIX}" .
done
```

## 运行示例

### Windows

```powershell
.\pingtool-windows-amd64.exe -port 8081
```

### Linux

```bash
chmod +x pingtool-linux-amd64
./pingtool-linux-amd64 -port 8081
```

### macOS

```bash
chmod +x pingtool-darwin-arm64
./pingtool-darwin-arm64 -port 8081
```

## 产物验证建议

编译后建议至少做这几步：

### 1. 查看帮助和启动

```bash
./pingtool-linux-amd64 -port 8081
```

或：

```powershell
.\pingtool-windows-amd64.exe -port 8081
```

### 2. 检查健康接口

```bash
curl http://localhost:8081/healthz
```

返回应为：

```json
{"status":"ok"}
```

### 3. 检查目标系统是否有 ping

```bash
ping 127.0.0.1
```

## 平台注意事项

### Windows

- 直接生成 `.exe`
- 通常不需要额外依赖

### Linux

- 某些最小化发行版可能默认没有 `ping`
- 需要安装 `iputils-ping`

### macOS

- 从非 macOS 系统交叉编译得到的二进制默认不会签名
- 首次运行时可能被系统安全策略拦截
- 如果用于正式分发，建议在 macOS 上做签名和公证

## 常见问题

### 1. 交叉编译成功，但目标机运行失败

先确认：

- 架构是否匹配
- 目标机是否存在 `ping`
- 是否有执行权限

### 2. Linux 或 macOS 上没有执行权限

执行：

```bash
chmod +x <binary>
```

### 3. 启动时报端口被占用

换一个端口：

```bash
./pingtool-linux-amd64 -port 8090
```

### 4. 为什么推荐 `CGO_ENABLED=0`

因为本项目不依赖 CGO，显式关闭后更利于跨平台构建和分发。
