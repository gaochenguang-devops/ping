# GitHub 自动交叉编译与上传

本文档说明如何在代码推送到 GitHub 后，自动执行测试、交叉编译，并把产物上传到 GitHub Actions Artifacts；在打版本标签时，再自动创建 GitHub Release 并上传产物。

## 已生成的配置

工作流文件：

- [.github/workflows/build-release.yml](../.github/workflows/build-release.yml)

## 这套流程会做什么

### 普通推送

当你把代码推送到 GitHub 后，工作流会自动：

1. 拉取代码
2. 安装 Go
3. 运行 `go test ./...`
4. 交叉编译以下目标：
   - `windows/amd64`
   - `windows/arm64`
   - `linux/amd64`
   - `linux/arm64`
   - `darwin/amd64`
   - `darwin/arm64`
5. 打包产物
6. 上传到当前 Actions Run 的 Artifacts

### 打版本标签

如果你推送的是 `v*` 标签，例如：

```bash
git tag v1.0.0
git push origin v1.0.0
```

除了上面的构建和 Artifact 上传之外，还会额外：

1. 自动创建 GitHub Release
2. 把所有构建好的包上传为 Release Assets
3. 自动生成 Release Notes

## 触发条件

当前工作流会在这些事件触发：

- `push`
- `pull_request`
- `workflow_dispatch`

其中：

- `push`：自动测试、构建、上传 Artifacts
- `pull_request`：自动测试、构建、上传 Artifacts
- `workflow_dispatch`：手动触发
- `refs/tags/v*`：自动发布 Release

## 产物格式

为了方便下载和分发，当前打包策略是：

- Windows：`.zip`
- Linux：`.tar.gz`
- macOS：`.tar.gz`

包内默认包含：

- 主程序
- `README.md`
- `docs/API.md`
- `docs/API_USAGE.md`
- `docs/CROSS_COMPILE.md`

## 产物命名

命名规则：

```text
pingtool-<version>-<goos>-<goarch>
```

例如：

- `pingtool-sha-1a2b3c4-windows-amd64.zip`
- `pingtool-v1.0.0-linux-amd64.tar.gz`
- `pingtool-v1.0.0-darwin-arm64.tar.gz`

说明：

- 普通分支推送使用 `sha-<短提交号>`
- 标签发布使用标签名，例如 `v1.0.0`

## 如何使用

### 方式 1：普通推送后下载 Artifacts

1. 把代码推送到 GitHub
2. 打开仓库的 `Actions`
3. 进入对应的 Workflow Run
4. 在页面底部下载 `Artifacts`

适合：

- 内部测试
- 临时分发
- 验证某次提交的构建结果

### 方式 2：打标签后下载 Release Assets

执行：

```bash
git tag v1.0.0
git push origin v1.0.0
```

然后：

1. 打开仓库 `Releases`
2. 找到新建的 `v1.0.0`
3. 下载对应平台的二进制包

适合：

- 正式发版
- 给别人下载固定版本
- 留存历史版本

## GitHub 仓库要求

### 1. 开启 GitHub Actions

仓库必须允许执行 GitHub Actions。

### 2. 默认权限

当前工作流已经在 YAML 中配置了：

- 全局 `contents: read`
- Release Job 使用 `contents: write`

正常情况下不需要额外配置 Secret。

### 3. Release 上传依赖

当前使用：

- `softprops/action-gh-release@v2`

这个 Action 会直接使用 GitHub 提供的 `GITHUB_TOKEN` 创建 Release 并上传产物。

## 推荐的发布流程

### 日常开发

直接推送分支：

```bash
git push origin main
```

然后到 Actions 下载 Artifacts。

### 正式版本

先确认主分支代码可用，再打标签：

```bash
git checkout main
git pull
git tag v1.0.0
git push origin main --tags
```

然后到 Releases 页面下载版本包。

## 如何改成只在主分支触发

如果你不想每个分支都跑，可以把：

```yaml
on:
  push:
```

改成：

```yaml
on:
  push:
    branches:
      - main
      - master
```

如果还想保留标签发布，可以写成：

```yaml
on:
  push:
    branches:
      - main
      - master
    tags:
      - "v*"
```

## 如何增加或减少目标平台

修改文件中的矩阵：

```yaml
strategy:
  matrix:
    include:
```

例如只保留常见平台：

- `windows/amd64`
- `linux/amd64`
- `darwin/arm64`

## 如何调整 Go 版本

修改：

```yaml
env:
  GO_VERSION: "1.25.0"
```

## 如何只上传二进制，不附带文档

删除或注释这个打包步骤里的复制逻辑：

```bash
cp README.md "package/${ARCHIVE_BASE}/"
cp docs/API.md docs/API_USAGE.md docs/CROSS_COMPILE.md "package/${ARCHIVE_BASE}/docs/"
```

## 如何只上传 Release，不保留 Artifacts

删除或注释：

```yaml
- name: Upload Workflow Artifact
  uses: actions/upload-artifact@v4
```

不过通常不建议删，因为 Artifacts 对调试构建过程很有用。

## 常见问题

### 1. 为什么 push 后没有创建 Release

因为当前配置只有在 `v*` 标签推送时才会创建 Release。

例如：

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 2. 为什么有 Artifacts 但没有 Release

这是正常行为。普通分支推送只上传 Artifacts，不发正式 Release。

### 3. 为什么 macOS 包可以生成，但目标机运行提示安全限制

因为从非 macOS 系统交叉编译出来的程序默认没有签名。正式分发时建议在 macOS 上做签名和公证。

### 4. 为什么工作流失败在测试阶段

因为构建前会先跑：

```bash
go test ./...
```

测试失败时不会继续发布产物。
