package configuration

import (
	_ "embed"
	"github.com/BurntSushi/toml"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/configs/config_template"
	"github.com/siaikin/home-dashboard/internal/pkg/utils"
	"os"
	"path"
	"time"
)

var (
	DefaultConfigFilename = ".config.toml"
)

var fileConfig *Configuration

// 读取当前目录下名为 config.toml 或 .config.toml 的配置文件.
// 如果在当前目录下未找到所需的配置文件则创建一个内容为 [config_template.ConfigTemplateToml] , 文件名为 [DefaultConfigFilename] 的配置文件.
func parseFile() *Configuration {
	if fileConfig == nil {
		fileConfig = generateOrReadConfigFile()
	}

	return fileConfig
}

func generateOrReadConfigFile() *Configuration {
	filePath := path.Join(utils.WorkspaceDir(), DefaultConfigFilename)

	var configFile *os.File

	// 创建 or 打开配置文件
	if exist, err := utils.FileExist(filePath); !exist || err != nil {
		file, err := os.Create(filePath)
		if err != nil {
			logger.Fatal("config file create failed, %s\n", err)
		}
		configFile = file

		_, err = file.WriteString(config_template.ConfigTemplateToml)
		if err != nil {
			logger.Fatal("config file write failed, %s\n", err)
		}

		logger.Info("default config file is created in %s\n", filePath)
	} else {
		if file, err := os.Open(filePath); err != nil {
			logger.Fatal("config file open failed, %s\n", err)
		} else {
			configFile = file
		}
	}
	// 获取文件修改时间
	var modificationTime time.Time
	if fileInfo, err := configFile.Stat(); err != nil {
		logger.Fatal("get stat failed, %s\n", err)
	} else {
		modificationTime = fileInfo.ModTime()
	}
	// 关闭文件
	if err := configFile.Close(); err != nil {
		logger.Fatal("config file close failed, %s\n", err)
	}

	// 从配置文件中读取的配置信息
	externalConfig := &Configuration{
		// 考虑到短时间多次启动的情况, 使用高精度的 unix 时间戳.
		ModificationTime: modificationTime.UnixNano(),
	}
	_, err := toml.DecodeFile(filePath, externalConfig)
	if err != nil {
		logger.Fatal("config file decode failed, %s\n", err)
	}

	defaultConfig := &Configuration{}
	_, err = toml.Decode(config_template.ConfigTemplateToml, defaultConfig)
	if err != nil {
		logger.Fatal("default config decode failed, %s\n", err)
	}

	err = copier.CopyWithOption(defaultConfig, externalConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		logger.Fatal("merge config failed, %s\n", err)
	}

	return defaultConfig
}
