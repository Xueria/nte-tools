package data

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadJSON 读取并解析 path 指向的 JSON 文件。
func ReadJSON(path string, target any) error {
	content, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("read %s failed: %w", path, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("unmarshal %s failed: %w", path, err)
	}

	return nil
}
