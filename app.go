package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

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

	cwd, _ := os.Getwd()
	exePath, exeErr := os.Executable()
	exeDir := ""
	if exeErr == nil {
		exeDir = filepath.Dir(exePath)
	}

	userConfigDir := ""
	if configDir, configErr := os.UserConfigDir(); configErr == nil {
		userConfigDir = filepath.Join(configDir, "tenq-interview")
	}

	service, err := workbench.NewServiceWithOptions(defaultCachePath(), configRootsFor(cwd, exeDir, userConfigDir, os.Getenv("TENQ_INTERVIEW_CONFIG_DIR"))...)
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

func (a *App) GenerateInterviewAudioFromCache() (workbench.AudioGenerationResult, error) {
	if err := a.ready(); err != nil {
		return workbench.AudioGenerationResult{}, err
	}
	return a.service.GenerateInterviewAudioFromCache()
}

func (a *App) StartInterviewAudioGenerationFromCache() (workbench.AudioGenerationStatus, error) {
	if err := a.ready(); err != nil {
		return workbench.AudioGenerationStatus{}, err
	}
	return a.service.StartInterviewAudioGenerationFromCache()
}

func (a *App) AudioGenerationStatus() (workbench.AudioGenerationStatus, error) {
	if err := a.ready(); err != nil {
		return workbench.AudioGenerationStatus{}, err
	}
	return a.service.AudioGenerationStatus(), nil
}

func (a *App) CancelInterviewAudioGeneration() (workbench.AudioGenerationStatus, error) {
	if err := a.ready(); err != nil {
		return workbench.AudioGenerationStatus{}, err
	}
	return a.service.CancelInterviewAudioGeneration()
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

func (a *App) SelectMarkdownExportPath() (string, error) {
	if a.ctx == nil {
		return "", errors.New("application context is not ready")
	}
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "选择导出 Markdown",
		DefaultFilename: "document.md",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Markdown",
				Pattern:     "*.md",
			},
		},
	})
}

func (a *App) ExportDocumentMarkdown(title string, answer string, outputPath string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.service.ExportDocumentMarkdown(title, answer, outputPath)
}

func (a *App) ExportDocumentsMarkdown(documents []workbench.MarkdownExportDocument, outputPath string) error {
	if err := a.ready(); err != nil {
		return err
	}
	return a.service.ExportDocumentsMarkdown(documents, outputPath)
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
	return filepath.Join(defaultCacheBaseDir(), ".cache", "tenq-interview", "index.json")
}

func defaultCacheBaseDir() string {
	executablePath, err := os.Executable()
	if err == nil {
		executableDir := strings.TrimSpace(filepath.Dir(executablePath))
		if executableDir != "" {
			return executableDir
		}
	}

	workingDir, err := os.Getwd()
	if err == nil {
		workingDir = strings.TrimSpace(workingDir)
		if workingDir != "" {
			return workingDir
		}
	}

	return "."
}

func configRootsFor(cwd string, exeDir string, userConfigDir string, explicitConfigDir string) []string {
	candidates := []string{
		strings.TrimSpace(explicitConfigDir),
		strings.TrimSpace(exeDir),
		strings.TrimSpace(cwd),
		strings.TrimSpace(userConfigDir),
	}

	roots := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, exists := seen[candidate]; exists {
			continue
		}
		seen[candidate] = struct{}{}
		roots = append(roots, candidate)
	}
	return roots
}
