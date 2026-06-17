package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	services      *ServiceManager
	quitRequested bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	manager, err := NewServiceManager()
	if err != nil {
		manager = &ServiceManager{loadErr: err}
	}
	a.services = manager
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitRequested {
		return false
	}
	a.startTray()
	runtime.WindowHide(ctx)
	return true
}

func (a *App) shutdown(ctx context.Context) {
	if a.services == nil {
		return
	}
	a.stopTray()
	_, _ = a.services.StopAllServices()
}

func (a *App) StartAllServices() (ServiceStatus, error) {
	return a.services.StartAllServices()
}

func (a *App) StopAllServices() (ServiceStatus, error) {
	return a.services.StopAllServices()
}

func (a *App) MinimizeToTray() {
	a.startTray()
	runtime.WindowHide(a.ctx)
}

func (a *App) requestQuit() {
	a.quitRequested = true
	runtime.Quit(a.ctx)
}
