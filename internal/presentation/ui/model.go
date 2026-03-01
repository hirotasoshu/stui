package ui

import (
	"stui/internal/application"
	"stui/internal/presentation/ui/pages"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenPathInput
	ScreenDlcSelection
	ScreenDownloading
	ScreenUnlockerPathInput
	ScreenUnlockerProcess
)

type Model struct {
	screen            Screen
	repo              application.DlcRepository
	downloaderFactory func() (application.DlcDownloader, error)
	pathFinder        application.IPathFinder
	unlocker          application.IUnlocker

	menuPage              pages.MenuPage
	pathInputPage         pages.PathInputPage
	dlcSelectionPage      pages.DlcSelectionPage
	downloadingPage       pages.DownloadingPage
	unlockerPathInputPage pages.UnlockerPathInputPage
	unlockerProcessPage   *pages.UnlockerProcessPage
	pendingUnlockerAction pages.UnlockerAction

	width  int
	height int
}

func NewModel(repo application.DlcRepository, pathFinder application.IPathFinder, unlocker application.IUnlocker, downloaderFactory func() (application.DlcDownloader, error)) Model {
	unlockerProcPage := pages.NewUnlockerProcessPage()
	return Model{
		screen:                ScreenMenu,
		repo:                  repo,
		downloaderFactory:     downloaderFactory,
		pathFinder:            pathFinder,
		unlocker:              unlocker,
		menuPage:              pages.NewMenuPage(),
		pathInputPage:         pages.NewPathInputPage(pathFinder),
		dlcSelectionPage:      pages.NewDlcSelectionPage(),
		downloadingPage:       pages.NewDownloadingPage(),
		unlockerPathInputPage: pages.NewUnlockerPathInputPage(pathFinder),
		unlockerProcessPage:   &unlockerProcPage,
	}
}

func (m Model) Init() tea.Cmd {
	return m.pathInputPage.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || (msg.String() == "q" && m.screen != ScreenDownloading && m.screen != ScreenUnlockerProcess) {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuPage, _ = m.menuPage.Update(msg)
		m.pathInputPage, _, _, _ = m.pathInputPage.Update(msg)
		m.dlcSelectionPage, _, _, _ = m.dlcSelectionPage.Update(msg)
		m.downloadingPage, _ = m.downloadingPage.Update(msg)
		m.unlockerPathInputPage, _, _, _ = m.unlockerPathInputPage.Update(msg)
		m.unlockerProcessPage, _ = m.unlockerProcessPage.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
	}

	switch m.screen {
	case ScreenMenu:
		page, action := m.menuPage.Update(msg)
		m.menuPage = page
		if action == pages.MenuActionInstallDLC {
			m.screen = ScreenPathInput
			return m, m.pathInputPage.Init()
		} else if action == pages.MenuActionInstallUnlocker {
			m.pendingUnlockerAction = pages.UnlockerActionInstallDLL
			m.screen = ScreenUnlockerPathInput
			return m, m.unlockerPathInputPage.Init()
		} else if action == pages.MenuActionUninstallUnlocker {
			m.pendingUnlockerAction = pages.UnlockerActionUninstallDLL
			m.screen = ScreenUnlockerPathInput
			return m, m.unlockerPathInputPage.Init()
		} else if action == pages.MenuActionInstallUnlockerConfig {
			m, cmd := m.handleInstallConfig()
			return m, cmd
		}
		return m, nil

	case ScreenPathInput:
		page, gamePath, done, cmd := m.pathInputPage.Update(msg)
		m.pathInputPage = page
		if done {
			if gamePath != "" {
				m.dlcSelectionPage.LoadDlcs(m.repo, gamePath)
				m.screen = ScreenDlcSelection
			} else {
				m.screen = ScreenMenu
				m.pathInputPage.Reset()
			}
			return m, nil
		}
		return m, cmd

	case ScreenDlcSelection:
		page, selectedDlcs, done, cmd := m.dlcSelectionPage.Update(msg)
		m.dlcSelectionPage = page
		if done {
			if len(selectedDlcs) > 0 {
				downloader, err := m.downloaderFactory()
				if err != nil {
					m.screen = ScreenMenu
					return m, nil
				}

				m.screen = ScreenDownloading
				m.downloadingPage = pages.NewDownloadingPage()
				m.downloadingPage, _ = m.downloadingPage.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
				cmd := m.downloadingPage.StartDownload(downloader, selectedDlcs, m.dlcSelectionPage.GetGamePath())
				return m, cmd
			} else {
				m.screen = ScreenMenu
			}
			return m, nil
		}
		return m, cmd

	case ScreenDownloading:
		page, cmd := m.downloadingPage.Update(msg)
		m.downloadingPage = page

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if m.downloadingPage.Completed && keyMsg.String() == "enter" {
				if m.downloadingPage.GetDownloader() != nil {
					m.downloadingPage.GetDownloader().Stop()
				}
				m.screen = ScreenMenu
				return m, nil
			}
		}

		return m, cmd

	case ScreenUnlockerPathInput:
		page, clientInfo, done, cmd := m.unlockerPathInputPage.Update(msg)
		m.unlockerPathInputPage = page
		if done {
			if isClientInfoValid(clientInfo) {
				m.screen = ScreenUnlockerProcess
				return m, m.unlockerProcessPage.Start(m.unlocker, clientInfo, m.pendingUnlockerAction)
			} else {
				m.screen = ScreenMenu
			}
			return m, cmd
		}
		return m, cmd

	case ScreenUnlockerProcess:
		page, cmd := m.unlockerProcessPage.Update(msg)
		m.unlockerProcessPage = page

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if m.unlockerProcessPage.IsComplete() && keyMsg.String() == "enter" {
				m.screen = ScreenMenu
				return m, nil
			}
		}

		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case ScreenMenu:
		return m.menuPage.View()
	case ScreenPathInput:
		return m.pathInputPage.View()
	case ScreenDlcSelection:
		return m.dlcSelectionPage.View()
	case ScreenDownloading:
		return m.downloadingPage.View()
	case ScreenUnlockerPathInput:
		return m.unlockerPathInputPage.View()
	case ScreenUnlockerProcess:
		return m.unlockerProcessPage.View()
	}
	return ""
}
