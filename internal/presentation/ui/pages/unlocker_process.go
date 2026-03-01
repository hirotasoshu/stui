package pages

import (
	"fmt"
	"time"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/presentation/ui/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UnlockerAction int

const (
	UnlockerActionInstallDLL UnlockerAction = iota
	UnlockerActionUninstallDLL
	UnlockerActionInstallConfig
)

type UnlockerProcessPage struct {
	action   UnlockerAction
	err      error
	complete bool
	width    int
	height   int
}

var (
	unlockerProcessTitleStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("205"))

	unlockerProcessErrorStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("196")).
					Bold(true)

	unlockerProcessSuccessStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("10")).
					Bold(true)

	unlockerProcessHelpStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					MarginTop(1)
)

func NewUnlockerProcessPage() UnlockerProcessPage {
	return UnlockerProcessPage{}
}

type unlockerTickMsg time.Time

func unlockerTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return unlockerTickMsg(t)
	})
}

func (p *UnlockerProcessPage) Start(unlocker application.IUnlocker, clientInfo domain.EAClientInfo, action UnlockerAction) tea.Cmd {
	p.action = action
	p.err = nil
	p.complete = false

	go func() {
		var err error
		switch action {
		case UnlockerActionInstallDLL:
			err = unlocker.InstallDLL(clientInfo)
		case UnlockerActionUninstallDLL:
			err = unlocker.UninstallDLL(clientInfo)
		case UnlockerActionInstallConfig:
			err = unlocker.InstallConfig(clientInfo)
		}
		p.err = err
		p.complete = true
	}()

	return unlockerTickCmd()
}

func (p UnlockerProcessPage) IsComplete() bool {
	return p.complete
}

func (p UnlockerProcessPage) Update(msg tea.Msg) (*UnlockerProcessPage, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height

	case unlockerTickMsg:
		if !p.complete {
			return &p, unlockerTickCmd()
		}

	case tea.KeyMsg:
		if p.complete && msg.String() == "enter" {
			return &p, nil
		}
	}

	if !p.complete {
		return &p, unlockerTickCmd()
	}
	return &p, nil
}

func (p UnlockerProcessPage) View() string {
	if p.err != nil {
		title := unlockerProcessErrorStyle.Render("✗ Error")
		message := fmt.Sprintf("Failed: %v", p.err)
		help := unlockerProcessHelpStyle.Render("\nPress ENTER to return to menu")
		content := fmt.Sprintf("%s\n\n%s%s", title, message, help)
		return utils.CenterContent(content, p.width, p.height)
	}

	if !p.complete {
		var title string
		switch p.action {
		case UnlockerActionInstallDLL:
			title = "Installing DLC Unlocker..."
		case UnlockerActionUninstallDLL:
			title = "Uninstalling DLC Unlocker..."
		case UnlockerActionInstallConfig:
			title = "Installing Unlocker Config..."
		}
		content := unlockerProcessTitleStyle.Render(title)
		return utils.CenterContent(content, p.width, p.height)
	}

	var title, message string
	switch p.action {
	case UnlockerActionInstallDLL:
		title = unlockerProcessSuccessStyle.Render("✓ DLC Unlocker Installed")
		message = "DLC Unlocker has been installed successfully!"
	case UnlockerActionUninstallDLL:
		title = unlockerProcessSuccessStyle.Render("✓ DLC Unlocker Uninstalled")
		message = "DLC Unlocker has been uninstalled successfully!"
	case UnlockerActionInstallConfig:
		title = unlockerProcessSuccessStyle.Render("✓ Config Installed")
		message = "Unlocker config has been installed successfully!"
	}

	help := unlockerProcessHelpStyle.Render("\nPress ENTER to return to menu")
	content := fmt.Sprintf("%s\n\n%s%s", title, message, help)
	return utils.CenterContent(content, p.width, p.height)
}
