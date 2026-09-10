package blindbox

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
)

// DefaultRemoteBaseURL 远程盲盒数据的 GitHub 目录。它必须指向一个
// raw.githubusercontent.com 路径，其结构与本地 DataDirectory 一致：
//
//	<folder>/
//	  currency.json          (可选的段落级全局货币)
//	  <blind-box>/
//	    manifest.json        (必需)
//	    currency.json        (可选，缺失时回退到段落级全局货币)
//
// 置为空字符串可关闭远程数据。
const DefaultRemoteBaseURL = "https://raw.githubusercontent.com/Xueria/blind-tools/refs/heads/master/data"

const (
	githubAPIBase = "https://api.github.com"
	githubRawBase = "https://raw.githubusercontent.com"
)

var errRemoteNotFound = errors.New("remote not found")

// remoteHTTP is the HTTP client used for all GitHub requests.
var remoteHTTP = &http.Client{Timeout: 15 * time.Second}

// LoadRemoteBoxes 从给定的 raw GitHub 目录 URL 加载盲盒。
// URL 由调用方显式传入，使本函数与具体远程源解耦，可单独测试。
func LoadRemoteBoxes(rawURL string) ([]BlindBox, error) {
	if rawURL == "" {
		return nil, nil
	}

	owner, repo, branch, path, err := parseRawURL(rawURL)
	if err != nil {
		return nil, err
	}

	// Section-global currency (mirrors the local data/currency.json).
	globalCurrency, err := fetchCurrency(rawFileURL(owner, repo, branch, path, CurrencyFile))
	if err != nil {
		log.Printf("remote: load global currency failed: %v", err)
	}

	entries, err := listRemoteFolder(owner, repo, branch, path)
	if err != nil {
		return nil, err
	}

	var boxes []BlindBox
	for _, entry := range entries {
		if entry.Type != "dir" {
			continue
		}

		manifest, err := fetchManifest(owner, repo, branch, entry.Path)
		if err != nil {
			if errors.Is(err, errRemoteNotFound) {
				continue // folder without manifest.json
			}
			log.Printf("remote: skip %s: %v", entry.Name, err)
			continue
		}

		localCurrency, err := fetchCurrency(rawFileURL(owner, repo, branch, entry.Path, CurrencyFile))
		if err != nil {
			log.Printf("remote: blind box %s: load local currency failed: %v", entry.Name, err)
		}

		box := BlindBox{Manifest: manifest}
		if localCurrency != nil {
			box.Currencies = localCurrency
		} else {
			box.Currencies = globalCurrency
		}
		if box.Currencies == nil {
			log.Printf("remote: skip %s: no currency info", entry.Name)
			continue
		}

		if err := ValidateManifestPrices(box); err != nil {
			log.Printf("remote: skip %s: %v", entry.Name, err)
			continue
		}
		if err := ValidateManifestDraws(box); err != nil {
			log.Printf("remote: skip %s: %v", entry.Name, err)
			continue
		}

		boxes = append(boxes, box)
	}

	return boxes, nil
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

// ghEntry is a single item of the GitHub contents API response.
type ghEntry struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// listRemoteFolder lists a folder via the GitHub contents API (raw URLs cannot
// list directories).
func listRemoteFolder(owner, repo, branch, path string) ([]ghEntry, error) {
	api := githubAPIBase + "/repos/" + owner + "/" + repo + "/contents"
	if p := strings.Trim(path, "/"); p != "" {
		api += "/" + p
	}
	if branch != "" {
		api += "?ref=" + url.QueryEscape(branch)
	}

	req, err := http.NewRequest(http.MethodGet, api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "blind-tools")

	resp, err := remoteHTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求远程目录失败：%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("远程目录不存在：%s", path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("远程目录请求返回 %s", resp.Status)
	}

	var entries []ghEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("远程路径不是文件夹：%s", path)
	}
	return entries, nil
}

// fetchManifest downloads and parses a manifest.json file.
func fetchManifest(owner, repo, branch, folder string) (Manifest, error) {
	var manifest Manifest
	err := fetchJSON(rawFileURL(owner, repo, branch, folder, ManifestFile), &manifest)
	return manifest, err
}

// fetchCurrency downloads and parses a currency.json file. A missing file
// returns (nil, nil) so callers can fall back to the section-global currency.
func fetchCurrency(rawURL string) ([]Currency, error) {
	var currencies []Currency
	if err := fetchJSON(rawURL, &currencies); err != nil {
		if errors.Is(err, errRemoteNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return currencies, nil
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
