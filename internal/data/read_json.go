package data

import (
	"encoding/json"
	"fmt"
)

// ReadJSON 从来源读出 name 指向的 JSON 文件并解析。
func ReadJSON(source Source, name string, target any) error {
	content, err := source.ReadFile(name)

	if err != nil {
		return fmt.Errorf("read %s failed: %w", name, err)
	}

	if err := json.Unmarshal(content, target); err != nil {
		return fmt.Errorf("unmarshal %s failed: %w", name, err)
	}

	return nil
}
