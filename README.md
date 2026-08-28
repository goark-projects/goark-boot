# goark boot

![goark](assets/goark-readme-logo.png)

`goark boot` is the bootstrap and convention layer for the Goark ecosystem. It provides application startup, lifecycle wiring, configuration loading, and framework module assembly on top of the core [`goark`](https://goark.dev/goark) contracts.

The first boot feature is Spring Boot style config data loading for `app.yml`,
`app.yaml`, `app.toml`, and `app.properties`.

## Goals

- Provide a Go-native application bootstrap model comparable to Spring Boot's role in the Spring ecosystem.
- Keep startup composition explicit and observable.
- Avoid hidden global state and reflection-heavy control flow where plain Go contracts are sufficient.
- Support future extension modules for web, data access, messaging, configuration, lifecycle, and observability.
- Keep the boot layer separate from the core framework contracts.

## Config Data

Config file loading belongs to `boot`, not the `goark` core module. The core module exposes `config.Environment`, `config.PropertySource`, and `config.Binder`; boot owns file discovery, default locations, profile naming, and parser integration.

Supported formats:

- `yml`
- `yaml`
- `toml`
- `properties`

Default base files:

- `app.yml`
- `app.yaml`
- `app.toml`
- `app.properties`

Default profile files:

- `app-dev.yml`, `app-prod.yml`, `app-xxx.yml`
- `app-dev.yaml`, `app-prod.yaml`, `app-xxx.yaml`
- `app-dev.toml`, `app-prod.toml`, `app-xxx.toml`
- `app-dev.properties`, `app-prod.properties`, `app-xxx.properties`

`resource` is the default Goark application resource entrypoint. Without an
explicit location, boot looks for config files in:

```text
<executable-dir>/resource
<working-dir>/resource
```

The working directory resource root has higher priority. If no external config
file is found, boot installs a built-in default config source.

Startup can override the config location, config base name, and active profiles:

```bash
go run ./cmd/admin --goark.config.location=resource/app.yml --goark.profiles.active=dev
go run ./cmd/admin --spring.config.location=resource --spring.profiles.active=prod
```

Environment variables use the same model:

```bash
GOARK_CONFIG_LOCATION=resource
GOARK_CONFIG_NAME=app
GOARK_PROFILES_ACTIVE=dev
```

Spring-compatible environment variables are also recognized:

```bash
SPRING_CONFIG_LOCATION=resource
SPRING_CONFIG_NAME=app
SPRING_PROFILES_ACTIVE=dev
```

`goark.config.location` and `spring.config.location` accept either directories or
exact files such as `resource/app.yml`, `resource/app.yaml`, `resource/app.toml`,
or `resource/app.properties`. When a profile is active and an exact file is
provided, boot looks for a sibling profile file such as `app-dev.yml`.

Priority rules:

- Profile config overrides base config.
- Explicit command-line args override environment variables and defaults.
- Explicit Go options override process environment and command-line defaults.
- The application `resource` directory is used when no location is specified.
- Later active profiles override earlier active profiles.
- For the same logical file, format priority is `yml > yaml > toml > properties`.
- Missing config files are allowed.

Profiles can be provided explicitly with `configdata.WithProfiles("dev")`. If not provided, boot reads the first available profile key from base config:

- `goark.profiles.active`
- `spring.profiles.active`
- `profiles.active`

Example:

```go
package main

import (
	"context"

	"goark.dev/boot/configdata"
)

func main() {
	result, err := configdata.Load(context.Background(), configdata.WithProfiles("dev"))
	if err != nil {
		panic(err)
	}

	_ = result.Environment
}
```

Config parsing uses [`spf13/viper`](https://github.com/spf13/viper) with the Java properties codec from [`go-viper/encoding`](https://github.com/go-viper/encoding).

## Module

```bash
go get goark.dev/boot
```

## Repository Status

This repository is in active early development. Public APIs should be treated as unstable until the first tagged release.

## Development

Requirements:

- Go 1.25 or later
- Git

Useful commands:

```bash
go mod tidy
go test ./...
```

## Repository Layout

```text
.
├── assets/       # README and brand assets
├── configdata/   # Spring Boot style config data loading
├── go.mod        # Go module definition
├── go.sum        # Go module checksums
├── LICENSE       # Apache License 2.0
└── README.md     # Project overview
```

## Related Repositories

- [`goark.dev/goark`](https://goark.dev/goark): core framework contracts.
- [`goark.dev/boot`](https://goark.dev/boot): application bootstrap and convention layer.
- [`goark.dev/cli`](https://goark.dev/cli): scaffolding and compile-time code generation.

## License

`goark boot` is released under the Apache License 2.0. See [LICENSE](LICENSE) for details.
