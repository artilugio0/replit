package main

import (
	"fmt"
	"os"
	"time"

	"github.com/artilugio0/replit"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	repl := replit.NewREPL(Evaluator{})

	p := tea.NewProgram(repl)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

type Evaluator struct{}

func (e Evaluator) Eval(input string) (*replit.Result, error) {
	switch input {
	case "hi":
		return &replit.Result{
			Output: "hi! how are you?",
		}, nil

	case "multi":
		return &replit.Result{
			Output: "line1\nline2\nline3\nline4\nline5",
		}, nil

	case "slow":
		time.Sleep(2 * time.Second)
		return &replit.Result{
			Output: "done!",
		}, nil

	case "vvv":
		return &replit.Result{
			View: TestView{},
		}, nil

	default:
		return nil, fmt.Errorf("invalid command '%s'", input)
	}
}

type TestView struct {
	initdone bool
}

func (tv TestView) Init() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return "initdone"
	}
}

func (tv TestView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(string); ok && msg == "initdone" {
		tv.initdone = true
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keys := msg.String()
		if keys == "q" {
			return tv, func() tea.Msg {
				return replit.ExitView{
					Output: "everything OK",
				}
			}
		}

		if keys == "x" {
			return tv, func() tea.Msg {
				return replit.ExitView{
					Error: fmt.Errorf("Something horrible happened"),
				}
			}
		}
	}

	return tv, nil
}

func (tv TestView) View() string {
	if tv.initdone {
		return "Everything is initialized!!"
	}

	return "Initializing..."
}
