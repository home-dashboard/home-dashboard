package configuration

import (
	"github.com/go-errors/errors"
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
)

var logger = comfy_log.New("[configuration]")

var config *Configuration

// 合并通过配置文件设置的配置和命令行传入的配置.
// 如配置冲突则以命令行配置为准.
func merge() *Configuration {
	argumentsConfig := parseArguments()
	fileConfig := parseFile()

	merged := Configuration{}

	copyOption := copier.Option{IgnoreEmpty: true, DeepCopy: true}
	if err := copier.CopyWithOption(&merged, &fileConfig, copyOption); err != nil {
		logger.Error("file config merge failed, %w\n", errors.New(err))
	}
	if err := copier.CopyWithOption(&merged, &argumentsConfig, copyOption); err != nil {
		logger.Error("arguments config merge failed, %w\n", errors.New(err))
	}

	return &merged
}

func Get() *Configuration {
	if config == nil {
		config = merge()
	}

	return config
}
