package data

import (
	"os"
	"path/filepath"
)

const (
	// RootEnv 环境变量：显式指定数据根目录，优先级最高。
	RootEnv = "BLIND_TOOLS_DATA"
	// LocalDirectory 数据根目录名，相对可执行文件或当前工作目录。
	LocalDirectory = "data"
)

// Root 按默认策略定位本地数据根目录：依次取环境变量、可执行文件同级的 data/、
// 当前工作目录下的 data/，保证从别处启动 exe 时仍能找到随包分发的数据。
func Root() string {
	if root := os.Getenv(RootEnv); root != "" {
		return root
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), LocalDirectory)

		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return LocalDirectory
}
