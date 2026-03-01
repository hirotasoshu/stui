package pages

import (
	"fmt"
	"strings"

	"stui/internal/application"
	"stui/internal/domain"
	"stui/internal/presentation/ui/utils"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dlcItem struct {
	dlc       domain.DLC
	selected  bool
	installed bool
}

type DlcSelectionPage struct {
	gamePath       string
	allDlcs        []dlcItem
	filtered       []dlcItem
	cursor         int
	category       domain.DLCType
	categories     []domain.DLCType
	paginator      paginator.Model
	width          int
	height         int
	installedCodes map[string]bool
}

const perPage = 12

var (
	dlcTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	dlcCategoryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				Underline(true).
				MarginTop(1).
				MarginBottom(1)

	dlcSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	dlcInstalledStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)

	dlcCursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	dlcNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	dlcHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			MarginTop(1)

	dlcBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1)

	dlcCheckboxSelected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Render("[✓]")

	dlcCheckboxUnselected = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render("[ ]")
)

func NewDlcSelectionPage() DlcSelectionPage {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = perPage
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("•")

	return DlcSelectionPage{
		paginator:  p,
		categories: []domain.DLCType{domain.TypeExpansionPack, domain.TypeFreePack, domain.TypeGamePack, domain.TypeStuffPack, domain.TypeKit},
		category:   domain.TypeExpansionPack,
	}
}

func (p *DlcSelectionPage) LoadDlcs(repo application.DlcRepository, gamePath string) {
	p.gamePath = gamePath
	p.cursor = 0

	installedEPs := repo.GetInstalledExpansionPacks(gamePath)
	installedFPs := repo.GetInstalledFreePacks(gamePath)
	installedGPs := repo.GetInstalledGamePacks(gamePath)
	installedSPs := repo.GetInstalledStuffPacks(gamePath)
	installedKits := repo.GetInstalledKits(gamePath)

	p.installedCodes = make(map[string]bool)
	for _, dlc := range append(append(append(append(installedEPs, installedFPs...), installedGPs...), installedSPs...), installedKits...) {
		p.installedCodes[dlc.Code] = true
	}

	allDlcs := []domain.DLC{}
	allDlcs = append(allDlcs, repo.GetExpansionPacks()...)
	allDlcs = append(allDlcs, repo.GetFreePacks()...)
	allDlcs = append(allDlcs, repo.GetGamePacks()...)
	allDlcs = append(allDlcs, repo.GetStuffPacks()...)
	allDlcs = append(allDlcs, repo.GetKits()...)

	p.allDlcs = make([]dlcItem, len(allDlcs))
	for i, dlc := range allDlcs {
		installed := p.installedCodes[dlc.Code]
		p.allDlcs[i] = dlcItem{
			dlc:       dlc,
			selected:  installed,
			installed: installed,
		}
	}

	p.applyFilter()
}

func (p *DlcSelectionPage) applyFilter() {
	p.filtered = []dlcItem{}
	for _, item := range p.allDlcs {
		if item.dlc.Type == p.category {
			p.filtered = append(p.filtered, item)
		}
	}

	total := len(p.filtered) + 1
	p.paginator.SetTotalPages(total)

	if p.cursor >= total {
		p.cursor = 0
	}
}

func (p DlcSelectionPage) GetGamePath() string {
	return p.gamePath
}

