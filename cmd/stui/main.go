package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"stui/internal/application"
	"stui/internal/infrastructure/downloader"
	"stui/internal/infrastructure/logger"
	"stui/internal/infrastructure/pathfinder"
	"stui/internal/infrastructure/repository"
	"stui/internal/infrastructure/unlocker"
	"stui/internal/presentation/ui"

	"go.uber.org/zap"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if err := logger.InitLogger(*debugFlag); err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	repo := repository.NewLocalDlcRepository()
	pathFinder := pathfinder.NewPathFinder()
	unlockerInst := unlocker.NewUnlocker()

	downloaderFactory := func() (application.DlcDownloader, error) {
		return downloader.NewTorrentDownloader()
	}

	model := ui.NewModel(repo, pathFinder, unlockerInst, downloaderFactory)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logger.Logger.Error("Error running program", zap.Error(err))
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
