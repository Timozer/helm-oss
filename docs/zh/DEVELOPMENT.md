# 开发指南

本指南面向希望为 helm-oss 做出贡献或从源代码构建的开发者。

## 前置要求

- **Go 1.25+** - 构建项目所需
- **golangci-lint** - 代码质量检查工具(可选但推荐)
- **Docker** - 构建 Docker 镜像(可选)

## 从源代码构建

克隆仓库并构建二进制文件:

```bash
# 克隆仓库
git clone https://github.com/Timozer/helm-oss.git
cd helm-oss

# 构建二进制文件
go build -o helm-oss ./cmd/helm-oss

# 或使用优化选项构建(更小的二进制文件)
CGO_ENABLED=0 go build \
  -ldflags="-s -w" \
  -trimpath \
  -o helm-oss \
  ./cmd/helm-oss
```

## 运行测试

运行测试套件:

```bash
# 运行所有测试
go test ./...

# 运行测试并显示覆盖率
go test -cover ./...

# 运行测试并显示详细输出
go test -v ./...

# 运行测试并启用竞态检测
go test -race ./...
```

## 代码质量

本项目使用 [golangci-lint](https://golangci-lint.run/) 进行代码质量检查。

### 安装 golangci-lint

**macOS:**
```bash
brew install golangci-lint
```

**Linux:**
```bash
# 二进制安装
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**使用 Go:**
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 运行 Linter

```bash
# 运行所有配置的 linters
golangci-lint run

# 运行并自动修复(自动修复可修复的问题)
golangci-lint run --fix

# 在特定文件或目录上运行
golangci-lint run ./cmd/...
```

### Linter 配置

项目使用 `.golangci.yml` 进行 linter 配置。启用的主要 linters:

- **errcheck** - 检查未检查的错误
- **govet** - 检查 Go 源代码并报告可疑构造
- **staticcheck** - 高级静态分析
- **unused** - 检查未使用的代码
- **misspell** - 查找常见拼写错误
- **gocyclo** - 检查圈复杂度
- **goconst** - 查找可以作为常量的重复字符串
- **revive** - 快速、可配置、可扩展的 linter

## 项目结构

```
helm-oss/
├── cmd/
│   └── helm-oss/          # 主应用程序入口
├── internal/
│   ├── helmutil/          # Helm 工具
│   └── oss/               # OSS 存储实现
├── docs/                  # 文档
│   ├── en/                # 英文文档
│   └── zh/                # 中文文档
├── .github/
│   └── workflows/         # GitHub Actions CI/CD
├── .golangci.yml          # Linter 配置
├── .goreleaser.yml        # 发布配置
├── go.mod                 # Go 模块定义
├── Dockerfile             # Docker 镜像定义
└── plugin.yaml            # Helm 插件元数据
```

## 开发工作流

1. **修改代码**
2. **运行测试**确保没有破坏任何功能:
   ```bash
   go test ./...
   ```
3. **运行 linter** 检查代码质量:
   ```bash
   golangci-lint run
   ```
4. **构建**二进制文件验证编译:
   ```bash
   go build ./cmd/helm-oss
   ```
5. **手动测试**构建的二进制文件:
   ```bash
   ./helm-oss --help
   ```

## 构建 Docker 镜像

在本地构建 Docker 镜像:

```bash
# 为当前平台构建
docker build -t helm-oss:dev .

# 为特定平台构建
docker buildx build --platform linux/amd64 -t helm-oss:dev .

# 构建多平台镜像
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t helm-oss:dev \
  .
```

## CI/CD 流水线

项目使用 GitHub Actions 进行 CI/CD:

### CI 工作流(每次 push/PR)

并行运行 4 个任务:
- **Test** - 运行所有测试并生成覆盖率报告
- **Lint** - 运行 golangci-lint
- **Build** - 验证编译
- **Docker** - 验证 Docker 镜像构建

### Release 工作流(推送 tag 时)

自动执行:
1. 运行测试
2. 为多个平台构建二进制文件
3. 构建并推送 Docker 镜像(多架构)
4. 使用 GPG 签名构建产物
5. 创建 GitHub Release 并上传所有资源

## 贡献

1. Fork 仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 进行更改
4. 运行测试和 linter
5. 提交更改 (`git commit -m 'feat: add amazing feature'`)
6. 推送到分支 (`git push origin feature/amazing-feature`)
7. 打开 Pull Request

### 提交消息规范

我们遵循 [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - 新功能
- `fix:` - Bug 修复
- `docs:` - 文档更改
- `test:` - 测试更改
- `chore:` - 维护任务
- `refactor:` - 代码重构
- `perf:` - 性能改进

## 发布流程

1. 确保所有测试通过
2. 更新相关文件中的版本号
3. 创建并推送 tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. GitHub Actions 将自动:
   - 为所有平台构建二进制文件
   - 构建并推送 Docker 镜像
   - 创建 GitHub Release

## 故障排除

### 构建问题

如果遇到构建问题:

```bash
# 清理 Go 缓存
go clean -cache -modcache -i -r

# 下载依赖
go mod download

# 验证依赖
go mod verify

# 重新构建
go build ./cmd/helm-oss
```

### 测试问题

如果测试失败:

```bash
# 运行测试并显示详细输出
go test -v ./...

# 运行特定测试
go test -v -run TestName ./path/to/package
```

## 获取帮助

- 对于 bug,请打开 [issue](https://github.com/Timozer/helm-oss/issues)
- 对于问题,请开始 [discussion](https://github.com/Timozer/helm-oss/discussions)
- 请先检查现有的 issues 和 discussions
