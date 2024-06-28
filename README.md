<div align="center">
<h1 align="center">
<img src='.docs/media/yae.png' width="200" />
<br>
</h1>
</div>

**yae** (Yet Another Env) is an environment getter that fits specific needs for secure and flexible configuration management. It is a simple package for storing and retrieving secrets using the system keyring for safer storage of credentials during local development. yae also supports retrieving environmental variables with or without a prefix. Currently, it supports `json` and `yaml` configurations but can easily be extended.

## Features

- Securely store and retrieve secrets using the system keyring.
- Load configuration from JSON and YAML files.
- Support for environment variables with or without a prefix.
- Fallback mechanism to load configuration from environment variables if the file is not found.
- Debug logging to help trace the loading process.

## Installation

To install yae, run the following command:

```shell
go get github.com/johnmikee/yae
```

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/johnmikee/yae"
)

type Config struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
	Baz int    `json:"baz"`
}

func main() {
	var cfg Config
	err := yae.Get(
		yae.EnvType(yae.DEV),
		&yae.Env{
			Name:         "YAE",
			ConfigStruct: &cfg,
			Type:         yae.JSON,
			Debug:        true,
		},
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("Bar: ", cfg.Bar)
	fmt.Println("Baz: ", cfg.Baz)
	fmt.Println("Foo: ", cfg.Foo)
}
```

![alt text](.docs/media/yae.gif)

## Configuration Options

The `Env` struct provides various configuration options:

- `Name`: Name of the config file.
- `Debug`: Enables debug messages when set to `true`.
- `Type`: Type of the config file (`json` or `yaml`).
- `Path`: Path to the config file.
- `EnvPrefix`: Prefix for environment variable names.
- `ConfigStruct`: Struct to store the config values.
- `SkipFields`: Fields to skip when loading from environment variables.

## Examples

### Loading Configuration from File

```go
package main

import (
	"fmt"
	"os"

	"github.com/johnmikee/yae"
)

type Config struct {
	APIKey      string `json:"api_key" yaml:"api_key"`
	DatabaseURL string `json:"database_url" yaml:"database_url"`
}

func main() {
	var cfg Config
	err := yae.Get(
		yae.PROD,
		&yae.Env{
			Name:         "config.json",
			ConfigStruct: &cfg,
			Type:         yae.JSON,
			Debug:        true,
		},
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("APIKey: ", cfg.APIKey)
	fmt.Println("DatabaseURL: ", cfg.DatabaseURL)
}
```

### Loading Configuration from Environment Variables with Prefix

```go
package main

import (
	"fmt"
	"os"

	"github.com/johnmikee/yae"
)

type Config struct {
	APIKey      string `json:"api_key" yaml:"api_key"`
	DatabaseURL string `json:"database_url" yaml:"database_url"`
}

func main() {
	var cfg Config
	err := yae.Get(
		yae.PROD,
		&yae.Env{
			Name:         "config.json",
			ConfigStruct: &cfg,
			Type:         yae.JSON,
			EnvPrefix:    "YAE",
			Debug:        true,
		},
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("APIKey: ", cfg.APIKey)
	fmt.Println("DatabaseURL: ", cfg.DatabaseURL)
}
```

### Handling Fallback to Environment Variables

If the configuration file is not found, `yae` will automatically fall back to loading configuration from environment variables. This is useful for scenarios where the configuration file is not available, but the necessary environment variables are set.

### Debug Logging

Enable debug logging to get detailed information about the configuration loading process. Set the `Debug` field to `true` in the `Env` struct.

```go
env := &yae.Env{
	Name:         "config.json",
	ConfigStruct: &cfg,
	Type:         yae.JSON,
	Debug:        true,
}
```


