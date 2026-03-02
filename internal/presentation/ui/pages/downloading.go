package pages

import (
	"fmt"
	"time"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/presentation/ui/utils"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DownloadingPage struct {
	progressBar progress.Model
	err         error
	downloader  application.DlcDownloader
	Completed   bool
	cleanedUp   bool
	width       int
	height      int
}

const plumbobArt = `
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⠛⣧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⠃⠠⠘⣧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⠃⠠⢁⠂⡘⣦⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢰⠏⢀⠁⢂⠐⠠⠘⣆⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⢀⡟⠠⢀⠘⡀⠘⡀⢃⢻⡄⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⢀⡾⠀⡁⢂⠐⠠⢁⠐⡀⢂⢹⡄⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢀⡼⢀⠡⠐⠠⢈⠐⡀⢂⠐⡀⢂⢳⡀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⡼⠃⠄⢂⠁⢂⠂⡐⢀⠂⡐⢀⠂⠌⣷⡀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⣼⠃⡈⠐⠠⠈⠄⢂⠐⡀⢂⠐⠠⢈⠐⡈⢷⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣰⠃⡐⠠⢁⠂⠡⠈⠄⢂⠐⠠⢈⠐⡀⢂⠐⡘⣧⠀⠀⠀⠀
⠀⠀⠀⣰⠏⢀⠐⡀⠂⠌⠠⠁⠌⡀⠌⠐⡀⢂⠐⡀⠂⠄⡘⣧⠀⠀⠀
⠀⠀⢠⠏⠀⠄⢂⠀⠡⢀⠁⢂⠐⠀⠄⡁⠠⢀⠂⠠⢈⠠⢀⠸⣆⠀⠀
⠀⢠⠏⠠⢈⠐⠠⠨⢤⠂⠈⢄⠄⡁⢂⠠⠁⣀⠪⡆⢀⡂⡄⢂⠹⡄⠀
⢠⠏⢀⠁⠢⠑⢑⠄⢐⠀⡁⡬⠃⢀⢠⠑⠣⢸⠸⢠⢰⠰⠐⡄⠠⠹⡄
⢿⡀⠂⠜⣰⠈⡂⣃⠢⠀⠇⠣⠓⠠⠠⠈⣒⠜⡀⠣⠢⠣⡱⠀⢂⠁⡿
⠘⣇⠐⡀⠉⠢⠕⠉⡀⠄⠠⠀⠄⠂⠂⠠⠍⠐⡁⠄⡀⠄⡁⠐⡈⣴⠇
⠀⠸⣦⢀⠁⢂⠐⠠⠐⡈⠄⠡⠈⠄⡁⠂⠌⡐⢀⠂⡐⠠⢀⠡⣰⠃⠀
⠀⠀⠹⣆⠈⠄⡈⠄⠡⠐⡈⠄⡁⠂⠄⡁⢂⠐⡀⠂⠄⡁⢂⣰⠏⠀⠀
⠀⠀⠀⢹⣆⠐⠠⠈⠄⠡⠐⠠⢀⠁⢂⠐⡀⠂⠄⡁⢂⠐⢠⠏⠀⠀⠀
⠀⠀⠀⠀⢻⡀⠡⠈⠄⠡⢈⠐⠠⠈⠄⠂⠄⡁⢂⠐⠠⢨⡟⠀⠀⠀⠀
⠀⠀⠀⠀⠀⢳⡁⠌⠠⢁⠂⠌⠠⠁⠌⡐⠠⠐⠠⢈⢐⡾⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⢧⡈⡐⠠⠈⠄⠡⢈⠐⠠⠁⠌⡐⢠⡞⠁⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠈⢷⡀⠡⠈⠄⡁⠂⠌⠠⢁⠂⣐⡾⠁⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠈⢷⡀⢁⠂⠄⠡⢈⠐⡀⢂⡼⠁⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠈⢧⠀⠌⠠⢁⠂⡐⠠⣼⠃⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣧⠈⠐⠀⠂⡄⢱⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠹⡆⠡⢈⠐⣰⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢳⡔⢀⢢⡏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢳⣠⡞⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠻⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
	`

var (
	downloadArtStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10"))

	downloadTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("12"))

	downloadErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9"))

	downloadCompleteStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10"))

	downloadHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1)
)

