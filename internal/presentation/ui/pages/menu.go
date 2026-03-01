package pages

import (
	"strings"

	"stui/internal/presentation/ui/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuAction int

const (
	MenuActionNone MenuAction = iota
	MenuActionInstallDLC
	MenuActionInstallUnlocker
	MenuActionUninstallUnlocker
	MenuActionInstallUnlockerConfig
)

type MenuPage struct {
	cursor  int
	options []menuOption
	width   int
	height  int
}

type menuOption struct {
	label  string
	action MenuAction
}

const asciiArt = `          _____                _____                    _____                    _____          
         /\    \              /\    \                  /\    \                  /\    \         
        /::\    \            /::\    \                /::\____\                /::\    \        
       /::::\    \           \:::\    \              /:::/    /                \:::\    \       
      /::::::\    \           \:::\    \            /:::/    /                  \:::\    \      
     /:::/\:::\    \           \:::\    \          /:::/    /                    \:::\    \     
    /:::/__\:::\    \           \:::\    \        /:::/    /                      \:::\    \    
    \:::\   \:::\    \          /::::\    \      /:::/    /                       /::::\    \   
  ___\:::\   \:::\    \        /::::::\    \    /:::/    /      _____    ____    /::::::\    \  
 /\   \:::\   \:::\    \      /:::/\:::\    \  /:::/____/      /\    \  /\   \  /:::/\:::\    \ 
/::\   \:::\   \:::\____\    /:::/  \:::\____\|:::|    /      /::\____\/::\   \/:::/  \:::\____\
\:::\   \:::\   \::/    /   /:::/    \::/    /|:::|____\     /:::/    /\:::\  /:::/    \::/    /
 \:::\   \:::\   \/____/   /:::/    / \/____/  \:::\    \   /:::/    /  \:::\/:::/    / \/____/ 
  \:::\   \:::\    \      /:::/    /            \:::\    \ /:::/    /    \::::::/    /          
   \:::\   \:::\____\    /:::/    /              \:::\    /:::/    /      \::::/____/           
    \:::\  /:::/    /    \::/    /                \:::\__/:::/    /        \:::\    \           
     \:::\/:::/    /      \/____/                  \::::::::/    /          \:::\    \          
      \::::::/    /                                 \::::::/    /            \:::\    \         
       \::::/    /                                   \::::/    /              \:::\____\        
        \::/    /                                     \::/____/                \::/    /        
         \/____/                                       ~~                       \/____/`

var (
	menuSelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				PaddingLeft(2)

	menuNormalItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				PaddingLeft(2)

	menuCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	menuHelpTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginTop(1)

	menuBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	asciiArtStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			MarginBottom(2)
)

func NewMenuPage() MenuPage {
	return MenuPage{
		cursor: 0,
		options: []menuOption{
			{"📦 Install DLC", MenuActionInstallDLC},
			{"🔓 Install Unlocker", MenuActionInstallUnlocker},
			{"⚙️ Install Unlocker Config", MenuActionInstallUnlockerConfig},
			{"🗑️ Uninstall Unlocker", MenuActionUninstallUnlocker},
		},
	}
}

func (p MenuPage) Update(msg tea.Msg) (MenuPage, MenuAction) {
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
			if p.cursor < len(p.options)-1 {
				p.cursor++
			}
		case "enter", " ":
			return p, p.options[p.cursor].action
		}
	}
	return p, MenuActionNone
}

func (p MenuPage) View() string {
	var b strings.Builder

	art := asciiArtStyle.Render(asciiArt)
	b.WriteString(art + "\n\n")

	var items []string
	for i, option := range p.options {
		cursor := "  "
		if i == p.cursor {
			cursor = menuCursorStyle.Render("▶ ")
		}

		style := menuNormalItemStyle
		if i == p.cursor {
			style = menuSelectedItemStyle
		}

		items = append(items, cursor+style.Render(option.label))
	}

	menu := menuBoxStyle.Render(strings.Join(items, "\n"))
	b.WriteString(menu + "\n")

	help := menuHelpTextStyle.Render("↑↓ or j/k - navigate | Enter/Space - select | q - quit")
	b.WriteString(help)

	return utils.CenterContent(b.String(), p.width, p.height)
}
