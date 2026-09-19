package pool

import (
	"errors"
	"io/fs"

	"blind-tools/internal/data"
)

// Resource 一种支付资源：标识与名称。
type Resource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// loadResources 读取一个资源文件。路径为空或文件不存在时返回 (nil, nil)，
// 便于调用方回退到其它资源来源。
func loadResources(file string) ([]Resource, error) {
	if file == "" {
		return nil, nil
	}

	var resources []Resource

	if err := data.ReadJSON(file, &resources); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	return resources, nil
}
