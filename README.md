# helm-oss

[![GitHub release (latest by date)](https://img.shields.io/github/v/release/Timozer/helm-oss)](https://github.com/Timozer/helm-oss/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/Timozer/helm-oss)](https://goreportcard.com/report/github.com/Timozer/helm-oss)
[![CI](https://github.com/Timozer/helm-oss/actions/workflows/ci.yml/badge.svg)](https://github.com/Timozer/helm-oss/actions/workflows/ci.yml)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://github.com/Timozer/helm-oss/blob/main/LICENSE)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/helm-oss)](https://artifacthub.io/packages/search?repo=helm-oss)

English | [中文](https://github.com/Timozer/helm-oss/blob/main/README_zh.md)

**helm-oss** is a Helm plugin that provides Alibaba Cloud OSS protocol support.

> **Based on [helm-s3](https://github.com/hypnoglow/helm-s3)** - This project is a fork of helm-s3, modified to support Alibaba Cloud OSS instead of AWS S3.

This allows you to have private or public Helm chart repositories hosted on Alibaba Cloud OSS.

The plugin supports Helm v3.

## Buy Me a Coffee

If you find this project useful, you can buy me a coffee!

|                                                   WeChat Pay                                                   |                                                     Alipay                                                     |                                                     PayPal                                                     |
| :------------------------------------------------------------------------------------------------------------: | :------------------------------------------------------------------------------------------------------------: | :------------------------------------------------------------------------------------------------------------: |
| <img src="https://raw.githubusercontent.com/Timozer/helm-oss/main/docs/images/sponsor/wechat.jpg" width="200"> | <img src="https://raw.githubusercontent.com/Timozer/helm-oss/main/docs/images/sponsor/alipay.jpg" width="200"> | <img src="https://raw.githubusercontent.com/Timozer/helm-oss/main/docs/images/sponsor/paypal.jpg" width="200"> |

*Please add a remark `helm-oss` when transferring, thank you for your support!*

## Table of Contents

- [helm-oss](#helm-oss)
  - [Table of Contents](#table-of-contents)
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

Install the latest version:

```bash
helm plugin install https://github.com/Timozer/helm-oss.git
```

Install a specific release version:

```bash
helm plugin install https://github.com/Timozer/helm-oss.git --version 0.1.0
```

To use the plugin, you do not need any special dependencies. The installer will download versioned release with prebuilt binary from [GitHub releases](https://github.com/Timozer/helm-oss/releases).

### Docker Images

The plugin is also distributed as Docker images. You can pull the image from [Docker Hub](https://hub.docker.com/r/zhenyuwang94/helm-oss):

```bash
docker pull zhenyuwang94/helm-oss:latest
```

Or you can build the image yourself:

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
export HELM_OSS_ENDPOINT="https://oss-cn-hangzhou.aliyuncs.com"
```

Optionally, you can also set:

```bash
export HELM_OSS_SESSION_TOKEN="your-session-token"  # For STS
```

To minimize security issues, remember to configure your RAM user policies properly. As an example, a setup can provide only read access for users, and write access for a CI that builds and pushes charts to your repository.

### Configuration File

You can also configure the plugin using a YAML file located at `~/.config/helm_plugin_oss.yaml`.

Example `~/.config/helm_plugin_oss.yaml`:

```yaml
endpoint: "https://oss-cn-hangzhou.aliyuncs.com"
region: "oss-cn-hangzhou"
accessKeyID: "your-access-key-id"
accessKeySecret: "your-access-key-secret"
# sessionToken: "your-session-token" # Optional, for STS
```

> **Note**: Environment variables take precedence over the configuration file.

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
helm pull oss://my-bucket/charts/mychart-0.1.0.tgz
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

The plugin **uses relative URLs** to support both direct OSS access and HTTP-based access (e.g., via CDN):

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

- **English**: [docs/en/](https://github.com/Timozer/helm-oss/blob/main/docs/en/)
  - [Development Guide](https://github.com/Timozer/helm-oss/blob/main/docs/en/DEVELOPMENT.md)
- **中文**: [docs/zh/](https://github.com/Timozer/helm-oss/blob/main/docs/zh/)
  - [开发指南](https://github.com/Timozer/helm-oss/blob/main/docs/zh/DEVELOPMENT.md)

## Acknowledgments

This project is based on [helm-s3](https://github.com/hypnoglow/helm-s3) by [Igor Zibarev](https://github.com/hypnoglow). We are deeply grateful for their excellent work that made this project possible.

Key differences from helm-s3:
- Uses Alibaba Cloud OSS SDK v2 instead of AWS S3 SDK
- Supports OSS-specific features and authentication
- Optimized for Alibaba Cloud infrastructure

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](https://github.com/Timozer/helm-oss/blob/main/LICENSE) file for details.

This project includes code from [helm-s3](https://github.com/hypnoglow/helm-s3), which is also licensed under the MIT License.

---

**Original work**: Copyright (c) 2017 Igor Zibarev (helm-s3)  
**Modified work**: Copyright (c) 2026 Timozer (helm-oss)
