//go:build linux

package ui

import (
	"stui/internal/domain"
	"stui/internal/presentation/ui/pages"

	tea "github.com/charmbracelet/bubbletea"
)

func isClientInfoValid(clientInfo domain.EAClientInfo) bool {
	return clientInfo.WinePrefix != ""
}

func (m Model) handleInstallConfig() (Model, tea.Cmd) {
	m.pendingUnlockerAction = pages.UnlockerActionInstallConfig
	m.screen = ScreenUnlockerPathInput
	return m, m.unlockerPathInputPage.Init()
}
