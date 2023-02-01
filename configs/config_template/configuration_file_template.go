package config_template

import _ "embed"

var (
	//go:embed .default-config.toml
	// ConfigTemplateToml 默认的配置文件模板
	ConfigTemplateToml string
)
