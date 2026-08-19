# goark boot

![goark](assets/goark-readme-logo.png)

`goark boot` is the bootstrap and convention layer for the Goark ecosystem. It is intended to provide application startup, lifecycle wiring, configuration loading, and framework module assembly on top of the core [`goark`](https://github.com/goark-projects/goark) contracts.

The project is in its initial public bootstrap stage. This repository currently defines the root module and project conventions; boot APIs will be added as the core framework stabilizes.

## Goals

- Provide a Go-native application bootstrap model comparable to Spring Boot's role in the Spring ecosystem.
- Keep startup composition explicit and observable.
- Avoid hidden global state and reflection-heavy control flow where plain Go contracts are sufficient.
- Support future extension modules for web, data access, messaging, configuration, lifecycle, and observability.
- Keep the boot layer separate from the core framework contracts.

## Module

```bash
go get github.com/goark-projects/boot
```

## Repository Status

This repository is an early skeleton. Public APIs should be treated as unstable until the first tagged release.

## Development

Requirements:

- Go 1.25 or later
- Git

Useful commands:

```bash
go mod tidy
go list -m
```

## Repository Layout

```text
.
├── assets/      # README and brand assets
├── go.mod       # Go module definition
├── LICENSE      # Apache License 2.0
└── README.md    # Project overview
```

## Related Repositories

- [`goark-projects/goark`](https://github.com/goark-projects/goark): core framework contracts.
- [`goark-projects/boot`](https://github.com/goark-projects/boot): application bootstrap and convention layer.

## License

`goark boot` is released under the Apache License 2.0. See [LICENSE](LICENSE) for details.
