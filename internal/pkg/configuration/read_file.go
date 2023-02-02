package configuration

import (
	_ "embed"
	"github.com/BurntSushi/toml"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/configs/config_template"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"log"
	"os"
	"path"
)

var (
	defaultConfigFilename = ".config.toml"
)

var fileConfig *Configuration = generateOrReadConfigFile()

// 读取当前目录下名为 config.toml 或 .config.toml 的配置文件.
// 如果在当前目录下未找到所需的配置文件则创建一个内容为 defaultConfig , 文件名为 defaultConfigFilename 的配置文件.
func generateOrReadConfigFile() *Configuration {
	filePath := path.Join(utils.WorkspaceDir, defaultConfigFilename)

	if !utils.FileExist(filePath) {
		file, err := os.Create(filePath)
		if err != nil {
			log.Fatalf("config file create failed, %s\n", err)
		}

		_, err = file.WriteString(config_template.ConfigTemplateToml)
		if err != nil {
			log.Fatalf("config file write failed, %s\n", err)
		}

		err = file.Close()
		if err != nil {
			log.Fatalf("config file close failed, %s\n", err)
		}

		log.Printf("default config file is created in %s\n", filePath)
	}

	// 从配置文件中读取的配置信息
	externalConfig := &Configuration{}
	_, err := toml.DecodeFile(filePath, externalConfig)
	if err != nil {
		log.Fatalf("config file decode failed, %s\n", err)
	}

	defaultConfig := &Configuration{}
	_, err = toml.Decode(config_template.ConfigTemplateToml, defaultConfig)
	if err != nil {
		log.Fatalf("default config decode failed, %s\n", err)
	}

	err = copier.CopyWithOption(defaultConfig, externalConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		log.Fatalf("merge config failed, %s\n", err)
	}

	return defaultConfig
}
