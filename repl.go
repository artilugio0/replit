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

	width  int
	height int

	activeView tea.Model
}

func NewREPL(e Evaluator, opts ...REPLOption) *REPL {
	r := &REPL{
		evaluator: e,

		prompt: NewPrompt(),
		vp:     NewViewport(ShowEmptyLines(false)),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
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
			if keys == "q" || keys == "ctrl+c" || keys == "esc" {
				r.readOnlyMode = false
				r.vp.EnableNormalMode()
				r.vp.DisableBlockHighlight()
				r.vp.GotoBottom()
				return r, nil
			}

			r.vp.Update(msg)
			return r, nil
		}

		switch keys {
		case "ctrl+d":
			return r, tea.Quit

		case "esc":
			r.readOnlyMode = true
			r.vp.EnableBlockHighlight()

		default:
			if !r.commandRunning {
				np, cmd := r.prompt.Update(msg)
				r.prompt = np.(*Prompt)
				return r, cmd
			}
		}

	case inputRequest:
		if r.commandRunning {
			return r, nil
		}

		input := msg.input
		if input == "clear" {
			r.Clear()
			return r, nil
		}

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

	case msgEvalErr:
		r.commandRunning = false
		block := StringBlock{fmt.Sprintf("%s\nError: %s", r.commandRunningPrompt, msg.err.Error())}
		r.vp.AppendBlock(block)
		if !r.readOnlyMode {
			r.vp.DisableBlockHighlight()
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
				r.vp.DisableBlockHighlight()
				r.vp.GotoBottom()
			}

			return r, nil
		}

		r.activeView = msg.result.View
		cmd := r.activeView.Init()
		return r, cmd

	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height

		vpHeight := max(0, msg.Height-1)
		vpWidth := max(0, msg.Width)

		r.vp.SetSize(vpWidth, vpHeight)
		r.vp.GotoBottom()
		return r, nil
	}

	return r, nil
}

func (r *REPL) View() string {
	if r.activeView != nil {
		return r.activeView.View()
	}

	result := r.vp.View()
	if result != "" {
		result += "\n"
	}

	if r.readOnlyMode {
		result += fmt.Sprintf("-- %d%% --", int(r.vp.ScrollPercent()*100))
	} else if r.commandRunning {
		result += "running..."
	} else {
		result += r.prompt.View()
	}

	return result
}

func (r *REPL) Clear() {
	r.vp.Clear()
}

func (r *REPL) GetWidth() int {
	return r.width
}

func (r *REPL) GetHeight() int {
	return r.height
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

type inputRequest struct {
	input string
}

type ExitView struct {
	Output string
	Error  error
}

type REPLOption func(*REPL)

func WithPromptInitialSuggestions(s []string) REPLOption {
	return func(r *REPL) {
		WithInitialSuggestions(s)(r.prompt)
	}
}
