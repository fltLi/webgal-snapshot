package figure

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Type 表示模型类型.
type Type = int

// 模型类型.
const (
	figImage = iota
	figLive2d
	figBundle
	figSpine
)

// GetType 识别模型类型.
// 参数:
//   - path: 模型配置文件路径 (可能含有类型标注)
//
// 返回:
//   - FigureType: 模型类型
//   - string: 模型实际路径
func GetType(path string) (Type, string) {
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

// GetAssets 识别模型类型并列出相关资源.
// 参数:
//   - path: 模型配置文件路径 (可能含有类型标注)
func GetAssets(path string) ([]string, error) {
	t, path := GetType(path)
	switch t {
	case figImage:
		return []string{path}, nil
	case figLive2d:
		return GetLive2dAssets(path)
	case figBundle:
		return GetBundleAssets(path)
	case figSpine:
		return []string{path}, fmt.Errorf("暂不支持 Spine 模型: %s", path)
	default:
		panic(fmt.Sprintf("未知模型枚举: %d", t)) // unreachable
	}
}
