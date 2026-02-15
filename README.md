# helm-oss

[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

English | [中文](README_zh.md)

**helm-oss** is a Helm plugin that provides Alibaba Cloud OSS protocol support.

> **Based on [helm-s3](https://github.com/hypnoglow/helm-s3)** - This project is a fork of helm-s3, modified to support Alibaba Cloud OSS instead of AWS S3.

This allows you to have private or public Helm chart repositories hosted on Alibaba Cloud OSS.

The plugin supports Helm v3.

## Table of Contents

- [Install](#install)
  - [Docker Images](#docker-images)
- [Configuration](#configuration)
  - [OSS Access](#oss-access)
- [Usage](#usage)
  - [Init](#init)
  - [Push](#push)
  - [Delete](#delete)
  - [Download](#download)
  - [Reindex](#reindex)
- [Uninstall](#uninstall)
- [Advanced Features](#advanced-features)
  - [Relative chart URLs](#relative-chart-urls)
  - [Serving charts via HTTP](#serving-charts-via-http)
- [Documentation](#documentation)
- [Acknowledgments](#acknowledgments)
- [Contributing](#contributing)
- [License](#license)

## Install

The installation itself is simple as:

```bash
helm plugin install https://github.com/Timozer/helm-oss.git
```

You can install a specific release version:

```bash
helm plugin install https://github.com/Timozer/helm-oss.git --version 0.1.0
```

To use the plugin, you do not need any special dependencies. The installer will download versioned release with prebuilt binary from [GitHub releases](https://github.com/Timozer/helm-oss/releases).

### Docker Images

The plugin is also distributed as Docker images. You can build the image yourself:

```bash
docker build -t helm-oss:latest .
```

## Configuration

### OSS Access

To publish charts to buckets and to fetch from private buckets, you need to provide valid Alibaba Cloud OSS credentials.

You can configure credentials using environment variables:

```bash
export HELM_OSS_ACCESS_KEY_ID="your-access-key-id"
export HELM_OSS_ACCESS_KEY_SECRET="your-access-key-secret"
export HELM_OSS_REGION="oss-cn-hangzhou"
```

Optionally, you can also set:

```bash
export HELM_OSS_ENDPOINT="https://oss-cn-hangzhou.aliyuncs.com"
export HELM_OSS_SESSION_TOKEN="your-session-token"  # For STS
```

To minimize security issues, remember to configure your RAM user policies properly. As an example, a setup can provide only read access for users, and write access for a CI that builds and pushes charts to your repository.

## Usage

### Init

Initialize a new chart repository:

```bash
helm oss init oss://my-bucket/charts
```

This command generates an empty **index.yaml** and uploads it to the OSS bucket under `/charts` key.

### Push

To push a chart to the repository:

```bash
helm oss push ./mychart-0.1.0.tgz oss://my-bucket/charts
```

If you want to replace an existing chart, use the `--force` flag:

```bash
helm oss push --force ./mychart-0.1.0.tgz oss://my-bucket/charts
```

### Delete

To delete a specific chart version from the repository:

```bash
helm oss delete mychart --version 0.1.0 oss://my-bucket/charts
```

### Download

To download a chart from the repository:

```bash
helm oss download oss://my-bucket/charts/mychart-0.1.0.tgz
```

### Reindex

If your repository somehow became inconsistent or broken, you can use reindex to rebuild the index:

```bash
helm oss reindex oss://my-bucket/charts
```

This command will rebuild the index file from scratch.

## Uninstall

```bash
helm plugin uninstall oss
```

## Advanced Features

### Relative chart URLs

By default, `helm oss push` generates absolute URLs in `index.yaml`. This means that the chart URLs will point directly to OSS:

```yaml
entries:
  mychart:
  - urls:
    - oss://my-bucket/charts/mychart-0.1.0.tgz
```

However, the plugin now **always uses relative URLs** to support both direct OSS access and HTTP-based access (e.g., via CDN):

```yaml
entries:
  mychart:
  - urls:
    - mychart-0.1.0.tgz
```

This allows you to:
1. Access charts directly via OSS plugin: `helm pull oss://my-bucket/charts/mychart`
2. Access charts via HTTP/CDN: `helm pull https://my-cdn.com/charts/mychart`

### Serving charts via HTTP

You can enable public read access to your OSS bucket and serve charts via HTTP or CDN:

1. Set bucket ACL to public-read
2. Configure CDN (optional)
3. Add the HTTP repository to Helm:

```bash
helm repo add my-charts https://my-bucket.oss-cn-hangzhou.aliyuncs.com/charts
# or with CDN
helm repo add my-charts https://my-cdn.com/charts
```

4. Use charts as usual:

```bash
helm search repo my-charts
helm install myrelease my-charts/mychart
```

## Documentation

- **English**: [docs/en/](docs/en/)
  - [Development Guide](docs/en/DEVELOPMENT.md)
- **中文**: [docs/zh/](docs/zh/)
  - [开发指南](docs/zh/DEVELOPMENT.md)

## Acknowledgments

This project is based on [helm-s3](https://github.com/hypnoglow/helm-s3) by [Igor Zibarev](https://github.com/hypnoglow). We are deeply grateful for their excellent work that made this project possible.

Key differences from helm-s3:
- Uses Alibaba Cloud OSS SDK v2 instead of AWS S3 SDK
- Supports OSS-specific features and authentication
- Optimized for Alibaba Cloud infrastructure

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details.

This project includes code from [helm-s3](https://github.com/hypnoglow/helm-s3), which is also licensed under the MIT License.

---

**Original work**: Copyright (c) 2017 Igor Zibarev (helm-s3)  
**Modified work**: Copyright (c) 2024-2026 Timozer (helm-oss)
