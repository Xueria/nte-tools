package bid

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"blind-tools/model/metadata"
)

// ListingFileSuffix 清单数据文件的后缀。
const ListingFileSuffix = ".json"

// LoadListings 按 directory 下的 metadata.json 里 bid.list 的顺序读取清单。
// 文件名带不带 <length>x<width> 都可以，读不出长宽的清单降级为普通拍品。
// 读取失败或内容不合法的文件会被跳过并记录日志，单个坏文件不影响其余数据。
func LoadListings(directory string) ([]Listing, error) {
	meta, err := metadata.Load(directory)

	if err != nil {
		return nil, err
	}

	listings := make([]Listing, 0, len(meta.Bid.List))

	for _, relative := range meta.Bid.List {
		path := filepath.Join(directory, relative)

		listing, err := LoadListing(path)

		if err != nil {
			log.Printf("skip bid listing %s: %v", path, err)
			continue
		}

		listings = append(listings, listing)
	}

	return listings, nil
}

// LoadListing 读取单个清单文件。文件名里的 <length>x<width> 只是缺省值：
// attribute 以数据为准、数据里没写时才看文件名，两者都没有即普通拍品清单；
// 清单名同理，数据里的 name 优先、缺省时才取文件名。因此文件名不再限制为
// <length>x<width>。
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
	listing.Attribute = resolveAttribute(listing.Attribute, file)

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

// resolveAttribute 返回清单的属性：数据里写了 attribute 就以它为准，否则用
// 文件名里的 <length>x<width>，两者都没有则降级为普通拍品。
func resolveAttribute(attribute Attribute, file string) Attribute {
	if !attribute.IsZero() {
		return attribute.Normalize()
	}

	length, width, ok := ParseSizeFromName(filepath.Base(file))

	if !ok {
		return PlainAttribute()
	}

	return GridAttribute(length, width)
}
