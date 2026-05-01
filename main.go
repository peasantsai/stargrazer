package main

import (
	"embed"
	"log"

	"stargrazer/internal/automation"
	"stargrazer/internal/browser"
	"stargrazer/internal/config"
	"stargrazer/internal/db"
	"stargrazer/internal/db/backfill"
	"stargrazer/internal/scheduler"
	"stargrazer/internal/social"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := config.Init(); err != nil {
		log.Printf("config warning: %v (using defaults)", err)
	}

	dbPath := social.DBPath()
	sqlDB, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open sqlite at %s: %v", dbPath, err)
	}
	defer sqlDB.Close()

	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	if err := backfill.RunIfNeeded(sqlDB, social.SharedSessionDirParent()); err != nil {
		log.Fatalf("backfill: %v", err)
	}

	autoRepo := automation.NewSQLiteRepo(sqlDB)
	sessRepo := social.NewSQLiteSessionRepo(sqlDB)
	schedRepo := scheduler.NewSQLiteRepo(sqlDB)
	browserMgr := browser.GetInstance()
	sched := scheduler.GetInstance(browserMgr, sessRepo, schedRepo)

	win := config.GetWindow()
	app := NewApp(autoRepo, sessRepo, sched, browserMgr)

	err = wails.Run(&options.App{
		Title:     win.Title,
		Width:     win.Width,
		Height:    win.Height,
		MinWidth:  win.MinWidth,
		MinHeight: win.MinHeight,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 15, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}
