package bid

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// GridFileSuffix 占格数据文件名的后缀，完整形如 <length>x<width>.json。
const GridFileSuffix = ".json"

// LoadGrids 读取 directory 下所有 <length>x<width>.json，按文件名顺序返回。
// 文件名不合约定、读取失败或内容不合法的文件会被跳过并记录日志，
// 单个坏文件不影响其余数据。
func LoadGrids(directory string) ([]Grid, error) {
	entries, err := os.ReadDir(directory)

	if err != nil {
		return nil, fmt.Errorf("error read directory %s: %w", directory, err)
	}

	grids := make([]Grid, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != GridFileSuffix {
			continue
		}

		path := filepath.Join(directory, entry.Name())

		grid, err := LoadGrid(path)

		if err != nil {
			log.Printf("skip bid grid %s: %v", path, err)
			continue
		}

		grids = append(grids, grid)
	}

	return grids, nil
}

// LoadGrid 读取单个占格数据文件，并用文件名的 <length>x<width> 校验顶层长宽。
func LoadGrid(file string) (Grid, error) {
	name := filepath.Base(file)

	length, width, ok := ParseGridName(name)

	if !ok {
		return Grid{}, fmt.Errorf("文件名 %s 不是 <length>x<width>.json 形式", name)
	}

	content, err := os.ReadFile(file)

	if err != nil {
		return Grid{}, fmt.Errorf("read %s failed: %w", file, err)
	}

	var grid Grid

	if err := json.Unmarshal(content, &grid); err != nil {
		return Grid{}, fmt.Errorf("unmarshal %s failed: %w", file, err)
	}

	if err := ValidateGrid(grid, length, width); err != nil {
		return Grid{}, fmt.Errorf("validate %s failed: %w", file, err)
	}

	return grid, nil
}
