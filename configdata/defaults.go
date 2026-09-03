package configdata

import (
	coreenv "goark.dev/goark/core/env"
)

const builtInDefaultSourceName = "config:built-in-defaults"

// builtInDefaults 返回没有外部配置文件时使用的最低优先级配置。
func builtInDefaults() map[string]any {
	return map[string]any{
		PropertyApplicationName: "goark",
		PropertyConfigName:      defaultBaseName,
	}
}

func builtInDefaultPropertySource() (coreenv.PropertySource, error) {
	return coreenv.NewMapPropertySource(builtInDefaultSourceName, builtInDefaults())
}

func builtInLoadedSource() LoadedSource {
	return LoadedSource{
		Name:    builtInDefaultSourceName,
		BuiltIn: true,
	}
}
