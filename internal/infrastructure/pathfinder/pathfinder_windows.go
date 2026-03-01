//go:build windows

package pathfinder

import (
	"fmt"
	"os"
	"path/filepath"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/infrastructure/logger"

	"go.uber.org/zap"
	"golang.org/x/sys/windows/registry"
)

type PathFinder struct{}

func NewPathFinder() application.IPathFinder {
	return &PathFinder{}
}

func (p *PathFinder) FindGamePath() string {
	logger.Logger.Debug("Starting game path search (Windows)")

	drives := []string{"C", "D", "E", "F", "G", "H"}
	paths := []string{
		`\Program Files (x86)\Steam\steamapps\common\The Sims 4`,
		`\Program Files\Steam\steamapps\common\The Sims 4`,
		`\SteamLibrary\steamapps\common\The Sims 4`,
		`\Program Files\EA Games\The Sims 4`,
		`\Program Files (x86)\EA Games\The Sims 4`,
		`\Program Files (x86)\Origin Games\The Sims 4`,
		`\The Sims 4`,
	}

	for _, drive := range drives {
		for _, path := range paths {
			fullPath := filepath.Join(drive+":", path)
			if _, err := os.Stat(fullPath); err == nil {
				logger.Logger.Debug("Game path found", zap.String("path", fullPath))
				return fullPath
			}
		}
	}

	logger.Logger.Debug("Game path not found")
	return ""
}

func (p *PathFinder) FindEAClient() (*domain.EAClientInfo, error) {
	logger.Logger.Debug("Starting EA client search (Windows)")

	clientPath, err := getClientPathFromRegistry(`SOFTWARE\Electronic Arts\EA Desktop`)
	if err == nil {
		logger.Logger.Debug("EA app found", zap.String("path", clientPath))
		return &domain.EAClientInfo{
			Path:       clientPath,
			ClientType: domain.EAClientTypeEAApp,
		}, nil
	}
	logger.Logger.Debug("EA app not found in registry", zap.Error(err))

	clientPath, err = getClientPathFromRegistry(`SOFTWARE\WOW6432Node\Origin`)
	if err == nil {
		logger.Logger.Debug("Origin found (WOW6432Node)", zap.String("path", clientPath))
		return &domain.EAClientInfo{
			Path:       clientPath,
			ClientType: domain.EAClientTypeOrigin,
		}, nil
	}
	logger.Logger.Debug("Origin not found in WOW6432Node registry", zap.Error(err))

	clientPath, err = getClientPathFromRegistry(`SOFTWARE\Origin`)
	if err == nil {
		logger.Logger.Debug("Origin found", zap.String("path", clientPath))
		return &domain.EAClientInfo{
			Path:       clientPath,
			ClientType: domain.EAClientTypeOrigin,
		}, nil
	}
	logger.Logger.Debug("Origin not found in registry", zap.Error(err))

	logger.Logger.Error("EA app/Origin not found")
	return nil, fmt.Errorf("EA app/Origin not found")
}

func getClientPathFromRegistry(registryPath string) (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	clientPath, _, err := key.GetStringValue("ClientPath")
	if err != nil {
		return "", err
	}

	parentPath := filepath.Dir(clientPath)
	return parentPath, nil
}
