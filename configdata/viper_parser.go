package configdata

import (
	"bytes"
	"os"

	"github.com/go-viper/encoding/javaproperties"
	"github.com/spf13/viper"
	arkerrors "goark.dev/goark/errors"
)

// ViperParser 使用 Viper 解析 YAML、TOML 和 Java properties。
type ViperParser struct{}

// NewViperParser 创建 Viper 配置解析器。
func NewViperParser() *ViperParser {
	return &ViperParser{}
}

// ParseFile 解析配置文件并返回扁平化键值表。
func (p *ViperParser) ParseFile(path string, format Format) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, arkerrors.Wrapf(arkerrors.CodeNotFound, err, "failed to read config file %q", path)
	}

	registry, err := newCodecRegistry()
	if err != nil {
		return nil, err
	}
	parser := viper.NewWithOptions(viper.WithCodecRegistry(registry))
	parser.SetConfigType(string(format))
	if err := parser.ReadConfig(bytes.NewReader(data)); err != nil {
		return nil, arkerrors.Wrapf(arkerrors.CodeInvalidArgument, err, "failed to parse config file %q", path)
	}
	return flattenSettings(parser.AllSettings()), nil
}

func newCodecRegistry() (*viper.DefaultCodecRegistry, error) {
	registry := viper.NewCodecRegistry()
	if err := registry.RegisterCodec(string(FormatProperties), &javaproperties.Codec{KeyDelimiter: "."}); err != nil {
		return nil, arkerrors.Wrap(arkerrors.CodeInvalidArgument, err, "failed to register properties codec")
	}
	return registry, nil
}
