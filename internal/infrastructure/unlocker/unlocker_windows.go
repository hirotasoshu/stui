//go:build windows

package unlocker

import (
	"fmt"
	"os"
	"os/exec"
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
	logger.Logger.Debug("Starting DLL installation (Windows)",
		zap.String("client_type", string(clientInfo.ClientType)),
		zap.String("client_path", clientInfo.Path))

	if clientInfo.ClientType != domain.EAClientTypeEAApp && clientInfo.ClientType != domain.EAClientTypeOrigin {
		logger.Logger.Error("Unsupported client type", zap.String("client_type", string(clientInfo.ClientType)))
		return fmt.Errorf("unsupported client type: %s", clientInfo.ClientType)
	}

	exePath, err := os.Executable()
	if err != nil {
		logger.Logger.Error("Failed to get executable path", zap.Error(err))
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	unlockerDir := filepath.Join(filepath.Dir(exePath), "unlocker")
	srcDLL := filepath.Join(unlockerDir, string(clientInfo.ClientType), "version.dll")
	logger.Logger.Debug("Source DLL path", zap.String("path", srcDLL))

	if _, err := os.Stat(srcDLL); err != nil {
		logger.Logger.Error("Source DLL not found", zap.String("path", srcDLL), zap.Error(err))
		return fmt.Errorf("source DLL not found: %s", srcDLL)
	}

	dstDLL := filepath.Join(clientInfo.Path, "version.dll")
	logger.Logger.Debug("Destination DLL path", zap.String("path", dstDLL))

	data, err := os.ReadFile(srcDLL)
	if err != nil {
		logger.Logger.Error("Failed to read source DLL", zap.Error(err))
		return fmt.Errorf("failed to read source DLL: %w", err)
	}

	if err := os.WriteFile(dstDLL, data, 0o644); err != nil {
		logger.Logger.Error("Failed to write DLL", zap.String("path", dstDLL), zap.Error(err))
		return fmt.Errorf("failed to write DLL: %w", err)
	}
	logger.Logger.Debug("DLL installed successfully", zap.String("path", dstDLL))

	stagedDir := filepath.Join(filepath.Dir(clientInfo.Path), "StagedEADesktop", "EA Desktop")
	dstDLL2 := filepath.Join(stagedDir, "version.dll")

	if clientInfo.ClientType == domain.EAClientTypeEAApp {
		logger.Logger.Debug("Installing for EA App")

		if _, err := os.Stat(stagedDir); err == nil {
			if err := os.WriteFile(dstDLL2, data, 0o644); err != nil {
				logger.Logger.Error("Failed to write staged DLL", zap.String("path", dstDLL2), zap.Error(err))
				return fmt.Errorf("failed to write staged DLL: %w", err)
			}
			logger.Logger.Debug("Staged DLL installed", zap.String("path", dstDLL2))
		}

		stagedDirWildcard := filepath.Join(stagedDir, "*")
		taskCommand := fmt.Sprintf(`xcopy.exe /Y "%s" "%s"`, dstDLL, stagedDirWildcard)
		logger.Logger.Debug("Creating scheduled task", zap.String("command", taskCommand))

		cmd := exec.Command("schtasks", "/Create", "/F", "/RL", "HIGHEST", "/SC", "ONCE",
			"/ST", "00:00", "/SD", "01/01/2000", "/TN", "copy_dlc_unlocker", "/TR", taskCommand)
		if err := cmd.Run(); err != nil {
			logger.Logger.Debug("First schtasks attempt failed, trying alternative date format", zap.Error(err))
			cmd = exec.Command("schtasks", "/Create", "/F", "/RL", "HIGHEST", "/SC", "ONCE",
				"/ST", "00:00", "/SD", "2000/01/01", "/TN", "copy_dlc_unlocker", "/TR", taskCommand)
			if err := cmd.Run(); err != nil {
				logger.Logger.Warn("Failed to create scheduled task", zap.Error(err))
			} else {
				logger.Logger.Debug("Scheduled task created successfully (alternative format)")
			}
		} else {
			logger.Logger.Debug("Scheduled task created successfully")
		}

		machineIniPath := filepath.Join(os.Getenv("ProgramData"), "EA Desktop", "machine.ini")
		logger.Logger.Debug("Updating machine.ini", zap.String("path", machineIniPath))
		f, err := os.OpenFile(machineIniPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			f.WriteString("machine.bgsstandaloneenabled=0\n")
			f.Close()
			logger.Logger.Debug("machine.ini updated")
		} else {
			logger.Logger.Warn("Failed to update machine.ini", zap.Error(err))
		}
	}

	logger.Logger.Debug("DLL installation completed successfully")
	return nil
}

func (u *Unlocker) UninstallDLL(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting DLL uninstallation (Windows)",
		zap.String("client_type", string(clientInfo.ClientType)),
		zap.String("client_path", clientInfo.Path))

	dstDLL := filepath.Join(clientInfo.Path, "version.dll")
	if err := os.Remove(dstDLL); err != nil && !os.IsNotExist(err) {
		logger.Logger.Error("Failed to remove DLL", zap.String("path", dstDLL), zap.Error(err))
		return fmt.Errorf("failed to remove DLL: %w", err)
	}
	logger.Logger.Debug("DLL removed", zap.String("path", dstDLL))

	stagedDir := filepath.Join(filepath.Dir(clientInfo.Path), "StagedEADesktop", "EA Desktop")
	dstDLL2 := filepath.Join(stagedDir, "version.dll")
	if err := os.Remove(dstDLL2); err != nil && !os.IsNotExist(err) {
		logger.Logger.Error("Failed to remove staged DLL", zap.String("path", dstDLL2), zap.Error(err))
		return fmt.Errorf("failed to remove staged DLL: %w", err)
	}
	logger.Logger.Debug("Staged DLL removed", zap.String("path", dstDLL2))

	if clientInfo.ClientType == domain.EAClientTypeEAApp {
		logger.Logger.Debug("Removing scheduled task")
		cmd := exec.Command("schtasks", "/Delete", "/TN", "copy_dlc_unlocker", "/F")
		if err := cmd.Run(); err != nil {
			logger.Logger.Warn("Failed to remove scheduled task", zap.Error(err))
		} else {
			logger.Logger.Debug("Scheduled task removed")
		}
	}

	logger.Logger.Debug("DLL uninstallation completed successfully")
	return nil
}

