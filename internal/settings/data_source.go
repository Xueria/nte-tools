// Package settings 保存运行期设置。目前只有一项：数据从远程还是本地读，
// 以及远程数据根在哪。
package settings

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
