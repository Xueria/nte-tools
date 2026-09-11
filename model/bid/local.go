package bid

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ListingFileSuffix 清单数据文件的后缀。
const ListingFileSuffix = ".json"

// LoadListings 读取 directory 下所有 .json，按文件名顺序返回。文件名带不带
// <length>x<width> 都可以，读不出长宽的清单降级为普通拍品。读取失败或内容
// 不合法的文件会被跳过并记录日志，单个坏文件不影响其余数据。
func LoadListings(directory string) ([]Listing, error) {
	entries, err := os.ReadDir(directory)

	if err != nil {
		return nil, fmt.Errorf("error read directory %s: %w", directory, err)
	}

	listings := make([]Listing, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ListingFileSuffix {
			continue
		}

		path := filepath.Join(directory, entry.Name())

		listing, err := LoadListing(path)

		if err != nil {
			log.Printf("skip bid listing %s: %v", path, err)
			continue
		}

		listings = append(listings, listing)
	}

	return listings, nil
}

// LoadListing 读取单个清单文件。文件名里的 <length>x<width> 与数据里的 name 都
// 只是缺省值：长宽以数据为准、缺省时才看文件名，两者都没有即普通拍品清单；
// 清单名以数据为准、缺省时才取文件名。因此文件名不再限制为 <length>x<width>。
func LoadListing(file string) (Listing, error) {
	content, err := os.ReadFile(file)

	if err != nil {
		return Listing{}, fmt.Errorf("read %s failed: %w", file, err)
	}

	var listing Listing

	if err := json.Unmarshal(content, &listing); err != nil {
		return Listing{}, fmt.Errorf("unmarshal %s failed: %w", file, err)
	}

	listing.Name = resolveListingName(listing.Name, file)
	listing.Footprint = resolveFootprint(listing.Footprint, file)

	if err := ValidateListing(listing); err != nil {
		return Listing{}, fmt.Errorf("validate %s failed: %w", file, err)
	}

	return listing, nil
}

// resolveListingName 返回清单名：数据里的 name 优先，缺省时取文件名。
func resolveListingName(name, file string) string {
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		return trimmed
	}

	return strings.TrimSuffix(filepath.Base(file), ListingFileSuffix)
}

// resolveFootprint 返回清单的占格形状：数据里的长宽优先，缺省时用文件名里的
// <length>x<width>，两者都没有则降级为普通拍品。
func resolveFootprint(footprint Footprint, file string) Footprint {
	if footprint.Length >= 1 && footprint.Width >= 1 {
		return NewFootprint(footprint.Length, footprint.Width)
	}

	length, width, ok := ParseSizeFromName(filepath.Base(file))

	if !ok {
		return PlainFootprint()
	}

	return NewFootprint(length, width)
}
