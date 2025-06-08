package replit

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Prompt struct {
	ti textinput.Model
}

func NewPrompt() *Prompt {
	ti := textinput.New()

	ti.Prompt = "> "
	ti.ShowSuggestions = true
	ti.Focus()

	return &Prompt{
		ti: ti,
	}
}

func (p *Prompt) View() string {
	return p.ti.View()
}

func (p *Prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	newTI, cmd := p.ti.Update(msg)
	p.ti = newTI
	return p, cmd
}

func (p *Prompt) Init() tea.Cmd {
	return nil
}

func (p *Prompt) Value() string {
	return p.ti.Value()
}

func (p *Prompt) SetValue(v string) {
	p.ti.SetValue(v)
}

func (p *Prompt) PromptString() string {
	return p.ti.Prompt
}
