package app

import (
	"context"
	"fmt"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) StartContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SelectFile() ([]string, error) {
	filePaths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Seleccionar archivos para enviar",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Todos los archivos (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return filePaths, nil
}
