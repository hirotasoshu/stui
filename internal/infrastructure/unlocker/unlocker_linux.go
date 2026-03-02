//go:build linux

package unlocker

import (
	"fmt"
	"os"
	"path/filepath"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/infrastructure/logger"

	"go.uber.org/zap"
)

type Unlocker struct{}

func NewUnlocker() application.IUnlocker {
	return &Unlocker{}
}

func (u *Unlocker) InstallDLL(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting DLL installation (Linux)",
		zap.String("wine_prefix", clientInfo.WinePrefix),
		zap.String("prefix_source", string(clientInfo.PrefixSource)))

	exePath, err := os.Executable()
	if err != nil {
		logger.Logger.Error("Failed to get executable path", zap.Error(err))
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	unlockerDir := filepath.Join(filepath.Dir(exePath), "unlocker")
	srcDLL := filepath.Join(unlockerDir, "ea_app", "version.dll")
	logger.Logger.Debug("Source DLL path", zap.String("path", srcDLL))

	if _, err := os.Stat(srcDLL); err != nil {
		logger.Logger.Error("Source DLL not found", zap.String("path", srcDLL), zap.Error(err))
		return fmt.Errorf("source DLL not found: %s", srcDLL)
	}

	data, err := os.ReadFile(srcDLL)
	if err != nil {
		logger.Logger.Error("Failed to read source DLL", zap.Error(err))
		return fmt.Errorf("failed to read source DLL: %w", err)
	}

	baseDir := filepath.Join(clientInfo.WinePrefix, "drive_c/Program Files/Electronic Arts/EA Desktop")
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		logger.Logger.Error("Failed to read EA Desktop directory", zap.String("path", baseDir), zap.Error(err))
		return fmt.Errorf("failed to read EA Desktop directory: %w", err)
	}

	installedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		eaDesktopDir := filepath.Join(baseDir, entry.Name(), "EA Desktop")
		if _, err := os.Stat(eaDesktopDir); err == nil {
			dstDLL := filepath.Join(eaDesktopDir, "version.dll")
			logger.Logger.Debug("Installing DLL", zap.String("path", dstDLL))
			if err := os.WriteFile(dstDLL, data, 0o644); err != nil {
				logger.Logger.Error("Failed to write DLL", zap.String("path", dstDLL), zap.Error(err))
				return fmt.Errorf("failed to write DLL to %s: %w", dstDLL, err)
			}
			logger.Logger.Debug("DLL installed", zap.String("path", dstDLL))
			installedCount++
		}
	}

	if installedCount == 0 {
		logger.Logger.Error("No EA Desktop directories found in base directory", zap.String("base_dir", baseDir))
		return fmt.Errorf("no EA Desktop directories found")
	}

	regFile := filepath.Join(clientInfo.WinePrefix, "user.reg")
	if err := addDllOverride(regFile); err != nil {
		logger.Logger.Error("Failed to add DLL override to registry", zap.String("reg_file", regFile), zap.Error(err))
		return fmt.Errorf("failed to add DLL override to registry: %w", err)
	}
	logger.Logger.Debug("DLL override added to registry", zap.String("reg_file", regFile))

	// Also install main config alongside DLL
	srcMainConfig := filepath.Join(unlockerDir, "config.ini")
	if _, err := os.Stat(srcMainConfig); err == nil {
		username := getPrefixUsername(clientInfo.WinePrefix, clientInfo.PrefixSource)
		configDir := filepath.Join(clientInfo.WinePrefix, "drive_c", "users", username,
			"AppData", "Roaming", "anadius", "EA DLC Unlocker v2")
		if err := os.MkdirAll(configDir, 0o755); err == nil {
			if err := copyFile(srcMainConfig, filepath.Join(configDir, "config.ini")); err != nil {
				logger.Logger.Warn("Failed to copy config.ini during DLL install", zap.Error(err))
			} else {
				logger.Logger.Debug("config.ini installed", zap.String("path", configDir))
			}
		}
	}

	logger.Logger.Debug("DLL installation completed successfully", zap.Int("installed_count", installedCount))
	return nil
}

