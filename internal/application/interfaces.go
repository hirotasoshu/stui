package application

import "stui/internal/domain"

type DownloadProgress struct {
	BytesDownloaded int64
	TotalBytes      int64
	Speed           int64
	IsComplete      bool
}

type DlcDownloader interface {
	Download(dlcs []domain.DLC, gamePath string) error
	GetProgress() DownloadProgress
	Stop() error
	Finalize() error
}

type IUnlocker interface {
	InstallDLL(clientInfo domain.EAClientInfo) error
	UninstallDLL(clientInfo domain.EAClientInfo) error
	InstallConfig(clientInfo domain.EAClientInfo) error
}

type IPathFinder interface {
	FindGamePath() string
	FindEAClient() (*domain.EAClientInfo, error)
}

type DlcRepository interface {
	GetExpansionPacks() []domain.DLC
	GetFreePacks() []domain.DLC
	GetGamePacks() []domain.DLC
	GetStuffPacks() []domain.DLC
	GetKits() []domain.DLC
	GetInstalledExpansionPacks(gamePath string) []domain.DLC
	GetInstalledFreePacks(gamePath string) []domain.DLC
	GetInstalledGamePacks(gamePath string) []domain.DLC
	GetInstalledStuffPacks(gamePath string) []domain.DLC
	GetInstalledKits(gamePath string) []domain.DLC
}