func NewDownloadingPage() DownloadingPage {
	return DownloadingPage{
		progressBar: progress.New(progress.WithDefaultGradient()),
	}
}

func (p *DownloadingPage) StartDownload(downloader application.DlcDownloader, dlcs []domain.DLC, gamePath string) tea.Cmd {
	p.err = nil
	p.downloader = downloader
	p.Completed = false
	p.cleanedUp = false
	return downloadCmd(downloader, dlcs, gamePath)
}

type downloadReadyMsg struct {
	err error
}

func downloadCmd(downloader application.DlcDownloader, dlcs []domain.DLC, gamePath string) tea.Cmd {
	return func() tea.Msg {
		err := downloader.Download(dlcs, gamePath)
		return downloadReadyMsg{err: err}
	}
}

type tickMsg time.Time

type cleanupDoneMsg struct {
	err error
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func cleanupCmd(downloader application.DlcDownloader) tea.Cmd {
	return func() tea.Msg {
		if err := downloader.Stop(); err != nil {
			return cleanupDoneMsg{err: err}
		}
		if err := downloader.MoveDLCs(); err != nil {
			return cleanupDoneMsg{err: err}
		}
		if err := downloader.DeleteTempDir(); err != nil {
			return cleanupDoneMsg{err: err}
		}
		return cleanupDoneMsg{}
	}
}

func (p DownloadingPage) Update(msg tea.Msg) (DownloadingPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.progressBar.Width = max(0, msg.Width-4)
	case downloadReadyMsg:
		if msg.err != nil {
			p.err = msg.err
			return p, nil
		}
		return p, tickCmd()
	case tickMsg:
		if p.downloader != nil && !p.Completed {
			if p.downloader.IsComplete() {
				p.Completed = true
				return p, cleanupCmd(p.downloader)
			}
		}
		return p, tickCmd()
	case cleanupDoneMsg:
		if msg.err != nil {
			p.err = msg.err
		} else {
			p.cleanedUp = true
		}
		return p, nil
	case tea.KeyMsg:
		if p.Completed && msg.String() == "enter" {
			return p, nil
		}
	}
	return p, nil
}

func (p DownloadingPage) GetDownloader() application.DlcDownloader {
	return p.downloader
}

func (p DownloadingPage) View() string {
	if p.err != nil {
		content := downloadErrorStyle.Render(fmt.Sprintf("Error: %v\n\nPress q to quit", p.err))
		return utils.CenterContent(content, p.width, p.height)
	}

	if p.downloader == nil {
		content := "Initializing download...\n\nPress q to cancel"
		return utils.CenterContent(content, p.width, p.height)
	}

	prog := p.downloader.GetProgress()

	if p.Completed {
		title := downloadCompleteStyle.Render("✓ Download completed!")

		message := "All DLCs have been downloaded successfully."
		if p.cleanedUp {
			message += "\nTemp directory has been cleaned up."
		} else {
			message += "\nMoving files and cleaning up temp directory..."
		}

		help := downloadHelpStyle.Render("\nPress ENTER to return to main menu")
		content := fmt.Sprintf("%s\n\n%s%s", title, message, help)
		return utils.CenterContent(content, p.width, p.height)
	}

	art := downloadArtStyle.Render(plumbobArt)

	percent := 0.0
	if prog.TotalBytes > 0 {
		percent = float64(prog.BytesDownloaded) / float64(prog.TotalBytes)
	}

	title := downloadTitleStyle.Render("Downloading DLCs...")
	progressBar := p.progressBar.ViewAs(percent)
	downloaded := fmt.Sprintf("Downloaded: %.2f MB / %.2f MB",
		float64(prog.BytesDownloaded)/(1024*1024),
		float64(prog.TotalBytes)/(1024*1024))
	speed := fmt.Sprintf("Speed: %.2f MB/s", float64(prog.Speed)/(1024*1024))
	help := downloadHelpStyle.Render("\nPress q to cancel")

	content := fmt.Sprintf("%s\n%s\n%s\n\n%s\n%s%s", art, title, progressBar, downloaded, speed, help)
	return utils.CenterContent(content, p.width, p.height)
}
