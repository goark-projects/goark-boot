# goark boot

![goark](assets/goark-readme-logo.png)

`goark boot` is the bootstrap and convention layer for the Goark ecosystem. It
provides application startup, lifecycle wiring, config data loading, and module
assembly on top of the core [`goark`](https://goark.dev/goark) contracts.

## Design Boundary

- `goark.dev/goark/core/env` owns `Environment`, `PropertySource`, conversion,
  placeholder resolution, and configuration property contracts.
- `goark.dev/boot/configdata` owns file discovery, config precedence, profiles,
  and parser integration.
- `goark.dev/cli` generates configuration property binders and deterministic
  bean registration at compile time.
- Runtime startup does not scan structs or create dynamic proxies.

## Config Data

Directory discovery recognizes these base files only:

```text
app.yml
app.properties
app.toml
```

For an active profile such as `dev`, it recognizes:

```text
app-dev.yml
app-dev.properties
app-dev.toml
```

The `.yaml` format is accepted only when an exact file is supplied explicitly.
It is never included in directory discovery and no legacy base-name aliases are
recognized.

By default, boot searches these locations in ascending priority:

```text
<executable-dir>/resource
<working-dir>/resource
```

Missing files are allowed. If no file is found, boot installs the lowest
priority built-in defaults:

```properties
goark.application.name=goark
goark.config.name=app
```

### Startup Properties

| Property | Environment variable | Purpose |
| --- | --- | --- |
| `goark.config.location` | `GOARK_CONFIG_LOCATION` | Replaces the default locations |
| `goark.config.additional-location` | `GOARK_CONFIG_ADDITIONAL_LOCATION` | Appends higher-priority locations |
| `goark.config.name` | `GOARK_CONFIG_NAME` | Changes the discovered base name |
| `goark.profiles.active` | `GOARK_PROFILES_ACTIVE` | Activates one or more profiles |
| `goark.profiles.include` | - | Includes additional profiles |
| `goark.profiles.default` | - | Selects profiles when none are active |
| `goark.profiles.group.<name>` | - | Expands a named profile group |
| `goark.application.name` | - | Identifies the application |

Examples:

```bash
go run ./cmd/admin --goark.config.location=resource --goark.profiles.active=dev
go run ./cmd/admin --goark.config.location=resource/custom.yaml
```

`goark.config.location` accepts directories or exact files. An exact base file
also enables its sibling profile files, such as `custom-dev.yaml`.

### Precedence

From lower to higher priority:

1. Built-in defaults.
2. Base files, with later locations overriding earlier locations.
3. Profile files, with later profiles and locations overriding earlier ones.
4. Process environment variables.
5. Command-line properties.

Explicit Go options are applied after process environment and process command
line parsing. For the same directory and base name, the first existing format
wins in this order: `yml`, `properties`, `toml`.

### API

```go
package main

import (
	"context"

	"goark.dev/boot/configdata"
)

func main() {
	result, err := configdata.Load(
		context.Background(),
		configdata.WithProfiles("dev"),
	)
	if err != nil {
		panic(err)
	}

	name, _ := result.Environment.GetProperty("goark.application.name")
	_ = name
}
```

Parsing uses [`spf13/viper`](https://github.com/spf13/viper) and the Java
properties codec from [`go-viper/encoding`](https://github.com/go-viper/encoding).

## Installation

```bash
go get goark.dev/boot
```

## Development

Requirements:

- Go 1.25 or later
- Git

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Repository Layout

```text
.
|-- assets/       # README and brand assets
|-- configdata/   # Config discovery, parsing, profiles, and precedence
|-- go.mod        # Go module definition
|-- go.sum        # Go module checksums
|-- LICENSE       # Apache License 2.0
`-- README.md     # Project overview
```

## Related Repositories

- [`goark.dev/goark`](https://goark.dev/goark): core framework contracts.
- [`goark.dev/cli`](https://goark.dev/cli): compile-time code generation.

## License

`goark boot` is released under the Apache License 2.0. See [LICENSE](LICENSE).
