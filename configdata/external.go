package configdata

import (
	"os"
	"strings"
)

const (
	propertySpringConfigLocation = "spring.config.location"
	propertySpringConfigName     = "spring.config.name"
	envSpringConfigLocation      = "SPRING_CONFIG_LOCATION"
	envSpringConfigName          = "SPRING_CONFIG_NAME"
	envSpringProfilesActive      = "SPRING_PROFILES_ACTIVE"
)

var (
	configLocationKeys = []string{PropertyConfigLocation, propertySpringConfigLocation}
	configNameKeys     = []string{PropertyConfigName, propertySpringConfigName}
	profilesActiveKeys = []string{PropertyProfilesActive, profileKeySpring, profileKeyShort}
)

// applyEnvironment 应用进程环境变量中的启动配置。
func applyEnvironment(options *Options) error {
	values := make(map[string]string, 3)
	if value := firstNonEmptyEnv(EnvConfigLocation, envSpringConfigLocation); value != "" {
		values[PropertyConfigLocation] = value
	}
	if value := firstNonEmptyEnv(EnvConfigName, envSpringConfigName); value != "" {
		values[PropertyConfigName] = value
	}
	if value := firstNonEmptyEnv(EnvProfilesActive, envSpringProfilesActive); value != "" {
		values[PropertyProfilesActive] = value
	}
	return applyExternalProperties(options, values)
}

// applyCommandLine 应用命令行中的启动配置。
func applyCommandLine(options *Options, args []string) error {
	return applyExternalProperties(options, commandLineProperties(args))
}

func applyExternalProperties(options *Options, values map[string]string) error {
	if options == nil {
		return nil
	}
	if value, ok := firstProperty(values, configLocationKeys); ok {
		if err := WithLocations(value)(options); err != nil {
			return err
		}
	}
	if value, ok := firstProperty(values, configNameKeys); ok {
		if err := WithBaseName(value)(options); err != nil {
			return err
		}
	}
	if value, ok := firstProperty(values, profilesActiveKeys); ok {
		if err := WithProfiles(value)(options); err != nil {
			return err
		}
	}
	return nil
}

func commandLineProperties(args []string) map[string]string {
	values := make(map[string]string)
	for index := 0; index < len(args); index++ {
		key, value, ok := commandLineProperty(args, index)
		if !ok {
			continue
		}
		values[key] = value
		if !strings.Contains(args[index], "=") && index+1 < len(args) && args[index+1] == value {
			index++
		}
	}
	return values
}

func commandLineProperty(args []string, index int) (string, string, bool) {
	arg := strings.TrimSpace(args[index])
	switch {
	case strings.HasPrefix(arg, "--"):
		return longOptionProperty(args, index, strings.TrimPrefix(arg, "--"))
	case strings.HasPrefix(arg, "-D"):
		return systemProperty(strings.TrimPrefix(arg, "-D"))
	default:
		return "", "", false
	}
}

func longOptionProperty(args []string, index int, body string) (string, string, bool) {
	key, value, found := strings.Cut(body, "=")
	key = strings.TrimSpace(key)
	if !isKnownExternalKey(key) {
		return "", "", false
	}
	if found {
		return key, strings.TrimSpace(value), true
	}
	if index+1 >= len(args) || strings.HasPrefix(strings.TrimSpace(args[index+1]), "-") {
		return key, "", true
	}
	return key, strings.TrimSpace(args[index+1]), true
}

func systemProperty(body string) (string, string, bool) {
	key, value, found := strings.Cut(body, "=")
	key = strings.TrimSpace(key)
	if !found || !isKnownExternalKey(key) {
		return "", "", false
	}
	return key, strings.TrimSpace(value), true
}

func firstProperty(values map[string]string, keys []string) (string, bool) {
	for _, key := range keys {
		if value, ok := values[key]; ok {
			return value, true
		}
	}
	return "", false
}

func firstNonEmptyEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func isKnownExternalKey(key string) bool {
	for _, group := range [][]string{configLocationKeys, configNameKeys, profilesActiveKeys} {
		for _, item := range group {
			if key == item {
				return true
			}
		}
	}
	return false
}
