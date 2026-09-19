package pool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"blind-tools/model/metadata"
)

const githubRawBase = "https://raw.githubusercontent.com"

var errRemoteNotFound = errors.New("remote not found")

// remoteHTTP is the HTTP client used for all GitHub requests.
var remoteHTTP = &http.Client{Timeout: 15 * time.Second}

// LoadRemotePools 从给定的 raw GitHub 数据根 URL 加载盲盒池。
// URL 由调用方显式传入，使本函数与具体远程源解耦，可单独测试；
// 该目录的结构必须与本地数据根一致，入口同样是 metadata.json。
func LoadRemotePools(rawURL string) ([]Pool, error) {
	if rawURL == "" {
		return nil, nil
	}

	owner, repo, branch, path, err := parseRawURL(rawURL)

	if err != nil {
		return nil, err
	}

	meta, err := fetchMetadata(owner, repo, branch, path)

	if err != nil {
		return nil, err
	}

	var global []Resource

	if file := meta.Pool.Global.Resources; file != "" {
		global, err = fetchResources(rawFileURL(owner, repo, branch, path, file))

		if err != nil {
			log.Printf("remote: load global resources %s failed: %v", file, err)
		}
	}

	pools := make([]Pool, 0, len(meta.Pool.List))

	for _, relative := range meta.Pool.List {
		var item Pool

		if err := fetchJSON(rawFileURL(owner, repo, branch, path, relative), &item); err != nil {
			log.Printf("remote: skip pool %s: %v", relative, err)
			continue
		}

		applyGlobalResources(&item, global)

		if len(item.Resources) == 0 {
			log.Printf("remote: skip pool %s: no resources", relative)
			continue
		}

		if err := ValidateManifestPrices(item); err != nil {
			log.Printf("remote: skip pool %s: %v", relative, err)
			continue
		}

		if err := ValidateManifestDraws(item); err != nil {
			log.Printf("remote: skip pool %s: %v", relative, err)
			continue
		}

		pools = append(pools, item)
	}

	return pools, nil
}

// parseRawURL splits a raw.githubusercontent.com link into owner, repo, branch
// and folder path. It accepts these ref forms:
//
//	https://raw.githubusercontent.com/owner/repo/master/folder
//	https://raw.githubusercontent.com/owner/repo/HEAD/folder
//	https://raw.githubusercontent.com/owner/repo/refs/heads/master/folder
//	https://raw.githubusercontent.com/owner/repo/refs/tags/v1.0/folder
func parseRawURL(rawURL string) (owner, repo, branch, path string, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", "", "", fmt.Errorf("远程链接无效：%v", err)
	}
	if u.Host != "raw.githubusercontent.com" {
		return "", "", "", "", fmt.Errorf("远程链接需为 raw.githubusercontent.com 形式")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 3 {
		return "", "", "", "", fmt.Errorf("远程链接缺少 owner/repo/ref")
	}
	owner, repo = parts[0], parts[1]
	rest := parts[2:]

	// Normalise "refs/heads/<branch>" / "refs/tags/<tag>" into the short ref,
	// which is what raw.githubusercontent.com URLs expect in their path.
	switch {
	case len(rest) >= 2 && rest[0] == "refs" && (rest[1] == "heads" || rest[1] == "tags") && len(rest) >= 3:
		branch = rest[2]
		path = strings.Join(rest[3:], "/")
	default:
		branch = rest[0]
		path = strings.Join(rest[1:], "/")
	}

	return owner, repo, branch, path, nil
}

// rawFileURL builds a raw.githubusercontent.com URL for a file in a folder.
// file may itself be a path relative to folder, as used by the data index.
func rawFileURL(owner, repo, branch, folder, file string) string {
	ref := branch
	if ref == "" {
		ref = "HEAD"
	}

	parts := []string{githubRawBase, owner, repo, ref}
	if folder = strings.Trim(folder, "/"); folder != "" {
		parts = append(parts, folder)
	}
	parts = append(parts, file)
	return strings.Join(parts, "/")
}

// fetchMetadata downloads and parses the data root index file.
func fetchMetadata(owner, repo, branch, folder string) (metadata.Metadata, error) {
	var meta metadata.Metadata

	err := fetchJSON(rawFileURL(owner, repo, branch, folder, metadata.File), &meta)

	if err != nil {
		return metadata.Metadata{}, fmt.Errorf("读取远程数据索引 %s 失败：%v", metadata.File, err)
	}

	return meta, nil
}

// fetchResources downloads and parses a resource file. A missing file returns
// (nil, nil) so callers can fall back to the global resources.
func fetchResources(rawURL string) ([]Resource, error) {
	var resources []Resource

	if err := fetchJSON(rawURL, &resources); err != nil {
		if errors.Is(err, errRemoteNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return resources, nil
}

// fetchJSON downloads rawURL and decodes it into out.
func fetchJSON(rawURL string, out any) error {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "blind-tools")

	resp, err := remoteHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("下载 %s 失败：%v", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errRemoteNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载 %s 返回 %s", rawURL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("解析 %s 失败：%v", rawURL, err)
	}
	return nil
}
