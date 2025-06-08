package replit

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Prompt struct {
	ti textinput.Model

	history []string

	reverseSearchMode bool
}

func NewPrompt(opts ...PromptOption) *Prompt {
	ti := textinput.New()

	ti.Prompt = "> "
	ti.ShowSuggestions = true
	ti.Focus()

	p := &Prompt{
		ti:      ti,
		history: []string{},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Prompt) View() string {
	return p.ti.View()
}

func (p *Prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keys := msg.String()
		switch keys {

		case "enter":
			v := p.ti.Value()
			if v == "" {
				return p, nil
			}

			p.history = append(p.history, v)
			p.ti.SetSuggestions(p.history)

			p.ti.SetValue("")
			return p, func() tea.Msg {
				return inputRequest{v}
			}
		}
	}

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

type PromptOption func(*Prompt)

func WithInitialSuggestions(s []string) PromptOption {
	return func(p *Prompt) {
		p.history = append(p.history, s...)
		p.ti.SetSuggestions(p.history)
	}
}
