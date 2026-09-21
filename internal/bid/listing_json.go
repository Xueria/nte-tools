package bid

import (
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"

	"nte-tools/internal/data"
)

// listingFileSuffix 清单数据文件的后缀。
const listingFileSuffix = ".json"

// LoadListings 按来源里的索引 bid.list 的顺序读取清单。
// 文件名带不带 <length>x<width> 都可以，读不出长宽的清单降级为普通拍品。
// 读取失败或内容不合法的文件会被跳过并记录日志，单个坏文件不影响其余数据。
func LoadListings(source data.Source) ([]Listing, error) {
	index, err := data.LoadIndex(source)

	if err != nil {
		return nil, err
	}

	listings := make([]Listing, 0, len(index.Bid.List))

	for _, relative := range index.Bid.List {
		listing, err := loadListing(source, relative)

		if err != nil {
			log.Printf("skip bid listing %s: %v", relative, err)
			continue
		}

		listings = append(listings, listing)
	}

	return listings, nil
}

// loadListing 读取单个清单文件。文件名里的 <length>x<width> 只是缺省值：
// attribute 以数据为准、数据里没写时才看文件名，两者都没有即普通拍品清单；
// 清单名同理，数据里的 name 优先、缺省时才取文件名。因此文件名不再限制为
// <length>x<width>。
func loadListing(source data.Source, name string) (Listing, error) {
	var listing Listing

	if err := data.ReadJSON(source, name, &listing); err != nil {
		return Listing{}, err
	}

	listing.Name = resolveListingName(listing.Name, name)
	listing.Attribute = resolveAttribute(listing.Attribute, name)

	if err := validateListing(listing); err != nil {
		return Listing{}, fmt.Errorf("validate %s failed: %w", name, err)
	}

	return listing, nil
}

// resolveListingName 返回清单名：数据里的 name 优先，缺省时取文件名。
func resolveListingName(name, file string) string {
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		return trimmed
	}

	return strings.TrimSuffix(path.Base(file), listingFileSuffix)
}

// resolveAttribute 返回清单的属性：数据里写了 attribute 就以它为准，否则用
// 文件名里的 <length>x<width>，两者都没有则降级为普通拍品。
func resolveAttribute(attribute Attribute, file string) Attribute {
	if !attribute.IsZero() {
		return attribute.Normalize()
	}

	length, width, ok := parseSizeFromName(path.Base(file))

	if !ok {
		return PlainAttribute()
	}

	return GridAttribute(length, width)
}

// parseSizeFromName 从清单文件名解析占格长宽（<length>x<width>.json）。
// ok 为 false 表示文件名没有带长宽，该清单按普通拍品处理。
func parseSizeFromName(name string) (length, width int, ok bool) {
	lengthText, widthText, found := strings.Cut(strings.TrimSuffix(name, listingFileSuffix), "x")

	if !found {
		return 0, 0, false
	}

	length, err := strconv.Atoi(lengthText)

	if err != nil || length < 1 {
		return 0, 0, false
	}

	width, err = strconv.Atoi(widthText)

	if err != nil || width < 1 {
		return 0, 0, false
	}

	return length, width, true
}
