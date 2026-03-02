package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/infrastructure/logger"

	"github.com/anacrolix/torrent"
	"go.uber.org/zap"
)

const magnetLink = "magnet:?xt=urn:btih:DBF793BE2FECECA6BACA0B836B283F399CF70F53&tr=http%3A%2F%2Fbt2.t-ru.org%2Fann%3Fmagnet&dn=%5BDL%5D%20The%20Sims%204%20%5BP%5D%20%5BRUS%20%2B%20ENG%20%2B%2016%5D%20(2014%2C%20Simulation)%20(1.121.372.1020%20%2B%20111%20DLC)%20%5BPortable%5D"

type TorrentDownloader struct {
	client        *torrent.Client
	torrent       *torrent.Torrent
	selectedSize  int64
	selectedDLCs  map[string]bool
	gamePath      string
	lastBytes     int64
	lastTime      time.Time
	speed         int64
	selectedFiles []*torrent.File
}

func NewTorrentDownloader() (*TorrentDownloader, error) {
	return &TorrentDownloader{
		lastTime: time.Now(),
	}, nil
}

func (d *TorrentDownloader) Download(dlcs []domain.DLC, gamePath string) error {
	d.gamePath = gamePath

	logger.Logger.Debug("Starting download",
		zap.Int("dlc_count", len(dlcs)),
		zap.String("game_path", gamePath),
		zap.String("temp_dir", os.TempDir()))

	cfg := torrent.NewDefaultClientConfig()
	cfg.Seed = false
	cfg.DataDir = os.TempDir()

	client, err := torrent.NewClient(cfg)
	if err != nil {
		logger.Logger.Error("Failed to create torrent client", zap.Error(err))
		return fmt.Errorf("failed to create torrent client: %w", err)
	}
	d.client = client

	logger.Logger.Debug("Adding magnet link")
	t, err := d.client.AddMagnet(magnetLink)
	if err != nil {
		logger.Logger.Error("Failed to add magnet link", zap.Error(err))
		return fmt.Errorf("failed to add magnet link: %w", err)
	}

	d.torrent = t

	logger.Logger.Debug("Waiting for torrent info")
	<-d.torrent.GotInfo()
	logger.Logger.Debug("Got torrent info", zap.Int("total_files", len(d.torrent.Files())))

	files := d.torrent.Files()
	d.selectedDLCs = make(map[string]bool)

	for _, dlc := range dlcs {
		d.selectedDLCs[dlc.Code] = true
		logger.Logger.Debug("Selected DLC", zap.String("code", dlc.Code), zap.String("name", dlc.Name))
	}

	downloadCount := 0
	d.selectedSize = 0
	d.selectedFiles = make([]*torrent.File, 0)

	for _, file := range files {
		filePath := file.Path()
		parts := strings.Split(filePath, "/")

		if len(parts) < 2 {
			continue
		}

		secondDir := parts[1]

		shouldDownload := false
		reason := ""

		if d.selectedDLCs[secondDir] {
			shouldDownload = true
			reason = fmt.Sprintf("matches DLC code: %s", secondDir)
		}

		if len(parts) >= 4 && secondDir == "__Installer" && parts[2] == "DLC" {
			dlcCode := parts[3]
			if d.selectedDLCs[dlcCode] {
				shouldDownload = true
				reason = fmt.Sprintf("matches installer DLC code: %s", dlcCode)
			}
		}

		if shouldDownload {
			file.Download()
			downloadCount++
			d.selectedSize += file.Length()
			d.selectedFiles = append(d.selectedFiles, file)
			logger.Logger.Debug("Downloading file",
				zap.String("path", filePath),
				zap.Int64("size", file.Length()),
				zap.String("reason", reason))
		}
	}

	logger.Logger.Debug("Download summary",
		zap.Int("total_files", downloadCount),
		zap.Int64("total_size_bytes", d.selectedSize),
		zap.Int64("total_size_mb", d.selectedSize/(1024*1024)))

	return nil
}

