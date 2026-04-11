package collect

import (
	"path/filepath"

	"github.com/fltLi/webgal-snapshot/pkg/parse"
)

// 收集配置文件及关联资源.
func CollectConfig(root string, archiver chan<- Resource) error {
	path := filepath.Join(root, "config.txt")
	archiver <- Resource{Path: path}

	// 解析配置
	config, err := parse.ParseConfig(path)
	if err != nil {
		return err
	}

	// 检出关联资源
	for name, value := range config {
		switch name {
		case "Title_img":
			archiver <- Resource{Path: filepath.Join(root, catBackground, value)}
		case "Title_bgm":
			archiver <- Resource{Path: filepath.Join(root, catBgm, value)}
		}
	}

	return nil
}
