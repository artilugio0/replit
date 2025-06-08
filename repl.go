package replit

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type REPL struct {
	evaluator Evaluator

	prompt *Prompt
	vp     *Viewport

	readOnlyMode bool

	commandRunning       bool
	commandRunningPrompt string

	activeView tea.Model
}

func NewREPL(e Evaluator) *REPL {

	return &REPL{
		evaluator: e,

		prompt: NewPrompt(),
		vp:     NewViewport(15, 20),
	}
}

func (r *REPL) Init() tea.Cmd {
	return nil
}

func (r *REPL) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if r.activeView != nil {
		if msg, ok := msg.(ExitView); ok {
			r.activeView = nil

			s := r.commandRunningPrompt
			if msg.Error != nil {
				s += fmt.Sprintf("\nError: %s", msg.Error.Error())
			} else if msg.Output != "" {
				s += "\n" + msg.Output
			}
			block := StringBlock{s}

			r.vp.AppendBlock(block)
			r.vp.GotoBottom()

			return r, nil
		}

		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "ctrl+c" {
			r.activeView = nil

			s := r.commandRunningPrompt
			block := StringBlock{s}

			r.vp.AppendBlock(block)
			r.vp.GotoBottom()

			return r, nil
		}

		activeView, cmd := r.activeView.Update(msg)
		r.activeView = activeView
		return r, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keys := msg.String()

		if r.readOnlyMode {
			if keys == "l" {
				r.vp.SetSize(min(50, r.vp.GetWidth()+5), r.vp.GetHeight())
				return r, nil
			}
			if keys == "h" {
				r.vp.SetSize(max(15, r.vp.GetWidth()-5), r.vp.GetHeight())
				return r, nil
			}

			if keys == "q" || keys == "ctrl+c" || keys == "esc" {
				r.vp.GotoBottom()
				r.readOnlyMode = false
				r.vp.GotoBottom()
				return r, nil
			}

			r.vp.Update(msg)
			return r, nil
		}

		switch keys {
		case "ctrl+d":
			return r, tea.Quit

		case "enter":
			if r.commandRunning {
				return r, nil
			}

			input := r.prompt.Value()
			if input == "" {
				return r, nil
			}
			r.prompt.SetValue("")

			r.commandRunningPrompt = r.prompt.PromptString() + input

			cmd := func() tea.Msg {
				result, err := r.evaluator.Eval(input)
				if err != nil {
					return msgEvalErr{err}
				}

				return msgEvalOk{result}
			}

			r.commandRunning = true

			return r, cmd

		case "esc":
			r.readOnlyMode = true

		default:
			if !r.commandRunning {
				r.prompt.Update(msg)
			}
		}

	case msgEvalErr:
		r.commandRunning = false
		block := StringBlock{fmt.Sprintf("%s\nError: %s", r.commandRunningPrompt, msg.err.Error())}
		r.vp.AppendBlock(block)
		if !r.readOnlyMode {
			r.vp.GotoBottom()
		}

		return r, nil

	case msgEvalOk:
		r.commandRunning = false

		if msg.result == nil {
			return r, nil
		}

		if msg.result.View == nil {
			block := StringBlock{fmt.Sprintf("%s\n%s", r.commandRunningPrompt, msg.result.Output)}
			r.vp.AppendBlock(block)
			if !r.readOnlyMode {
				r.vp.GotoBottom()
			}

			return r, nil
		}

		r.activeView = msg.result.View
		cmd := r.activeView.Init()
		return r, cmd
	}

	return r, nil
}

func (r *REPL) View() string {
	if r.activeView != nil {
		return r.activeView.View()
	}

	result := r.vp.View()
	result += "\n"

	if r.readOnlyMode {
		result += fmt.Sprintf("-- %d%% --", int(r.vp.ScrollPercent()*100))
	} else if r.commandRunning {
		result += "running..."
	} else {
		result += r.prompt.View()
	}

	return result
}

type Evaluator interface {
	Eval(string) (*Result, error)
}

type Result struct {
	Output string
	View   tea.Model
}

type msgEvalErr struct {
	err error
}

type msgEvalOk struct {
	result *Result
}

type ExitView struct {
	Output string
	Error  error
}
