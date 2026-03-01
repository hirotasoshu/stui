//go:build linux

package pathfinder

import (
	"fmt"
	"os"
	"path/filepath"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/infrastructure/logger"

	"go.uber.org/zap"
)

type PathFinder struct{}

func NewPathFinder() application.IPathFinder {
	return &PathFinder{}
}

func (p *PathFinder) FindGamePath() string {
	logger.Logger.Debug("Starting game path search (Linux)")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Logger.Error("Failed to get home directory", zap.Error(err))
		return ""
	}

	paths := []string{
		filepath.Join(homeDir, ".steam/steam/steamapps/common/The Sims 4"),
		filepath.Join(homeDir, ".local/share/Steam/steamapps/common/The Sims 4"),
		filepath.Join(homeDir, "Downloads/Games/The Sims 4"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			logger.Logger.Debug("Game path found", zap.String("path", path))
			return path
		}
	}

	logger.Logger.Debug("Game path not found")
	return ""
}

func (p *PathFinder) FindEAClient() (*domain.EAClientInfo, error) {
	logger.Logger.Debug("Starting EA client search (Linux)")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Logger.Error("Failed to get home directory", zap.Error(err))
		return nil, err
	}

	steamPaths := []string{
		filepath.Join(homeDir, ".steam/steam"),
		filepath.Join(homeDir, ".local/share/Steam"),
		filepath.Join(homeDir, "snap/steam/common/.local/share/Steam"),
		filepath.Join(homeDir, "steam/root"),
		filepath.Join(homeDir, "steam"),
		filepath.Join(homeDir, ".var/app/com.valvesoftware.Steam/.steam/steam"),
		filepath.Join(homeDir, ".var/app/com.valvesoftware.Steam/.local/share/Steam"),
		filepath.Join(homeDir, ".var/app/com.valvesoftware.Steam/.steam/root"),
		filepath.Join(homeDir, ".var/app/com.valvesoftware.Steam/.steam"),
	}

	lutrisPath := filepath.Join(homeDir, "Games")

	bottlesPaths := []string{
		filepath.Join(homeDir, ".var/app/com.usebottles.bottles/data/bottles/bottles"),
		filepath.Join(homeDir, ".local/share/bottles"),
	}

	logger.Logger.Debug("Checking Wine prefix")
	if prefixPath := checkWinePrefix(); prefixPath != "" {
		logger.Logger.Debug("EA app found in Wine prefix", zap.String("path", prefixPath))
		return &domain.EAClientInfo{
			WinePrefix:   prefixPath,
			PrefixSource: domain.WinePrefixSourceWine,
		}, nil
	}

	logger.Logger.Debug("Checking Steam prefixes")
	for _, steamPath := range steamPaths {
		if prefixPath := checkSteamPrefixes(steamPath); prefixPath != "" {
			logger.Logger.Debug("EA app found in Steam prefix", zap.String("path", prefixPath))
			return &domain.EAClientInfo{
				WinePrefix:   prefixPath,
				PrefixSource: domain.WinePrefixSourceSteam,
			}, nil
		}
	}

	logger.Logger.Debug("Checking Lutris prefixes")
	if prefixPath := checkLutrisPrefixes(lutrisPath); prefixPath != "" {
		logger.Logger.Debug("EA app found in Lutris prefix", zap.String("path", prefixPath))
		return &domain.EAClientInfo{
			WinePrefix:   prefixPath,
			PrefixSource: domain.WinePrefixSourceLutris,
		}, nil
	}

	logger.Logger.Debug("Checking Bottles prefixes")
	for _, bottlesPath := range bottlesPaths {
		if prefixPath := checkBottlesPrefixes(bottlesPath); prefixPath != "" {
			logger.Logger.Debug("EA app found in Bottles prefix", zap.String("path", prefixPath))
			return &domain.EAClientInfo{
				WinePrefix:   prefixPath,
				PrefixSource: domain.WinePrefixSourceBottles,
			}, nil
		}
	}

	logger.Logger.Error("EA app not found in any Wine/Proton prefix")
	return nil, fmt.Errorf("EA app not found in any Wine/Proton prefix")
}

func checkWinePrefix() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	winePrefix := os.Getenv("WINEPREFIX")
	if winePrefix == "" {
		winePrefix = filepath.Join(homeDir, ".wine")
	}

	if hasEADesktopDirs(winePrefix) {
		return winePrefix
	}

	return ""
}

func checkSteamPrefixes(steamPath string) string {
	steamappsPath := filepath.Join(steamPath, "steamapps")
	if _, err := os.Stat(steamappsPath); err != nil {
		return ""
	}

	compatdataPath := filepath.Join(steamappsPath, "compatdata")
	entries, err := os.ReadDir(compatdataPath)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		prefixPath := filepath.Join(compatdataPath, entry.Name(), "pfx")
		if hasEADesktopDirs(prefixPath) {
			return prefixPath
		}
	}

	return ""
}

func checkLutrisPrefixes(lutrisPath string) string {
	entries, err := os.ReadDir(lutrisPath)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		prefixPath := filepath.Join(lutrisPath, entry.Name())
		if hasEADesktopDirs(prefixPath) {
			return prefixPath
		}
	}

	return ""
}

func checkBottlesPrefixes(bottlesPath string) string {
	entries, err := os.ReadDir(bottlesPath)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		prefixPath := filepath.Join(bottlesPath, entry.Name())
		if hasEADesktopDirs(prefixPath) {
			return prefixPath
		}
	}

	return ""
}

func hasEADesktopDirs(prefixPath string) bool {
	baseDir := filepath.Join(prefixPath, "drive_c/Program Files/Electronic Arts/EA Desktop")

	if _, err := os.Stat(baseDir); err != nil {
		return false
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		eaDesktopSubdir := filepath.Join(baseDir, entry.Name(), "EA Desktop")
		if _, err := os.Stat(eaDesktopSubdir); err == nil {
			return true
		}
	}

	return false
}
