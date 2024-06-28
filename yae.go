package yae

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config holds the configuration parameters for retrieving a config.
type Env struct {
	Name         string      // Name of the config file
	Debug        bool        // Print debug messages
	Type         ConfigType  // Type of the config file ("json" or "yaml")
	Path         string      // Path to the config file
	EnvPrefix    string      // Prefix for environment variable names
	ConfigStruct interface{} // Struct to store the config values
	SkipFields   []string    // Fields to skip when loading from env
}

// EnvType represents the environment type.
type EnvType string

const (
	LOCAL EnvType = "local" // Local environment will use the keychain
	DEV   EnvType = "dev"   // Dev environment will use the keychain
	PROD  EnvType = "prod"  // Prod environment will use the config file or env vars
)

type ConfigType string

const (
	JSON ConfigType = "json"
	YAML ConfigType = "yaml"
)

var (
	CUSTOM ConfigType   = "" // This will search for whatever custom tag you specify
	log    *slog.Logger      // Logger for debug messages
)

func init() {
	log = logger(true)
}

// Get retrieves the configuration based on the specified environment type.
func Get(t EnvType, c *Env) error {
	log = logger(c.Debug)

	switch t {
	case DEV, LOCAL:
		log.Debug("loading config from keychain")
		return BuildDevEnv(c, nil)
	case PROD:
		log.Debug("loading config from file", "file", c.Name, "path", c.Path)
		return LoadConfig(c)
	default:
		return fmt.Errorf("unsupported environment type: %s", t)
	}
}

// LoadConfig loads the config from the file or falls back to environmental variables.
func LoadConfig(c *Env) error {
	// first check if the file exists, if not, try the full path, and finally fallback to env
	var confFile string

	f, fp := buildFilePath(c.Name, c.Path)
	if _, err := os.Stat(f); os.IsNotExist(err) {
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			log.Debug("config file not found, falling back to environment variables")
			if err := c.loadFromEnv(); err != nil {
				return fmt.Errorf("failed to load config from file and env: %w", err)
			}
			return nil
		} else {
			confFile = fp
		}
	} else {
		confFile = f
	}

	file, err := os.Open(confFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %s, error:%w", confFile, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", err)
	}

	switch strings.ToLower(string(c.Type)) {
	case string(JSON):
		err = json.Unmarshal(data, &c.ConfigStruct)
	case string(YAML):
		err = yaml.Unmarshal(data, c.ConfigStruct)
	default:
		return fmt.Errorf("unsupported file type: %s", c.Type)
	}

	return err
}

func buildFilePath(name, path string) (string, string) {
	if path == "" {
		path = "./"
	}
	return name, filepath.Join(path, name)
}

func (c *Env) loadFromEnv() error {
	log.Debug("loading config from env", "prefix", c.EnvPrefix)

	valueOf := reflect.ValueOf(c.ConfigStruct).Elem()
	typeOf := valueOf.Type()

	// we dont want to stop the loop if we skip a field so we add it to the slice and check at the end
	var envErr []error
	for i := 0; i < valueOf.NumField(); i++ {
		fieldType := typeOf.Field(i)
		log.Debug("loading field", "field", fieldType.Name, "type", fieldType.Type.String())
		if contains(c.SkipFields, fieldType.Name) {
			continue
		}

		field := valueOf.Field(i)
		envName := getEnvName(fieldType, c.Type, c.EnvPrefix)

		log.Debug("loading env", "env", envName)

		if envValue := os.Getenv(envName); envValue != "" {
			err := setField(field, envValue)
			if err != nil {
				return fmt.Errorf("failed to set field %s: %s", fieldType.Name, err)
			}
		} else {
			envErr = append(envErr, fmt.Errorf("env not found: %s", envName))
		}
	}

	// check the errors
	if len(envErr) > 0 {
		var sb strings.Builder
		for _, err := range envErr {
			log.Debug("error loading env", "error", err.Error())
			sb.WriteString(err.Error() + "\n")
		}
		return fmt.Errorf(sb.String())
	}
	return nil
}

func getEnvName(fieldType reflect.StructField, configType ConfigType, envPrefix string) string {
	var envName string
	if tag := fieldType.Tag.Get(string(configType)); tag != "" {
		envName = strings.ToUpper(tag)
	} else if tag := fieldType.Tag.Get("yaml"); tag != "" {
		envName = strings.ToUpper(tag)
	} else {
		envName = strings.ToUpper(fieldType.Tag.Get("env"))
		if envName == "" {
			envName = fieldType.Name
		}
	}

	if envPrefix != "" {
		envName = envPrefix + "_" + envName
	}

	return envName
}

// setField sets the value of a field in the struct based on its type.
func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse integer value: %s", err)
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse unsigned integer value: %s", err)
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse boolean value: %s", err)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float value: %s", err)
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported field type")
	}

	return nil
}

// GetKeys returns the keys for the struct.
func (c *Env) GetKeys() []string {
	var keys []string

	valueOf := reflect.ValueOf(c.ConfigStruct).Elem()
	typeOf := valueOf.Type()

	for i := 0; i < valueOf.NumField(); i++ {
		fieldType := typeOf.Field(i)
		if !contains(c.SkipFields, fieldType.Name) {
			if tag := fieldType.Tag.Get(string(c.Type)); tag != "" {
				keys = append(keys, tag)
			}
		}
	}
	return keys
}

// BuildDevEnv fills the values of the struct with the values from the keychain.
func BuildDevEnv(c *Env, secrets *Secrets, skipFields ...string) error {
	if secrets == nil {
		envKeys := c.GetKeys()
		secrets = GetConfig(c.Name, envKeys...)
	}
	secretMap := secrets.ToMap(skipFields...)

	valueOf := reflect.ValueOf(c.ConfigStruct).Elem()
	typeOf := valueOf.Type()

	for i := 0; i < valueOf.NumField(); i++ {
		field := valueOf.Field(i)
		fieldType := typeOf.Field(i)

		var tag string
		switch c.Type {
		case "json":
			tag = fieldType.Tag.Get("json")
		case "yaml":
			tag = fieldType.Tag.Get("yaml")
		case CUSTOM:
			tag = fieldType.Tag.Get(string(CUSTOM))
		default:
			continue
		}

		if val, ok := secretMap[tag]; ok {
			err := setField(field, val)
			if err != nil {
				return fmt.Errorf("failed to set field %s: %s", fieldType.Name, err)
			}
		}
	}
	return nil
}

func logger(debug bool) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: func() slog.Level {
		if debug {
			return slog.LevelDebug
		}
		return slog.LevelInfo
	}()})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
