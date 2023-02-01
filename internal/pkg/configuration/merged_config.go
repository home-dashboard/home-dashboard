package configuration

import (
	"github.com/jinzhu/copier"
	"log"
)

var Config *Configuration = merge()

// 合并通过配置文件设置的配置和命令行传入的配置.
// 如配置冲突则以命令行配置为准.
func merge() *Configuration {
	merged := Configuration{}
	copyOption := copier.Option{IgnoreEmpty: true}
	if err := copier.CopyWithOption(&merged, &fileConfig, copyOption); err != nil {
		log.Fatalf("file config merge failed, %s\n", err)
	}
	if err := copier.CopyWithOption(&merged, &argumentsConfig, copyOption); err != nil {
		log.Fatalf("arguments config merge failed, %s\n", err)
	}

	log.Printf("merged configuration: %s\n", merged)

	return &merged
}
