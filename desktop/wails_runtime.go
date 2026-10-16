//go:build desktop

package main

import (
	"context"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func windowMinimise(ctx context.Context) {
	wruntime.WindowMinimise(ctx)
}

func windowToggleMaximise(ctx context.Context) {
	wruntime.WindowToggleMaximise(ctx)
}

func appQuit(ctx context.Context) {
	wruntime.Quit(ctx)
}

func openDirectoryDialog(ctx context.Context, title, defaultDir string) (string, error) {
	return wruntime.OpenDirectoryDialog(ctx, wruntime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	})
}

func browserOpenURL(ctx context.Context, url string) {
	wruntime.BrowserOpenURL(ctx, url)
}

func openFileDialog(ctx context.Context, title string) (string, error) {
	return wruntime.OpenFileDialog(ctx, wruntime.OpenDialogOptions{
		Title: title,
	})
}

func questionDialog(ctx context.Context, title, message string) bool {
	choice, err := wruntime.MessageDialog(ctx, wruntime.MessageDialogOptions{
		Type:          wruntime.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	return err == nil && choice == "Yes"
}