func (d *TorrentDownloader) MoveDLCs() error {
	logger.Logger.Debug("Moving DLCs from temp to game path")

	tempGamePath := filepath.Join(os.TempDir(), "The Sims 4")

	// Move DLC folders from temp to game path
	entries, err := os.ReadDir(tempGamePath)
	if err != nil {
		logger.Logger.Error("Failed to read temp game path", zap.String("path", tempGamePath), zap.Error(err))
	} else {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			if isDLCCode(name) && d.selectedDLCs[name] {
				srcPath := filepath.Join(tempGamePath, name)
				dstPath := filepath.Join(d.gamePath, name)
				logger.Logger.Debug("Moving DLC", zap.String("from", srcPath), zap.String("to", dstPath))

				if _, err := os.Stat(dstPath); err == nil {
					if err := os.RemoveAll(dstPath); err != nil {
						logger.Logger.Error("Failed to remove existing DLC", zap.String("path", dstPath), zap.Error(err))
						return fmt.Errorf("failed to remove existing DLC %s: %w", name, err)
					}
				}

				err := os.Rename(srcPath, dstPath)
				if err != nil {
					logger.Logger.Error("Failed to move DLC", zap.String("from", srcPath), zap.String("to", dstPath), zap.Error(err))
					return fmt.Errorf("failed to move DLC %s: %w", name, err)
				}
				logger.Logger.Debug("Successfully moved DLC", zap.String("path", dstPath))
			}
		}
	}

	// Move installer DLCs from temp to game path
	tempInstallerPath := filepath.Join(tempGamePath, "__Installer", "DLC")
	entries, err = os.ReadDir(tempInstallerPath)
	if err != nil {
		logger.Logger.Debug("No installer DLCs in temp", zap.String("path", tempInstallerPath))
	} else {
		// Ensure __Installer/DLC exists in game path
		gameInstallerPath := filepath.Join(d.gamePath, "__Installer", "DLC")
		if err := os.MkdirAll(gameInstallerPath, 0o755); err != nil {
			logger.Logger.Error("Failed to create installer DLC directory", zap.String("path", gameInstallerPath), zap.Error(err))
			return fmt.Errorf("failed to create installer DLC directory: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			if isDLCCode(name) && d.selectedDLCs[name] {
				srcPath := filepath.Join(tempInstallerPath, name)
				dstPath := filepath.Join(gameInstallerPath, name)
				logger.Logger.Debug("Moving installer DLC", zap.String("from", srcPath), zap.String("to", dstPath))

				if _, err := os.Stat(dstPath); err == nil {
					if err := os.RemoveAll(dstPath); err != nil {
						logger.Logger.Error("Failed to remove existing installer DLC", zap.String("path", dstPath), zap.Error(err))
						return fmt.Errorf("failed to remove existing installer DLC %s: %w", name, err)
					}
				}

				err := os.Rename(srcPath, dstPath)
				if err != nil {
					logger.Logger.Error("Failed to move installer DLC", zap.String("from", srcPath), zap.String("to", dstPath), zap.Error(err))
					return fmt.Errorf("failed to move installer DLC %s: %w", name, err)
				}
				logger.Logger.Debug("Successfully moved installer DLC", zap.String("path", dstPath))
			}
		}
	}

	logger.Logger.Debug("Successfully moved all DLCs")
	return nil
}

func (d *TorrentDownloader) DeleteTempDir() error {
	tempGamePath := filepath.Join(os.TempDir(), "The Sims 4")
	logger.Logger.Debug("Removing temp directory", zap.String("path", tempGamePath))

	err := os.RemoveAll(tempGamePath)
	if err != nil {
		// Best-effort cleanup: DLCs were already moved, so just log the warning
		logger.Logger.Warn("Failed to remove temp directory (will be cleaned up later)", zap.String("path", tempGamePath), zap.Error(err))
		return nil
	}

	logger.Logger.Debug("Successfully removed temp directory")
	return nil
}

func isDLCCode(name string) bool {
	return strings.HasPrefix(name, "EP") ||
		strings.HasPrefix(name, "GP") ||
		strings.HasPrefix(name, "SP") ||
		strings.HasPrefix(name, "FP")
}

func (d *TorrentDownloader) IsComplete() bool {
	if d.torrent == nil || len(d.selectedFiles) == 0 {
		return false
	}
	for _, file := range d.selectedFiles {
		if file.BytesCompleted() < file.Length() {
			return false
		}
	}
	return true
}

func (d *TorrentDownloader) GetProgress() application.DownloadProgress {
	if d.torrent == nil {
		return application.DownloadProgress{}
	}

	var bytesDownloaded int64 = 0
	for _, file := range d.selectedFiles {
		bytesDownloaded += file.BytesCompleted()
	}

	totalBytes := d.selectedSize
	if totalBytes == 0 {
		totalBytes = d.torrent.Length()
	}

	now := time.Now()
	timeDiff := now.Sub(d.lastTime).Seconds()

	if timeDiff >= 1.0 {
		bytesDiff := bytesDownloaded - d.lastBytes
		d.speed = int64(float64(bytesDiff) / timeDiff)
		d.lastBytes = bytesDownloaded
		d.lastTime = now
	}

	return application.DownloadProgress{
		BytesDownloaded: bytesDownloaded,
		TotalBytes:      totalBytes,
		Speed:           d.speed,
	}
}

func (d *TorrentDownloader) Stop() error {
	logger.Logger.Debug("Stopping downloader")
	if d.torrent != nil {
		d.torrent.Drop()
		d.torrent = nil
	}
	if d.client != nil {
		d.client.Close()
	}
	return nil
}
