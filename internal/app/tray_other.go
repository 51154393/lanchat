//go:build !windows

package app

import (
	"log"
)

func (a *App) initTray() {
	// 非 Windows 平台不启用系统托盘（避免与 Wails 冲突）
	log.Printf("[Tray] System tray not supported on this platform (would conflict with Wails)")
}