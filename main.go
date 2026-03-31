package main

import (
	"embed"
	"os"
	"path/filepath"
	"runtime/debug"

	"investment-platform/internal/shared"

	"go-stock/backend/data"
	log "go-stock/backend/logger"
	"go-stock/backend/runtimepaths"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.SugaredLogger.Error("panic: ", r)
			log.SugaredLogger.Error("stack: ", string(debug.Stack()))
		}
	}()

	checkDir(runtimepaths.DataDir())
	checkDir(runtimepaths.QuantTemplatesDir())
	checkDir(runtimepaths.LogsDir())

	shared.InitDB(runtimepaths.DBPath())
	data.InitAnalyzeSentiment()

	log.SugaredLogger.Info("rubin investment starting")

	app := NewApp()
	appMenu := menu.NewMenu()
	backgroundColour := &options.RGBA{R: 24, G: 28, B: 36, A: 1}

	err := wails.Run(&options.App{
		Title:                    "Rubin Investment",
		Width:                    1456 * 4 / 5,
		Height:                   920,
		MinWidth:                 1200,
		MinHeight:                768,
		DisableResize:            false,
		Fullscreen:               false,
		Frameless:                false,
		StartHidden:              false,
		HideWindowOnClose:        false,
		EnableDefaultContextMenu: true,
		BackgroundColour:         backgroundColour,
		Assets:                   assets,
		Menu:                     appMenu,
		Logger:                   logger.NewFileLogger(filepath.Join(runtimepaths.LogsDir(), "platform.log")),
		LogLevel:                 logger.DEBUG,
		LogLevelProduction:       logger.INFO,
		OnStartup:                app.startup,
		OnDomReady:               app.domReady,
		OnBeforeClose:            app.beforeClose,
		OnShutdown:               app.shutdown,
		WindowStartState:         options.Maximised,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "rubin-investment",
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			WebviewUserDataPath:  "",
		},
	})

	if err != nil {
		log.SugaredLogger.Fatal(err)
	}
}

func checkDir(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
		log.SugaredLogger.Info("create dir: " + dir)
	}
}
