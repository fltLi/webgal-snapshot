package figure

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Live2D 捆版模型.
type Bundle struct {
	SubModels []BundleSubModel `json:"subModels"`
}

// Live2D 捆绑子模型.
type BundleSubModel struct {
	Model string `json:"modelRelativaPath"`
}

// 解析并列出 Live2D 捆绑模型相关资源.
// 参数:
//   - path: 模型配置文件路径
func BundleAssets(path string) ([]string, error) {
	// 读取配置
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var model Bundle
	if err = json.Unmarshal(data, &model); err != nil {
		return nil, err
	}

	// 列出资源
	var errs []error
	dir := filepath.Dir(path)
	assets := []string{path}
	for _, sub := range model.SubModels {
		subModel := filepath.Join(dir, sub.Model)
		subAssets, err := Live2dAssets(subModel)
		if err != nil {
			errs = append(errs, err)
		} else {
			assets = append(assets, subAssets...)
		}
	}

	return assets, errors.Join(errs...)
}
