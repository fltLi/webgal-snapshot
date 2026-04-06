package figure

import (
	"fmt"
	"path/filepath"
	"strings"
)

type FigureType = int

// 模型类型.
const (
	figImage = iota
	figLive2d
	figBundle
	figSpine
)

// 识别模型类型.
// 参数:
//   - path: 模型配置文件路径 (可能含有类型标注)
//
// 返回:
//   - FigureType: 模型类型
//   - string: 模型实际路径
func Type(path string) (FigureType, string) {
	if realPath, ok := strings.CutSuffix(path, "?type=spine"); ok {
		return figSpine, realPath
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return figLive2d, path
	case ".wmdl":
		return figBundle, path
	case ".skel":
		return figSpine, path
	default:
		return figImage, path
	}
}

// 识别模型类型并列出相关资源.
// 参数:
//   - path: 模型配置文件路径 (可能含有类型标注)
func Assets(path string) ([]string, error) {
	t, path := Type(path)
	switch t {
	case figImage:
		return []string{path}, nil
	case figLive2d:
		return Live2dAssets(path)
	case figBundle:
		return BundleAssets(path)
	case figSpine:
		return []string{path}, fmt.Errorf("暂不支持 Spine 模型: %s", path)
	default:
		panic(fmt.Sprintf("未知模型枚举: %d", t)) // unreachable
	}
}
