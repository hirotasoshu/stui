//go:build linux

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
	textInput    textinput.Model
	pathFinder   application.IPathFinder
	width        int
	height       int
	prefixSource domain.WinePrefixSource
	cursor       int
}

var (
	unlockerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205"))

	unlockerHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1)

	unlockerPrefixTypeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				MarginTop(1)

	unlockerSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)
)

func NewUnlockerPathInputPage(pathFinder application.IPathFinder) UnlockerPathInputPage {
	ti := textinput.New()
	ti.Placeholder = "Enter Wine prefix path..."
	ti.Focus()
	ti.Width = 50

	return UnlockerPathInputPage{
		textInput:    ti,
		pathFinder:   pathFinder,
		prefixSource: domain.WinePrefixSourceWine,
		cursor:       0,
	}
}

func (p *UnlockerPathInputPage) Init() tea.Cmd {
	clientInfo, err := p.pathFinder.FindEAClient()
	if err == nil && clientInfo != nil {
		p.textInput.SetValue(clientInfo.WinePrefix)
		p.prefixSource = clientInfo.PrefixSource
	}
	return textinput.Blink
}

func (p *UnlockerPathInputPage) Reset() {
	p.textInput.SetValue("")
	p.prefixSource = domain.WinePrefixSourceWine
	p.cursor = 0
}

func (p UnlockerPathInputPage) GetClientInfo() domain.EAClientInfo {
	return domain.EAClientInfo{
		WinePrefix:   p.textInput.Value(),
		PrefixSource: p.prefixSource,
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

		case "left", "h":
			if p.cursor == 1 {
				p.cyclePrefixSourceBackward()
			}

		case "right", "l":
			if p.cursor == 1 {
				p.cyclePrefixSourceForward()
			}

		case "enter":
			prefixPath := p.textInput.Value()
			if prefixPath != "" {
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

func (p *UnlockerPathInputPage) cyclePrefixSourceForward() {
	switch p.prefixSource {
	case domain.WinePrefixSourceWine:
		p.prefixSource = domain.WinePrefixSourceSteam
	case domain.WinePrefixSourceSteam:
		p.prefixSource = domain.WinePrefixSourceLutris
	case domain.WinePrefixSourceLutris:
		p.prefixSource = domain.WinePrefixSourceBottles
	case domain.WinePrefixSourceBottles:
		p.prefixSource = domain.WinePrefixSourceWine
	}
}

func (p *UnlockerPathInputPage) cyclePrefixSourceBackward() {
	switch p.prefixSource {
	case domain.WinePrefixSourceWine:
		p.prefixSource = domain.WinePrefixSourceBottles
	case domain.WinePrefixSourceSteam:
		p.prefixSource = domain.WinePrefixSourceWine
	case domain.WinePrefixSourceLutris:
		p.prefixSource = domain.WinePrefixSourceSteam
	case domain.WinePrefixSourceBottles:
		p.prefixSource = domain.WinePrefixSourceLutris
	}
}

func (p UnlockerPathInputPage) View() string {
	title := unlockerTitleStyle.Render("🔓 Wine Prefix Path")

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

	prefixTypeLabel := "Prefix Type: "
	var sourceText string

	switch p.prefixSource {
	case domain.WinePrefixSourceWine:
		sourceText = "Wine"
	case domain.WinePrefixSourceSteam:
		sourceText = "Steam (Proton)"
	case domain.WinePrefixSourceLutris:
		sourceText = "Lutris"
	case domain.WinePrefixSourceBottles:
		sourceText = "Bottles"
	}

	var prefixTypeSection string
	if p.cursor == 1 {
		sourceText = unlockerSelectedStyle.Render(fmt.Sprintf("< %s >", sourceText))
		prefixTypeSection = unlockerPrefixTypeStyle.Render(fmt.Sprintf("%s%s", prefixTypeLabel, sourceText))
	} else {
		prefixTypeSection = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%s%s", prefixTypeLabel, sourceText))
	}

	help := unlockerHelpStyle.Render("↑↓/j/k - navigate | ←→/h/l - cycle type | ENTER - confirm | ESC - back")

	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", title, pathSection, prefixTypeSection, help)
	return utils.CenterContent(content, p.width, p.height)
}
