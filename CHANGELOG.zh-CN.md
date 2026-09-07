# 变更日志

[English](CHANGELOG.md) | 中文

这里记录 Goark Boot 的重要变更。

## [未发布]

暂无未发布变更。

## [0.0.1] - 2026-09-07

### 新增

- 基于 Goark Core 的应用启动和有序自动配置。
- `app.yml`、`app.properties` 和 `app.toml` 配置数据发现。
- 基础、激活、包含、默认和分组 Profile 加载。
- 环境变量与命令行系统属性优先级。
- Resource 与可执行文件目录配置位置。
- 基于 Go 1.26 的跨平台测试、vet 和 race 门禁。

### 变更

- 将配置键统一到 `goark.*` 命名空间。
- 配置解析与应用启动保持独立包边界。
- 将所有实际使用的 `golang.org/x` 模块对齐到最新稳定版本。

[未发布]: https://github.com/goark-projects/goark-boot/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/goark-projects/goark-boot/releases/tag/v0.0.1
