//go:build windows

package unlocker

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/infrastructure/logger"

	"go.uber.org/zap"
)

type Unlocker struct{}

func NewUnlocker() application.IUnlocker {
	return &Unlocker{}
}

// ps escapes a string for use inside a single-quoted PowerShell string literal.
func psQ(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// encodePSCommand encodes a PowerShell script as UTF-16LE base64 for -EncodedCommand.
func encodePSCommand(cmd string) string {
	runes := utf16.Encode([]rune(cmd))
	b := make([]byte, len(runes)*2)
	for i, r := range runes {
		b[i*2] = byte(r)
		b[i*2+1] = byte(r >> 8)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// runPSElevated runs a PowerShell script with UAC elevation (-Verb RunAs -Wait).
// If already admin, UAC prompt is skipped automatically.
func runPSElevated(psScript string) error {
	encoded := encodePSCommand(psScript)
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		fmt.Sprintf(`Start-Process powershell -Verb RunAs -Wait -ArgumentList '-NoProfile -NonInteractive -EncodedCommand %s'`, encoded),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("elevated script failed: %w (output: %s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (u *Unlocker) InstallDLL(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting DLL installation (Windows)",
		zap.String("client_type", string(clientInfo.ClientType)),
		zap.String("client_path", clientInfo.Path))

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

	dstDLL := filepath.Join(clientInfo.Path, "version.dll")
	stagedDir := filepath.Join(filepath.Dir(clientInfo.Path), "StagedEADesktop", "EA Desktop")
	dstDLL2 := filepath.Join(stagedDir, "version.dll")
	machineIniPath := filepath.Join(os.Getenv("ProgramData"), "EA Desktop", "machine.ini")

	var sb strings.Builder
	sb.WriteString("$ErrorActionPreference = 'Stop'\n")
	fmt.Fprintf(&sb, "Copy-Item -LiteralPath '%s' -Destination '%s' -Force\n", psQ(srcDLL), psQ(dstDLL))

	if clientInfo.ClientType == domain.EAClientTypeEAApp {
		fmt.Fprintf(&sb, "If (Test-Path -LiteralPath '%s') { Copy-Item -LiteralPath '%s' -Destination '%s' -Force }\n",
			psQ(stagedDir), psQ(srcDLL), psQ(dstDLL2))

		stagedDirWildcard := filepath.Join(stagedDir, "*")
		taskCmd := fmt.Sprintf(`xcopy.exe /Y '%s' '%s'`, psQ(dstDLL), psQ(stagedDirWildcard))
		fmt.Fprintf(&sb,
			"$r = & schtasks /Create /F /RL HIGHEST /SC ONCE /ST 00:00 /SD 01/01/2000 /TN copy_dlc_unlocker /TR \"%s\" 2>&1\n"+
				"If ($LASTEXITCODE -Ne 0) { & schtasks /Create /F /RL HIGHEST /SC ONCE /ST 00:00 /SD 2000/01/01 /TN copy_dlc_unlocker /TR \"%s\" 2>&1 | Out-Null }\n",
			taskCmd, taskCmd)

		fmt.Fprintf(&sb, "Add-Content -LiteralPath '%s' -Value 'machine.bgsstandaloneenabled=0' -Encoding utf8 -Force\n",
			psQ(machineIniPath))
	}

	logger.Logger.Debug("Running elevated DLL install script")
	if err := runPSElevated(sb.String()); err != nil {
		logger.Logger.Error("DLL installation failed", zap.Error(err))
		return fmt.Errorf("DLL installation failed: %w", err)
	}

	logger.Logger.Debug("DLL installation completed successfully")
	return nil
}

func (u *Unlocker) UninstallDLL(clientInfo domain.EAClientInfo) error {
	logger.Logger.Debug("Starting DLL uninstallation (Windows)",
		zap.String("client_type", string(clientInfo.ClientType)),
		zap.String("client_path", clientInfo.Path))

	dstDLL := filepath.Join(clientInfo.Path, "version.dll")
	stagedDir := filepath.Join(filepath.Dir(clientInfo.Path), "StagedEADesktop", "EA Desktop")
	dstDLL2 := filepath.Join(stagedDir, "version.dll")

	var sb strings.Builder
	sb.WriteString("$ErrorActionPreference = 'Stop'\n")
	fmt.Fprintf(&sb, "If (Test-Path -LiteralPath '%s') { Remove-Item -LiteralPath '%s' -Force }\n", psQ(dstDLL), psQ(dstDLL))
	fmt.Fprintf(&sb, "If (Test-Path -LiteralPath '%s') { Remove-Item -LiteralPath '%s' -Force }\n", psQ(dstDLL2), psQ(dstDLL2))

	if clientInfo.ClientType == domain.EAClientTypeEAApp {
		sb.WriteString("& schtasks /Delete /TN copy_dlc_unlocker /F 2>&1 | Out-Null\n")
	}

	logger.Logger.Debug("Running elevated DLL uninstall script")
	if err := runPSElevated(sb.String()); err != nil {
		logger.Logger.Error("DLL uninstallation failed", zap.Error(err))
		return fmt.Errorf("DLL uninstallation failed: %w", err)
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
	srcMainConfig := filepath.Join(unlockerDir, "config.ini")
	srcGameConfig := filepath.Join(unlockerDir, "g_The Sims 4.ini")

	if _, err := os.Stat(srcMainConfig); err != nil {
		return fmt.Errorf("main config not found: %s", srcMainConfig)
	}
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

	if err := copyFile(srcMainConfig, filepath.Join(configDir, "config.ini")); err != nil {
		return fmt.Errorf("failed to copy main config: %w", err)
	}
	if err := copyFile(srcGameConfig, filepath.Join(configDir, "g_The Sims 4.ini")); err != nil {
		return fmt.Errorf("failed to copy game config: %w", err)
	}

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
