//go:build windows

package pages

import (
	"fmt"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/presentation/ui/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UnlockerPathInputPage struct {
	textInput  textinput.Model
	pathFinder application.IPathFinder
	width      int
	height     int
	clientType domain.EAClientType
	cursor     int
}

var (
	unlockerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))

	unlockerHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1)

	unlockerClientTypeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				MarginTop(1)

	unlockerSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)
)

func NewUnlockerPathInputPage(pathFinder application.IPathFinder) UnlockerPathInputPage {
	ti := textinput.New()
	ti.Placeholder = "Enter EA app/Origin path..."
	ti.Focus()
	ti.Width = 50

	return UnlockerPathInputPage{
		textInput:  ti,
		pathFinder: pathFinder,
		clientType: domain.EAClientTypeEAApp,
		cursor:     0,
	}
}

func (p *UnlockerPathInputPage) Init() tea.Cmd {
	clientInfo, err := p.pathFinder.FindEAClient()
	if err == nil && clientInfo != nil {
		p.textInput.SetValue(clientInfo.Path)
		p.clientType = clientInfo.ClientType
	}
	return textinput.Blink
}

func (p *UnlockerPathInputPage) Reset() {
	p.textInput.SetValue("")
	p.clientType = domain.EAClientTypeEAApp
	p.cursor = 0
}

func (p UnlockerPathInputPage) GetClientInfo() domain.EAClientInfo {
	return domain.EAClientInfo{
		Path:       p.textInput.Value(),
		ClientType: p.clientType,
	}
}

func (p UnlockerPathInputPage) Update(msg tea.Msg) (UnlockerPathInputPage, domain.EAClientInfo, bool, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}

		case "down", "j":
			if p.cursor < 1 {
				p.cursor++
			}

		case " ":
			if p.cursor == 1 {
				if p.clientType == domain.EAClientTypeEAApp {
					p.clientType = domain.EAClientTypeOrigin
				} else {
					p.clientType = domain.EAClientTypeEAApp
				}
			}

		case "enter":
			clientPath := p.textInput.Value()
			if clientPath != "" {
				return p, p.GetClientInfo(), true, nil
			}

		case "esc":
			return p, domain.EAClientInfo{}, true, nil
		}
	}

	if p.cursor == 0 {
		p.textInput, cmd = p.textInput.Update(msg)
	}

	return p, domain.EAClientInfo{}, false, cmd
}

func (p UnlockerPathInputPage) View() string {
	title := unlockerTitleStyle.Render("🔓 EA Client Path")

	var pathSection string
	if p.cursor == 0 {
		pathSection = p.textInput.View()
	} else {
		value := p.textInput.Value()
		if value == "" {
			value = p.textInput.Placeholder
		}
		pathSection = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(value)
	}

	clientTypeLabel := "Client Type: "
	var eaAppOption, originOption string

	if p.clientType == domain.EAClientTypeEAApp {
		eaAppOption = "[✓] EA App"
		originOption = "[ ] Origin"
	} else {
		eaAppOption = "[ ] EA App"
		originOption = "[✓] Origin"
	}

	var clientTypeSection string
	if p.cursor == 1 {
		if p.clientType == domain.EAClientTypeEAApp {
			eaAppOption = unlockerSelectedStyle.Render(eaAppOption)
		} else {
			originOption = unlockerSelectedStyle.Render(originOption)
		}
		clientTypeSection = unlockerClientTypeStyle.Render(fmt.Sprintf("%s%s  %s", clientTypeLabel, eaAppOption, originOption))
	} else {
		clientTypeSection = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%s%s  %s", clientTypeLabel, eaAppOption, originOption))
	}

	help := unlockerHelpStyle.Render("↑↓/j/k - navigate | SPACE - toggle | ENTER - confirm | ESC - back")

	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", title, pathSection, clientTypeSection, help)
	return utils.CenterContent(content, p.width, p.height)
}
