package model

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	// DataDirectory 数据根目录，包含全局货币文件与各盲盒子目录
	DataDirectory = "data"
	// ManifestFile 盲盒数据
	// 包含盲盒价格
	ManifestFile = "manifest.json"
	// CurrencyFile 货币信息
	CurrencyFile = "currency.json"
)

// LoadLocalBoxes 读取 directory 下的所有盲盒子目录。
// 没有自带 currency.json 的盲盒回退到根目录的全局货币文件。
func LoadLocalBoxes(directory string) ([]BlindBox, error) {
	// 加载全局货币信息
	globalCurrency, err := LoadCurrency(filepath.Join(directory, CurrencyFile))

	if err != nil {
		log.Printf("load global currency %s failed: %v", directory, err)
	}

	// 读取目录下所有文件
	files, err := os.ReadDir(directory)

	if err != nil {
		return nil, fmt.Errorf("error read directory %s: %w", directory, err)
	}

	var boxes []BlindBox

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		boxDirectory := filepath.Join(directory, file.Name())
		manifestFile := filepath.Join(boxDirectory, ManifestFile)

		content, err := os.ReadFile(manifestFile)

		if err != nil {
			// 没有 manifest 的文件夹跳过
			if os.IsNotExist(err) {
				continue
			}
			log.Printf("skip blind box %s: read manifest failed: %v", boxDirectory, err)
			continue
		}

		var manifest Manifest

		if err := json.Unmarshal(content, &manifest); err != nil {
			log.Printf("skip blind box %s: unmarshal manifest failed: %v", boxDirectory, err)
			continue
		}

		// 加载本地货币信息
		localCurrency, err := LoadCurrency(filepath.Join(boxDirectory, CurrencyFile))

		if err != nil {
			log.Printf("blind box %s: load local currency failed: %v", boxDirectory, err)
		}

		if globalCurrency == nil && localCurrency == nil {
			log.Printf("skip blind box %s: global currency and local currency are nil", boxDirectory)
			continue
		}

		box := BlindBox{Manifest: manifest}

		if localCurrency != nil {
			box.Currencies = localCurrency
		} else {
			box.Currencies = globalCurrency
		}

		if err := ValidateManifestPrices(box); err != nil {
			log.Printf("skip blind box: %v", err)
			continue
		}

		if err := ValidateManifestDraws(box); err != nil {
			log.Printf("skip blind box: %v", err)
			continue
		}

		boxes = append(boxes, box)
	}

	return boxes, nil
}

// LoadCurrency 读取一个 currency.json 文件。
// 文件不存在时返回 (nil, nil)，便于调用方回退到其它货币来源。
func LoadCurrency(file string) ([]Currency, error) {
	text, err := os.ReadFile(file)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error read currency file %s : %w", file, err)
	}

	var currencies []Currency

	if err := json.Unmarshal(text, &currencies); err != nil {
		return nil, fmt.Errorf("error unmarshal currency file %s : %w", file, err)
	}

	return currencies, nil
}
