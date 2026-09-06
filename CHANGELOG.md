# Changelog

English | [中文](CHANGELOG.zh-CN.md)

All notable changes to Goark Boot are recorded here.

## [Unreleased]

No unreleased changes.

## [0.0.1] - 2026-09-06

### Added

- Application startup and ordered auto-configuration on Goark Core.
- Config data discovery for `app.yml`, `app.properties`, and `app.toml`.
- Base, active, included, default, and grouped profile loading.
- Environment-variable and command-line system-property precedence.
- Resource and executable-directory configuration locations.
- Cross-platform CI with Go 1.26 test, vet, and race gates.

### Changed

- Standardized configuration keys under `goark.*` namespaces.
- Kept configuration parsing and application startup in separate packages.

[Unreleased]: https://github.com/goark-projects/goark-boot/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/goark-projects/goark-boot/releases/tag/v0.0.1