func (p DlcSelectionPage) Update(msg tea.Msg) (DlcSelectionPage, []domain.DLC, bool, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			idx := -1
			for i, cat := range p.categories {
				if cat == p.category {
					idx = i
					break
				}
			}
			p.category = p.categories[(idx+1)%len(p.categories)]
			p.cursor = 0
			p.paginator.Page = 0
			p.applyFilter()

		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
				start, _ := p.paginator.GetSliceBounds(len(p.filtered) + 1)
				if p.cursor < start && p.paginator.Page > 0 {
					p.paginator.PrevPage()
				}
			}

		case "down", "j":
			total := len(p.filtered) + 1
			if p.cursor < total-1 {
				p.cursor++
				_, end := p.paginator.GetSliceBounds(len(p.filtered) + 1)
				if p.cursor >= end && p.paginator.Page < p.paginator.TotalPages-1 {
					p.paginator.NextPage()
				}
			}

		case " ":
			if p.cursor == 0 {
				p.toggleSelectAll()
			} else {
				idx := p.cursor - 1
				if idx >= 0 && idx < len(p.filtered) {
					realIdx := p.findRealIndex(p.filtered[idx].dlc.Code)
					if realIdx >= 0 && !p.allDlcs[realIdx].installed {
						p.allDlcs[realIdx].selected = !p.allDlcs[realIdx].selected
						p.applyFilter()
					}
				}
			}

		case "enter":
			selectedDlcs := []domain.DLC{}
			for _, item := range p.allDlcs {
				if item.selected && !item.installed {
					selectedDlcs = append(selectedDlcs, item.dlc)
				}
			}
			return p, selectedDlcs, true, nil

		case "esc":
			return p, nil, true, nil
		}
	}

	p.paginator, cmd = p.paginator.Update(msg)
	cmds = append(cmds, cmd)

	return p, nil, false, tea.Batch(cmds...)
}

func (p *DlcSelectionPage) findRealIndex(code string) int {
	for i, item := range p.allDlcs {
		if item.dlc.Code == code {
			return i
		}
	}
	return -1
}

func (p *DlcSelectionPage) toggleSelectAll() {
	allSelected := p.isAllSelected()
	for i := range p.allDlcs {
		if p.allDlcs[i].dlc.Type == p.category && !p.allDlcs[i].installed {
			p.allDlcs[i].selected = !allSelected
		}
	}
	p.applyFilter()
}

func (p DlcSelectionPage) isAllSelected() bool {
	for _, item := range p.filtered {
		if !item.installed && !item.selected {
			return false
		}
	}
	return true
}

func (p DlcSelectionPage) View() string {
	var b strings.Builder

	title := dlcTitleStyle.Render("📦 Select DLCs to Install")
	b.WriteString(title + "\n\n")

	categoryInfo := dlcCategoryStyle.Render(fmt.Sprintf("Category: %s", p.category))
	b.WriteString(categoryInfo + "\n")

	start, end := p.paginator.GetSliceBounds(len(p.filtered) + 1)

	var items []string
	for i := start; i < end; i++ {
		cursor := "  "
		if i == p.cursor {
			cursor = dlcCursorStyle.Render("▶ ")
		} else {
			cursor = "  "
		}

		if i == 0 {
			checkbox := dlcCheckboxUnselected
			if p.isAllSelected() {
				checkbox = dlcCheckboxSelected
			}

			label := "Select All"
			if i == p.cursor {
				label = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170")).Render(label)
			}

			items = append(items, cursor+checkbox+" "+label)
			continue
		}

		idx := i - 1
		if idx >= len(p.filtered) {
			break
		}

		item := p.filtered[idx]

		checkbox := dlcCheckboxUnselected
		if item.selected {
			checkbox = dlcCheckboxSelected
		}

		codeStyle := dlcNormalStyle
		if i == p.cursor {
			codeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
		}

		line := fmt.Sprintf("%s - %s", item.dlc.Code, item.dlc.Name)

		if item.installed {
			line = dlcInstalledStyle.Render(line + " (installed)")
		} else {
			line = codeStyle.Render(line)
		}

		items = append(items, cursor+checkbox+" "+line)
	}

	menu := dlcBoxStyle.Render(strings.Join(items, "\n"))
	b.WriteString(menu + "\n")

	b.WriteString(p.paginator.View() + "\n")

	help := dlcHelpStyle.Render("TAB - switch category | ↑↓/j/k - navigate | SPACE - select | ENTER - download | ESC - back")
	b.WriteString(help)

	return utils.CenterContent(b.String(), p.width, p.height)
}
