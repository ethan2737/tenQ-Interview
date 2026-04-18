package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"tenq-interview/internal/workbench"
)

type App struct {
	ctx        context.Context
	service    *workbench.Service
	startupErr error
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	rootDir, _ := os.Getwd()
	service, err := workbench.NewServiceWithOptions(defaultCachePath(), rootDir)
	if err != nil {
		a.startupErr = err
		return
	}

	a.service = service
}

func (a *App) ImportPath(target string) (workbench.ImportResult, error) {
	if err := a.ready(); err != nil {
		return workbench.ImportResult{}, err
	}
	return a.service.ImportPath(target)
}

func (a *App) PrepareImport(target string) (workbench.ImportResult, error) {
	if err := a.ready(); err != nil {
		return workbench.ImportResult{}, err
	}
	return a.service.PrepareImport(target)
}

func (a *App) ProcessDocument(path string, relativePath string, provider string) (workbench.DocumentSummary, error) {
	if err := a.ready(); err != nil {
		return workbench.DocumentSummary{}, err
	}
	return a.service.ProcessDocument(path, relativePath, provider)
}

func (a *App) PreviewDocument(path string) (workbench.DocumentPreview, error) {
	if err := a.ready(); err != nil {
		return workbench.DocumentPreview{}, err
	}
	return a.service.PreviewDocument(path)
}

func (a *App) ListImportedDocuments() (workbench.ImportResult, error) {
	if err := a.ready(); err != nil {
		return workbench.ImportResult{}, err
	}
	return a.service.ListImportedDocuments()
}

func (a *App) AgentSettings() (workbench.AgentSettings, error) {
	if err := a.ready(); err != nil {
		return workbench.AgentSettings{}, err
	}
	return a.service.AgentSettings(), nil
}

func (a *App) ClearImportedDocuments() error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.service.ClearImportedDocuments()
}

func (a *App) SelectMarkdownFile() (string, error) {
	if a.ctx == nil {
		return "", errors.New("application context is not ready")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Markdown 文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Markdown",
				Pattern:     "*.md",
			},
		},
	})
}

func (a *App) SelectMarkdownDirectory() (string, error) {
	if a.ctx == nil {
		return "", errors.New("application context is not ready")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Markdown 目录",
	})
}

func (a *App) ready() error {
	if a.startupErr != nil {
		return a.startupErr
	}
	if a.service == nil {
		return errors.New("service is not ready")
	}
	return nil
}

func defaultCachePath() string {
	baseDir, err := os.UserCacheDir()
	if err != nil || baseDir == "" {
		baseDir = ".cache"
	}
	return filepath.Join(baseDir, "tenq-interview", "index.json")
}
