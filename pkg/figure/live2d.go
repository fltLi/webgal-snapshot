package figure

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// WebGAL Live2D 模型配置.
type Live2d struct {
	Model       string                    `json:"model"`
	Physics     string                    `json:"physics"`
	Textures    []string                  `json:"textures"`
	Motions     map[string][]Live2dMotion `json:"motions"`
	Expressions []Live2dMotion            `json:"expressions"`
}

// Live2D 模型动作表情片段.
type Live2dMotion struct {
	File string `json:"file"`
}

// 解析并列出 Live2D 模型相关资源.
// 参数:
//   - path: 模型配置文件路径
func Live2dAssets(path string) ([]string, error) {
	// 读取配置
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var model Live2d
	if err = json.Unmarshal(data, &model); err != nil {
		return nil, err
	}

	// 列出资源
	assets := []string{model.Model, model.Physics}
	assets = append(assets, model.Textures...)
	for _, motions := range model.Motions {
		for _, motion := range motions {
			assets = append(assets, motion.File)
		}
	}
	for _, expression := range model.Expressions {
		assets = append(assets, expression.File)
	}

	// 整理路径
	dir := filepath.Dir(path)
	for i := range assets {
		assets[i] = filepath.Join(dir, assets[i])
	}
	assets = append(assets, path)

	return assets, nil
}
