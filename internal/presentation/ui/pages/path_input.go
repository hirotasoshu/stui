package pages

import (
	"fmt"

	"stui/internal/application"
	"stui/internal/presentation/ui/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PathInputPage struct {
	textInput  textinput.Model
	pathFinder application.IPathFinder
	width      int
	height     int
}

func NewPathInputPage(pathFinder application.IPathFinder) PathInputPage {
	ti := textinput.New()
	ti.Placeholder = "Enter game path..."
	ti.CharLimit = 256
	ti.Width = 50

	return PathInputPage{
		textInput:  ti,
		pathFinder: pathFinder,
	}
}

func (p *PathInputPage) Init() tea.Cmd {
	if p.pathFinder != nil {
		if autoPath := p.pathFinder.FindGamePath(); autoPath != "" {
			p.textInput.SetValue(autoPath)
		}
	}
	p.textInput.Focus()
	return textinput.Blink
}

func (p *PathInputPage) Reset() {
	p.textInput.SetValue("")
}

func (p PathInputPage) Update(msg tea.Msg) (PathInputPage, string, bool, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return p, p.textInput.Value(), true, nil
		case "esc":
			return p, "", true, nil
		}
	}

	p.textInput, cmd = p.textInput.Update(msg)
	return p, "", false, cmd
}

func (p PathInputPage) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginTop(1)

	title := titleStyle.Render("Enter game path:")
	help := helpStyle.Render("(Press ENTER to continue, ESC to go back)")

	content := fmt.Sprintf("%s\n\n%s\n\n%s", title, p.textInput.View(), help)
	return utils.CenterContent(content, p.width, p.height)
}
