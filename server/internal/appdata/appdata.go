// Package appdata 固定程序的数据目录。go-music-dl 的 core 包使用相对路径 data/...
// 保存配置和下载文件，所以启动时切换工作目录到数据目录，避免依赖启动位置。
package appdata

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const AppDirName = "KaiShuiMusic"

// Prepare 创建并进入数据目录。dir 为空时使用系统用户配置目录（Windows: %AppData%\KaiShuiMusic）。
func Prepare(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("获取用户配置目录失败: %w", err)
		}
		dir = filepath.Join(base, AppDirName)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(abs, "data"), 0o755); err != nil {
		return "", fmt.Errorf("创建数据目录失败: %w", err)
	}
	if err := os.Chdir(abs); err != nil {
		return "", err
	}
	return abs, nil
}
