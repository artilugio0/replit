package main

import (
	"fmt"
	"os"

	"github.com/artilugio0/replit"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	app := NewApp(20, 10)
	app.AppendBlock(replit.StringBlock{"line A"})
	app.AppendBlock(replit.StringBlock{lipgloss.NewStyle().Background(lipgloss.Color("#0000FF")).Render(`Block X line 1
Block X line 2
Block X line 3
Block X line 4
Block X line 5`)})

	app.AppendFoldedBlock(replit.StringBlock{lipgloss.NewStyle().Background(lipgloss.Color("#FF00FF")).Render(`Block X line 1
Block X line 2
Block X line 3
Block X line 4
Block X line 5
Block X line 6
Block X line 7`)})
	app.AppendBlock(replit.StringBlock{lipgloss.NewStyle().Background(lipgloss.Color("#FF0000")).Render(`line C and
line D
line E`)})
	app.AppendBlock(replit.StringBlock{"linea muy muy muy mumuy muy muy muyy larga\njunto con otra linea larga"})
	app.AppendBlock(replit.StringBlock{"fin"})

	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// App

type App struct {
	vp *replit.Viewport
}

func NewApp(w, h int) App {
	return App{
		vp: replit.NewViewport(w, h),
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return a, tea.Quit

		case "l":
			max := 50
			a.vp.SetSize(min(a.vp.GetWidth()+5, max), a.vp.GetHeight())
			return a, nil

		case "h":
			min := 10
			a.vp.SetSize(max(min, a.vp.GetWidth()-5), a.vp.GetHeight())
			return a, nil

		case "3":
			a.vp.GotoLine(300)
			return a, nil

		default:
			a.vp.Update(msg)
		}
	}

	return a, nil
}

func (a App) View() string {
	border := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

	result := border.Render(a.vp.View())
	result += fmt.Sprintf("\nscroll: %d%%", int(a.vp.ScrollPercent()*100))
	result += fmt.Sprintf("\ncurrent line: %d\n", a.vp.GetCurrentLine())

	return result
}

func (a *App) AppendBlock(b replit.Block) {
	a.vp.AppendBlock(b)
}

func (a *App) AppendFoldedBlock(b replit.Block) {
	a.vp.AppendFoldedBlock(b)
}
