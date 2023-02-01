package configuration

import (
	_ "embed"
	"github.com/BurntSushi/toml"
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
	config := Configuration{}

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
	}

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("config file read failed, %s\n", err)
	}

	_, err = toml.Decode(string(bytes), &config)
	if err != nil {
		log.Fatalf("config file decode failed, %s\n", err)
	}

	return &config
}
