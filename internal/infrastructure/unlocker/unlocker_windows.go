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
	"golang.org/x/sys/windows"
)

type Unlocker struct{}

func NewUnlocker() application.IUnlocker {
	return &Unlocker{}
}

func isAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

func requireAdmin() error {
	if !isAdmin() {
		return fmt.Errorf("administrator rights required — please restart the application as Administrator (right-click → Run as administrator)")
	}
	return nil
}

func (u *Unlocker) InstallDLL(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting DLL installation (Windows)",
		zap.String("client_type", string(clientInfo.ClientType)),
		zap.String("client_path", clientInfo.Path))

	if err := requireAdmin(); err != nil {
		return err
	}

	if clientInfo.ClientType != domain.EAClientTypeEAApp && clientInfo.ClientType != domain.EAClientTypeOrigin {
		return fmt.Errorf("unsupported client type: %s", clientInfo.ClientType)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	unlockerDir := filepath.Join(filepath.Dir(exePath), "unlocker")
	srcDLL := filepath.Join(unlockerDir, string(clientInfo.ClientType), "version.dll")

	if _, err := os.Stat(srcDLL); err != nil {
		return fmt.Errorf("source DLL not found: %s", srcDLL)
	}

	data, err := os.ReadFile(srcDLL)
	if err != nil {
		return fmt.Errorf("failed to read source DLL: %w", err)
	}

	dstDLL := filepath.Join(clientInfo.Path, "version.dll")
	if err := os.WriteFile(dstDLL, data, 0o644); err != nil {
		logger.Logger.Error("Failed to write DLL", zap.String("path", dstDLL), zap.Error(err))
		return fmt.Errorf("failed to write DLL: %w", err)
	}
	logger.Logger.Debug("DLL installed successfully", zap.String("path", dstDLL))

	if clientInfo.ClientType == domain.EAClientTypeEAApp {
		stagedDir := filepath.Join(filepath.Dir(clientInfo.Path), "StagedEADesktop", "EA Desktop")
		dstDLL2 := filepath.Join(stagedDir, "version.dll")

		if _, err := os.Stat(stagedDir); err == nil {
			if err := os.WriteFile(dstDLL2, data, 0o644); err != nil {
				logger.Logger.Warn("Failed to write staged DLL", zap.String("path", dstDLL2), zap.Error(err))
			} else {
				logger.Logger.Debug("Staged DLL installed", zap.String("path", dstDLL2))
			}
		}

		// Schedule copy of DLL into staged dir on next reboot
		stagedDirWildcard := filepath.Join(stagedDir, "*")
		taskCommand := fmt.Sprintf(`xcopy.exe /Y "%s" "%s"`, dstDLL, stagedDirWildcard)
		cmd := exec.Command("schtasks", "/Create", "/F", "/RL", "HIGHEST", "/SC", "ONCE",
			"/ST", "00:00", "/SD", "01/01/2000", "/TN", "copy_dlc_unlocker", "/TR", taskCommand)
		if err := cmd.Run(); err != nil {
			cmd = exec.Command("schtasks", "/Create", "/F", "/RL", "HIGHEST", "/SC", "ONCE",
				"/ST", "00:00", "/SD", "2000/01/01", "/TN", "copy_dlc_unlocker", "/TR", taskCommand)
			if err := cmd.Run(); err != nil {
				logger.Logger.Warn("Failed to create scheduled task", zap.Error(err))
			}
		}

		machineIniPath := filepath.Join(os.Getenv("ProgramData"), "EA Desktop", "machine.ini")
		logger.Logger.Debug("Updating machine.ini", zap.String("path", machineIniPath))
		f, err := os.OpenFile(machineIniPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			f.WriteString("machine.bgsstandaloneenabled=0\n")
			f.Close()
		} else {
			logger.Logger.Warn("Failed to update machine.ini", zap.Error(err))
		}
	}

	// Also install main config alongside DLL
	srcMainConfig := filepath.Join(unlockerDir, "config.ini")
	if _, err := os.Stat(srcMainConfig); err == nil {
		appDataDir := os.Getenv("AppData")
		if appDataDir != "" {
			configDir := filepath.Join(appDataDir, "anadius", "EA DLC Unlocker v2")
			if err := os.MkdirAll(configDir, 0o755); err == nil {
				if err := copyFile(srcMainConfig, filepath.Join(configDir, "config.ini")); err != nil {
					logger.Logger.Warn("Failed to copy config.ini during DLL install", zap.Error(err))
				} else {
					logger.Logger.Debug("config.ini installed", zap.String("path", configDir))
				}
			}
		}
	}

	logger.Logger.Debug("DLL installation completed successfully")
	return nil
}

func (u *Unlocker) UninstallDLL(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting DLL uninstallation (Windows)",
		zap.String("client_type", string(clientInfo.ClientType)),
		zap.String("client_path", clientInfo.Path))

	if err := requireAdmin(); err != nil {
		return err
	}

	dstDLL := filepath.Join(clientInfo.Path, "version.dll")
	if err := os.Remove(dstDLL); err != nil && !os.IsNotExist(err) {
		logger.Logger.Error("Failed to remove DLL", zap.String("path", dstDLL), zap.Error(err))
		return fmt.Errorf("failed to remove DLL: %w", err)
	}
	logger.Logger.Debug("DLL removed", zap.String("path", dstDLL))

	if clientInfo.ClientType == domain.EAClientTypeEAApp {
		stagedDir := filepath.Join(filepath.Dir(clientInfo.Path), "StagedEADesktop", "EA Desktop")
		dstDLL2 := filepath.Join(stagedDir, "version.dll")
		if err := os.Remove(dstDLL2); err != nil && !os.IsNotExist(err) {
			logger.Logger.Warn("Failed to remove staged DLL", zap.String("path", dstDLL2), zap.Error(err))
		}

		cmd := exec.Command("schtasks", "/Delete", "/TN", "copy_dlc_unlocker", "/F")
		if err := cmd.Run(); err != nil {
			logger.Logger.Warn("Failed to remove scheduled task", zap.Error(err))
		}
	}

	logger.Logger.Debug("DLL uninstallation completed successfully")
	return nil
}

func (u *Unlocker) InstallConfig(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting config installation (Windows)")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	unlockerDir := filepath.Join(filepath.Dir(exePath), "unlocker")
	srcGameConfig := filepath.Join(unlockerDir, "g_The Sims 4.ini")

	if _, err := os.Stat(srcGameConfig); err != nil {
		return fmt.Errorf("game config not found: %s", srcGameConfig)
	}

	appDataDir := os.Getenv("AppData")
	if appDataDir == "" {
		return fmt.Errorf("AppData environment variable not set")
	}

	configDir := filepath.Join(appDataDir, "anadius", "EA DLC Unlocker v2")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := copyFile(srcGameConfig, filepath.Join(configDir, "g_The Sims 4.ini")); err != nil {
		return fmt.Errorf("failed to copy game config: %w", err)
	}
	logger.Logger.Debug("Game config copied", zap.String("path", configDir))

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
