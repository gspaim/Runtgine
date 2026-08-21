//go:build (linux && cgo) || (darwin && cgo) || windows

package desktop

import (
	"encoding/json"
	"errors"
	"io/fs"

	"github.com/gspaim/Runtgine/internal/core/result"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type wailsEmitter struct {
	app *application.App
}

func (e wailsEmitter) Emit(name string, data any) {
	if e.app != nil {
		e.app.Event.Emit(name, data)
	}
}

// Run starts the Wails v3 desktop Entry Point (one window).
func Run(core CoreAPI) error {
	svc := NewService(core)
	app := application.New(application.Options{
		Name:        "Runtgine",
		Description: "Constellation Mission Control",
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Services: []application.Service{
			application.NewService(svc),
		},
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(mustSub(assets, "frontend/dist")),
		},
		MarshalError: marshalCallError,
	})
	if app == nil {
		return result.Runtime(result.CodeInternal, "wails application.New returned nil", false, nil)
	}
	svc.SetEmitter(wailsEmitter{app: app})
	svc.Start()
	defer svc.Stop()

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "RUNTGINE / CONSTELLATION MISSION CONTROL",
		Width:            1280,
		Height:           800,
		MinWidth:         900,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(7, 11, 20),
		URL:              "/",
	})
	return app.Run()
}

func marshalCallError(err error) []byte {
	var re result.Error
	if errors.As(err, &re) {
		b, _ := json.Marshal(re)
		return b
	}
	return nil
}

func mustSub(root fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(root, dir)
	if err != nil {
		return root
	}
	return sub
}