func (u *Unlocker) InstallConfig(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting config installation (Windows)")

	exePath, err := os.Executable()
	if err != nil {
		logger.Logger.Error("Failed to get executable path", zap.Error(err))
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	unlockerDir := filepath.Join(filepath.Dir(exePath), "unlocker")
	srcMainConfig := filepath.Join(unlockerDir, "config.ini")
	srcGameConfig := filepath.Join(unlockerDir, "g_The Sims 4.ini")
	logger.Logger.Debug("Source configs",
		zap.String("main_config", srcMainConfig),
		zap.String("game_config", srcGameConfig))

	if _, err := os.Stat(srcMainConfig); err != nil {
		logger.Logger.Error("Main config not found", zap.String("path", srcMainConfig), zap.Error(err))
		return fmt.Errorf("main config not found: %s", srcMainConfig)
	}
	if _, err := os.Stat(srcGameConfig); err != nil {
		logger.Logger.Error("Game config not found", zap.String("path", srcGameConfig), zap.Error(err))
		return fmt.Errorf("game config not found: %s", srcGameConfig)
	}

	appDataDir := os.Getenv("AppData")
	if appDataDir == "" {
		logger.Logger.Error("AppData environment variable not set")
		return fmt.Errorf("AppData environment variable not set")
	}

	configDir := filepath.Join(appDataDir, "anadius", "EA DLC Unlocker v2")
	logger.Logger.Debug("Config directory", zap.String("path", configDir))

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		logger.Logger.Error("Failed to create config directory", zap.String("path", configDir), zap.Error(err))
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	logger.Logger.Debug("Config directory created")

	dstMainConfig := filepath.Join(configDir, "config.ini")
	dstGameConfig := filepath.Join(configDir, "g_The Sims 4.ini")

	if err := copyFile(srcMainConfig, dstMainConfig); err != nil {
		logger.Logger.Error("Failed to copy main config", zap.Error(err))
		return fmt.Errorf("failed to copy main config: %w", err)
	}
	logger.Logger.Debug("Main config copied", zap.String("path", dstMainConfig))

	if err := copyFile(srcGameConfig, dstGameConfig); err != nil {
		logger.Logger.Error("Failed to copy game config", zap.Error(err))
		return fmt.Errorf("failed to copy game config: %w", err)
	}
	logger.Logger.Debug("Game config copied", zap.String("path", dstGameConfig))

	logger.Logger.Debug("Config installation completed successfully")
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
