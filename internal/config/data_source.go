// Package config 保存应用级设置。目前只有一项：数据从哪个来源读。
package config

import (
	"strings"

	"blind-tools/internal/data"
)

// 内置的远程数据根。
const (
	// CloudflareBaseURL Cloudflare Worker 上的数据根。
	CloudflareBaseURL = "https://nte-data.xueria.workers.dev/nte/data/"
	// GitHubBaseURL 数据仓库 raw 地址上的数据根。
	GitHubBaseURL = "https://raw.githubusercontent.com/Xueria/nte-data/master/public/nte/data/"
)

// Kind 数据来源的种类。
type Kind string

const (
	// CloudflareKind 从 Cloudflare Worker 读。
	CloudflareKind Kind = "cloudflare"
	// GitHubKind 从 GitHub 仓库的 raw 地址读。
	GitHubKind Kind = "github"
	// CustomKind 从用户自己填的地址读。
	CustomKind Kind = "custom"
	// LocalKind 从本地数据目录读。
	LocalKind Kind = "local"
)

// Option 设置页上的一个选项：种类与展示名。
type Option struct {
	Kind  Kind
	Label string
}

// options 是全部可选项，按设置页上的顺序排列。
var options = []Option{
	{CloudflareKind, "远程（Cloudflare）"},
	{GitHubKind, "远程（GitHub）"},
	{CustomKind, "自定义远程"},
	{LocalKind, "本地"},
}

// Options 返回可选的数据来源。
func Options() []Option {
	return append([]Option(nil), options...)
}

// KindOfLabel 返回展示名对应的种类；不是已知选项时 ok 为 false。
func KindOfLabel(label string) (Kind, bool) {
	for _, option := range options {
		if option.Label == label {
			return option.Kind, true
		}
	}

	return "", false
}

// Choice 一次数据来源的选择：种类，以及自定义远程时的地址。
type Choice struct {
	Kind      Kind
	CustomURL string
}

// Default 返回默认的数据来源。
func Default() Choice {
	return Choice{Kind: CloudflareKind}
}

// Label 返回该选择在设置页上的展示名。
func (c Choice) Label() string {
	for _, option := range options {
		if option.Kind == c.Kind {
			return option.Label
		}
	}

	return Default().Label()
}

// Open 按选择打开数据来源。
func (c Choice) Open() data.Source {
	if c.Kind == LocalKind {
		return data.Local(data.Root())
	}

	return data.Remote(c.baseURL())
}

// Hint 说明该选择是从哪儿读数据的。
func (c Choice) Hint() string {
	if c.Kind == LocalKind {
		return "从程序目录下的 data/ 读索引与数据，目录可用 " + data.RootEnv + " 环境变量指定。"
	}

	if baseURL := c.baseURL(); baseURL != "" {
		return "从 " + baseURL + " 读索引与数据。"
	}

	return "填一个远程数据根地址（以 / 结尾），索引与索引里列出的文件都从它读取。"
}

// baseURL 返回该选择对应的远程根；本地或没填自定义地址时为空。
func (c Choice) baseURL() string {
	switch c.Kind {
	case CloudflareKind:
		return CloudflareBaseURL
	case GitHubKind:
		return GitHubBaseURL
	case CustomKind:
		return strings.TrimSpace(c.CustomURL)
	default:
		return ""
	}
}