func (u *Unlocker) UninstallDLL(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting DLL uninstallation (Linux)",
		zap.String("wine_prefix", clientInfo.WinePrefix))

	baseDir := filepath.Join(clientInfo.WinePrefix, "drive_c/Program Files/Electronic Arts/EA Desktop")
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		logger.Logger.Error("Failed to read EA Desktop directory", zap.String("path", baseDir), zap.Error(err))
		return fmt.Errorf("failed to read EA Desktop directory: %w", err)
	}

	removedCount := 0
	var lastErr error

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		eaDesktopDir := filepath.Join(baseDir, entry.Name(), "EA Desktop")
		dstDLL := filepath.Join(eaDesktopDir, "version.dll")

		if err := os.Remove(dstDLL); err != nil && !os.IsNotExist(err) {
			logger.Logger.Error("Failed to remove DLL", zap.String("path", dstDLL), zap.Error(err))
			lastErr = err
		} else if err == nil {
			logger.Logger.Debug("DLL removed", zap.String("path", dstDLL))
			removedCount++
		}
	}

	logger.Logger.Debug("DLL uninstallation completed", zap.Int("removed_count", removedCount))
	return lastErr
}

func (u *Unlocker) InstallConfig(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting config installation (Linux)")

	exePath, err := os.Executable()
	if err != nil {
		logger.Logger.Error("Failed to get executable path", zap.Error(err))
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	unlockerDir := filepath.Join(filepath.Dir(exePath), "unlocker")
	srcGameConfig := filepath.Join(unlockerDir, "g_The Sims 4.ini")
	logger.Logger.Debug("Source game config", zap.String("game_config", srcGameConfig))

	if _, err := os.Stat(srcGameConfig); err != nil {
		logger.Logger.Error("Game config not found", zap.String("path", srcGameConfig), zap.Error(err))
		return fmt.Errorf("game config not found: %s", srcGameConfig)
	}

	username := getPrefixUsername(clientInfo.WinePrefix, clientInfo.PrefixSource)
	logger.Logger.Debug("Wine prefix username", zap.String("username", username))

	configDir := filepath.Join(clientInfo.WinePrefix, "drive_c", "users", username,
		"AppData", "Roaming", "anadius", "EA DLC Unlocker v2")
	logger.Logger.Debug("Config directory", zap.String("path", configDir))

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		logger.Logger.Error("Failed to create config directory", zap.String("path", configDir), zap.Error(err))
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	logger.Logger.Debug("Config directory created")

	dstGameConfig := filepath.Join(configDir, "g_The Sims 4.ini")
	if err := copyFile(srcGameConfig, dstGameConfig); err != nil {
		logger.Logger.Error("Failed to copy game config", zap.Error(err))
		return fmt.Errorf("failed to copy game config: %w", err)
	}
	logger.Logger.Debug("Game config copied", zap.String("path", dstGameConfig))

	logger.Logger.Debug("Config installation completed successfully")
	return nil
}

func getPrefixUsername(prefixPath string, source domain.WinePrefixSource) string {
	if source == domain.WinePrefixSourceSteam {
		return "steamuser"
	}

	usersDir := filepath.Join(prefixPath, "drive_c", "users")
	entries, err := os.ReadDir(usersDir)
	if err != nil {
		return os.Getenv("USER")
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "Public" {
			return entry.Name()
		}
	}

	return os.Getenv("USER")
}

func addDllOverride(regFile string) error {
	timestamp := os.Getenv("EPOCHSECONDS")
	if timestamp == "" {
		timestamp = "1234567890"
	}

	override := fmt.Sprintf("\n\n[Software\\\\Wine\\\\DllOverrides] %s\n\"version\"=\"native,builtin\"\n", timestamp)

	f, err := os.OpenFile(regFile, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(override)
	return err
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
