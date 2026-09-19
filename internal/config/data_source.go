// Package config 保存应用级设置。目前只有一项：数据从远程还是本地读，
// 以及远程数据根在哪。
package config

import "blind-tools/internal/data"

// RemoteBaseURL 远程数据根：索引与索引里列出的数据文件都从这里取。
const RemoteBaseURL = "https://nte-data.xueria.workers.dev/nte/data/"

// 数据来源的取值。界面上的下拉直接用这两个名字，因此不必再维护一层映射。
const (
	RemoteSource = "远程"
	LocalSource  = "本地"
)

// Sources 返回可选的数据来源，按界面上的顺序排列。
func Sources() []string {
	return []string{RemoteSource, LocalSource}
}

// DefaultSource 返回默认的数据来源。
func DefaultSource() string {
	return RemoteSource
}

// OpenSource 按选择的数据来源打开数据来源；取值不认识时按远程处理。
func OpenSource(source string) data.Source {
	if source == LocalSource {
		return data.Local(data.Root())
	}

	return data.Remote(RemoteBaseURL)
}

// SourceHint 返回某个数据来源是从哪儿读数据的，用于界面上说明这个设置。
func SourceHint(source string) string {
	if source == LocalSource {
		return "从程序目录下的 data/ 读索引与数据，目录可用 " + data.RootEnv + " 环境变量指定。"
	}

	return "从 " + RemoteBaseURL + " 读索引与数据。"
}
