//go:build windows

package ui

import (
	"stui/internal/domain"
	"stui/internal/presentation/ui/pages"

	tea "github.com/charmbracelet/bubbletea"
)

func isClientInfoValid(clientInfo domain.EAClientInfo) bool {
	return clientInfo.Path != ""
}

func (m Model) handleInstallConfig() (Model, tea.Cmd) {
	m.screen = ScreenUnlockerProcess
	clientInfo, err := m.pathFinder.FindEAClient()
	if err != nil {
		m.screen = ScreenMenu
		return m, nil
	}
	return m, m.unlockerProcessPage.Start(m.unlocker, *clientInfo, pages.UnlockerActionInstallConfig)
}
