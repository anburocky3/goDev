//go:build windows

package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"image/png"
	"sync"

	systray "github.com/energye/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed frontend/src/assets/images/logo-universal.png
var trayIconPNG []byte

var trayOnce sync.Once

func (a *App) startTray() {
	if a == nil {
		return
	}
	trayOnce.Do(func() {
		systray.Register(func() {
			if trayIcon, err := pngToICO(trayIconPNG); err == nil {
				systray.SetIcon(trayIcon)
			}
			systray.SetTooltip("goDev App")

			systray.SetOnClick(func(menu systray.IMenu) {
				runtime.Show(a.ctx)
				runtime.WindowUnminimise(a.ctx)
				runtime.WindowShow(a.ctx)
			})
			
			// 3. Define the menu (triggers automatically on native Right-Click)
			openItem := systray.AddMenuItem("Open goDev", "Show the main window")
			systray.AddSeparator()
			quitItem := systray.AddMenuItem("Quit", "Quit goDev and stop services")

			openItem.Click(func() {
				runtime.Show(a.ctx)
				runtime.WindowUnminimise(a.ctx)
				runtime.WindowShow(a.ctx)
			})

			quitItem.Click(func() {
				a.requestQuit()
				systray.Quit()
			})
		}, func() {})
	})
}

func (a *App) stopTray() {
	systray.Quit()
}

func pngToICO(pngData []byte) ([]byte, error) {
	config, err := png.DecodeConfig(bytes.NewReader(pngData))
	if err != nil {
		return nil, err
	}

	iconDir := make([]byte, 6+16)
	binary.LittleEndian.PutUint16(iconDir[0:2], 0)
	binary.LittleEndian.PutUint16(iconDir[2:4], 1)
	binary.LittleEndian.PutUint16(iconDir[4:6], 1)
	if config.Width >= 256 {
		iconDir[6] = 0
	} else {
		iconDir[6] = byte(config.Width)
	}
	if config.Height >= 256 {
		iconDir[7] = 0
	} else {
		iconDir[7] = byte(config.Height)
	}
	iconDir[8] = 0
	iconDir[9] = 0
	binary.LittleEndian.PutUint16(iconDir[10:12], 1)
	binary.LittleEndian.PutUint16(iconDir[12:14], 32)
	binary.LittleEndian.PutUint32(iconDir[14:18], uint32(len(pngData)))
	binary.LittleEndian.PutUint32(iconDir[18:22], uint32(len(iconDir)))

	icon := make([]byte, 0, len(iconDir)+len(pngData))
	icon = append(icon, iconDir...)
	icon = append(icon, pngData...)
	return icon, nil
}
