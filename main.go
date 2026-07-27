package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/arxdsilva/vault/internal/platform"
	"github.com/arxdsilva/vault/internal/service"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/fileserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "Vault",
		Description: "A minimal, cross-platform file manager",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.RegisterService(application.NewService(service.New(platform.NewTrasher(), app.Event)))

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	frecencyPath := filepath.Join(configDir, "vault", "frecency.json")
	app.RegisterService(application.NewService(service.NewSearch(home, frecencyPath)))

	pinnedPath := filepath.Join(configDir, "vault", "pinned.json")
	app.RegisterService(application.NewService(service.NewPinned(pinnedPath, home, platform.ApplicationsDir())))

	app.RegisterService(application.NewService(service.NewVolumes(platform.NewVolumes(), app.Event)))

	permissionPromptPath := filepath.Join(configDir, "vault", "permission-prompt.json")
	app.RegisterService(application.NewService(service.NewPermissions(platform.NewPermissions(), permissionPromptPath)))

	// Streams arbitrary file bytes to the frontend for quick preview (spec
	// §5.4): images/video/audio/PDF load directly via <img>/<video>/<audio>/
	// <iframe> src, text/code is fetched and rendered client-side. RootPath
	// "/" serves any path the user can browse to; only verified on macOS —
	// Windows drive-letter paths ("C:\...") need adapter-specific handling
	// here as a follow-up (see docs/STATUS.md).
	app.RegisterService(application.NewServiceWithOptions(
		fileserver.NewWithConfig(&fileserver.Config{RootPath: "/"}),
		application.ServiceOptions{Route: "/preview"},
	))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Vault",
		Width:  1000,
		Height: 640,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropNormal,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
