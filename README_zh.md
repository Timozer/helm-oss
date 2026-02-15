# helm-oss

[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

[English](README.md) | 中文

**helm-oss** 是一个提供阿里云 OSS 协议支持的 Helm 插件。

> **基于 [helm-s3](https://github.com/hypnoglow/helm-s3)** - 本项目是 helm-s3 的分叉版本，修改为支持阿里云 OSS 而非 AWS S3。

这允许您拥有托管在阿里云 OSS 上的私有或公共 Helm Chart 仓库。

该插件支持 Helm v3。

## 目录

- [安装](#安装)
  - [Docker 镜像](#docker-镜像)
- [配置](#配置)
  - [OSS 访问凭证](#oss-访问凭证)
- [使用](#使用)
  - [初始化](#初始化)
  - [推送](#推送)
  - [删除](#删除)
  - [下载](#下载)
  - [重建索引](#重建索引)
- [卸载](#卸载)
- [高级功能](#高级功能)
  - [相对 Chart URL](#相对-chart-url)
  - [通过 HTTP 提供 Chart](#通过-http-提供-chart)
- [文档](#文档)
- [致谢](#致谢)
- [贡献](#贡献)
- [许可证](#许可证)

## 安装

安装非常简单：

```bash
helm plugin install https://github.com/Timozer/helm-oss.git
```

您可以安装特定的发布版本：

```bash
helm plugin install https://github.com/Timozer/helm-oss.git --version 0.1.0
```

使用该插件不需要任何特殊的依赖。安装程序会从 [GitHub releases](https://github.com/Timozer/helm-oss/releases) 下载带有预编译二进制文件的版本化发布包。

### Docker 镜像

该插件也以 Docker 镜像形式分发。您可以自己构建镜像：

```bash
docker build -t helm-oss:latest .
```

## 配置

### OSS 访问凭证

要发布 Chart 到 Bucket 或从私有 Bucket 获取 Chart，您需要提供有效的阿里云 OSS 凭证。

您可以使用环境变量配置凭证：

```bash
export HELM_OSS_ACCESS_KEY_ID="your-access-key-id"
export HELM_OSS_ACCESS_KEY_SECRET="your-access-key-secret"
export HELM_OSS_REGION="oss-cn-hangzhou"
```

可选配置：

```bash
export HELM_OSS_ENDPOINT="https://oss-cn-hangzhou.aliyuncs.com"
export HELM_OSS_SESSION_TOKEN="your-session-token"  # 用于 STS
```

为了最大限度地减少安全问题，请记住正确配置您的 RAM 用户策略。例如，可以设置为用户仅提供读取访问权限，而为构建和推送 Chart 到仓库的 CI 提供写入访问权限。

## 使用

### 初始化

初始化一个新的 Chart 仓库：

```bash
helm oss init oss://my-bucket/charts
```

该命令会生成一个空的 **index.yaml** 并将其上传到 OSS Bucket 的 `/charts` 路径下。

### 推送

要推送一个 Chart 到仓库：

```bash
helm oss push ./mychart-0.1.0.tgz oss://my-bucket/charts
```

如果您想替换现有的 Chart，请使用 `--force` 标志：

```bash
helm oss push --force ./mychart-0.1.0.tgz oss://my-bucket/charts
```

### 删除

要从仓库中删除特定的 Chart 版本：

```bash
helm oss delete mychart --version 0.1.0 oss://my-bucket/charts
```

### 下载

要从仓库中下载 Chart：

```bash
helm oss download oss://my-bucket/charts/mychart-0.1.0.tgz
```

### 重建索引

如果您的仓库因某种原因变得不一致或损坏，您可以使用 reindex 重建索引：

```bash
helm oss reindex oss://my-bucket/charts
```

该命令将从头开始重建索引文件。

## 卸载

```bash
helm plugin uninstall oss
```

## 高级功能

### 相对 Chart URL

默认情况下，`helm oss push` 在 `index.yaml` 中生成绝对 URL。这意味着 Chart URL 将直接指向 OSS：

```yaml
entries:
  mychart:
  - urls:
    - oss://my-bucket/charts/mychart-0.1.0.tgz
```

然而，插件现在**总是使用相对 URL**，以支持直接 OSS 访问和基于 HTTP 的访问（例如，通过 CDN）：

```yaml
entries:
  mychart:
  - urls:
    - mychart-0.1.0.tgz
```

这允许您：
1. 通过 OSS 插件直接访问 Chart：`helm pull oss://my-bucket/charts/mychart`
2. 通过 HTTP/CDN 访问 Chart：`helm pull https://my-cdn.com/charts/mychart`

### 通过 HTTP 提供 Chart

您可以启用 OSS Bucket 的公共读访问权限，并通过 HTTP 或 CDN 提供 Chart：

1. 设置 Bucket ACL 为公共读（public-read）
2. 配置 CDN（可选）
3. 将 HTTP 仓库添加到 Helm：

```bash
helm repo add my-charts https://my-bucket.oss-cn-hangzhou.aliyuncs.com/charts
# 或者使用 CDN
helm repo add my-charts https://my-cdn.com/charts
```

4. 像往常一样使用 Chart：

```bash
helm search repo my-charts
helm install myrelease my-charts/mychart
```

## 文档

- **English**: [docs/en/](docs/en/)
  - [Development Guide](docs/en/DEVELOPMENT.md)
- **中文**: [docs/zh/](docs/zh/)
  - [开发指南](docs/zh/DEVELOPMENT.md)

## 致谢

本项目基于 [Igor Zibarev](https://github.com/hypnoglow) 的 [helm-s3](https://github.com/hypnoglow/helm-s3)。我们非常感谢他们出色的工作，使本项目成为可能。

与 helm-s3 的主要区别：
- 使用阿里云 OSS SDK v2 替代 AWS S3 SDK
- 支持 OSS 特有的功能和认证
- 针对阿里云基础设施进行了优化

## 贡献

欢迎贡献！请随时提交 Pull Request。

## 许可证

MIT 许可证 - 详情请见 [LICENSE](LICENSE) 文件。

本项目包含来自 [helm-s3](https://github.com/hypnoglow/helm-s3) 的代码，该代码也基于 MIT 许可证授权。

---

**Original work**: Copyright (c) 2017 Igor Zibarev (helm-s3)  
**Modified work**: Copyright (c) 2024-2026 Timozer (helm-oss)
