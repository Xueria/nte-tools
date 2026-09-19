package data

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"
)

// remoteTimeout 单次远程请求的超时：数据文件都不大，超时短一些，远程不通时
// 窗口才不至于等太久。
const remoteTimeout = 5 * time.Second

// Remote 返回通过 HTTP 读取 baseURL 下数据文件的来源，请求路径按索引里的
// 相对路径拼接。
func Remote(baseURL string) Source {
	return remoteSource{
		baseURL: strings.TrimSuffix(baseURL, "/") + "/",
		client:  &http.Client{Timeout: remoteTimeout},
	}
}

// remoteSource 从 HTTP 服务读文件。
type remoteSource struct {
	baseURL string
	client  *http.Client
}

// ReadFile 下载数据文件。文件不存在（404）按 fs.ErrNotExist 返回，让上层能像读
// 本地文件一样判断；其余非 200 状态与网络错误都算读取失败。
func (s remoteSource) ReadFile(name string) ([]byte, error) {
	cleaned, err := cleanName(name)

	if err != nil {
		return nil, err
	}

	target := s.baseURL + cleaned

	response, err := s.client.Get(target)

	if err != nil {
		return nil, fmt.Errorf("get %s failed: %w", target, err)
	}

	defer response.Body.Close()

	switch {
	case response.StatusCode == http.StatusNotFound:
		return nil, fmt.Errorf("get %s: %w", target, fs.ErrNotExist)
	case response.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("get %s: unexpected status %s", target, response.Status)
	}

	content, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("read %s failed: %w", target, err)
	}

	return content, nil
}
