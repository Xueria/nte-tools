package pool

import (
	"encoding/json"
	"fmt"
	"os"
)

// Resource 一种支付资源：标识与名称。
type Resource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LoadResources 读取一个资源文件。路径为空或文件不存在时返回 (nil, nil)，
// 便于调用方回退到其它资源来源。
func LoadResources(file string) ([]Resource, error) {
	if file == "" {
		return nil, nil
	}

	content, err := os.ReadFile(file)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read resource file %s failed: %w", file, err)
	}

	var resources []Resource

	if err := json.Unmarshal(content, &resources); err != nil {
		return nil, fmt.Errorf("unmarshal resource file %s failed: %w", file, err)
	}

	return resources, nil
}
