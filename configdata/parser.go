package configdata

// Parser 将配置文件解析为扁平键值表。
type Parser interface {
	ParseFile(path string, format Format) (map[string]string, error)
}
