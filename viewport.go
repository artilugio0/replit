package replit

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewportMode int

const (
	viewportModeNormal viewportMode = iota
	viewportModeSearch
	viewportModeSearchNavigation
)

type StringBlock struct {
	S string
}

func (s StringBlock) String() string {
	return s.S
}

func (s StringBlock) FoldedString() string {
	lines := strings.Split(s.S, "\n")
	if len(lines) <= 4 {
		return s.S
	}

	return strings.Join(lines[:min(3, len(lines))], "\n") + "\n..."
}

type Block interface {
	fmt.Stringer

	FoldedString() string
}

type blockState struct {
	block Block

	start int
	end   int

	folded bool
}

type Viewport struct {
	blocks []blockState

	lines []string

	width  int
	height int

	currentLine int
	totalLines  int

	style             lipgloss.Style
	currentBlockStyle lipgloss.Style

	highlightCurrentBlock bool

	showEmptyLines bool

	currentMode viewportMode

	searchInput         textinput.Model
	searchResults       []int
	currentSearchResult int
}

func NewViewport(viewportOptions ...ViewportOption) *Viewport {
	vp := &Viewport{
		blocks: []blockState{},
		lines:  []string{},

		width:  0,
		height: 0,

		currentLine: 0,
		totalLines:  0,

		style: lipgloss.NewStyle().
			Width(0).
			MaxHeight(0),

		currentBlockStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("#3F3F3F")).
			Width(0),

		highlightCurrentBlock: false,

		currentMode: viewportModeNormal,
	}

	for _, opt := range viewportOptions {
		opt(vp)
	}

	return vp
}

func (v *Viewport) AppendBlock(b Block) int {
	s := v.widthAdjustedString(b.String())
	blockHeight := lipgloss.Height(s)

	v.blocks = append(v.blocks, blockState{
		block:  b,
		start:  v.totalLines,
		end:    v.totalLines + blockHeight - 1,
		folded: false,
	})

	v.totalLines += blockHeight

	v.lines = append(v.lines, strings.Split(s, "\n")...)

	if v.totalLines != len(v.lines) {
		panic(fmt.Sprintf("totalLines != len(lines): %d != %d", v.totalLines, len(v.lines)))
	}

	return len(v.blocks) - 1
}

func (v *Viewport) AppendFoldedBlock(b Block) int {
	s := v.widthAdjustedString(b.FoldedString())
	blockHeight := lipgloss.Height(s)

	v.blocks = append(v.blocks, blockState{
		block:  b,
		start:  v.totalLines,
		end:    v.totalLines + blockHeight - 1,
		folded: true,
	})

	v.totalLines += blockHeight

	v.lines = append(v.lines, strings.Split(s, "\n")...)

	if v.totalLines != len(v.lines) {
		panic(fmt.Sprintf("totalLines != len(lines): %d != %d", v.totalLines, len(v.lines)))
	}

	return len(v.blocks) - 1
}

func (v *Viewport) ScrollDown(lines int) {
	v.currentLine = max(0, min(v.totalLines-1, v.currentLine+lines))
}

func (v *Viewport) ScrollUp(lines int) {
	v.currentLine = max(0, v.currentLine-lines)
}

func (v *Viewport) View() string {
	if v.totalLines == 0 || v.width == 0 || v.availableLines() == 0 {
		return ""
	}

	contentBuilder := strings.Builder{}

	var currentBlock blockState
	if len(v.lines) > 0 {
		currentBlock = v.blocks[v.CurrentBlockIndex()]
	}

	availableLines := v.availableLines()
	currentLine := v.currentLine
	if v.currentMode == viewportModeSearch || v.currentMode == viewportModeSearchNavigation {
		availableLines = max(0, availableLines-1)
	}

	lineCount := 0
	currentBlockContent := ""
	limit := min(currentLine+availableLines, v.totalLines)

	for l := currentLine; l < limit; l++ {
		// Handle highlight of current block
		if l >= currentBlock.start && l < currentBlock.end {
			currentBlockContent += v.lines[l] + "\n"
		} else if l == currentBlock.end {
			currentBlockContent += v.lines[l]
			if v.highlightCurrentBlock {
				contentBuilder.WriteString(v.currentBlockStyle.Render(currentBlockContent))
			} else {
				contentBuilder.WriteString(currentBlockContent)
			}

			if l < limit-1 {
				contentBuilder.WriteRune('\n')
			}
			// End Handle highlight of current block
		} else {
			contentBuilder.WriteString(v.lines[l])

			if l < limit-1 {
				contentBuilder.WriteRune('\n')
			}
		}

		lineCount++
	}

	// Handle highlight of current block
	if lineCount > 0 && limit <= currentBlock.end {
		if v.highlightCurrentBlock {
			contentBuilder.WriteString(v.currentBlockStyle.Render(currentBlockContent))
		} else {
			contentBuilder.WriteString(currentBlockContent)
		}
	}
	// End Handle highlight of current block

	if v.showEmptyLines {
		for ; lineCount < availableLines; lineCount++ {
			contentBuilder.WriteRune('\n')
		}
	}

	content := contentBuilder.String()

	if v.currentMode == viewportModeSearch {
		content += v.searchInput.View()
	} else if v.currentMode == viewportModeSearchNavigation {
		if len(v.searchResults) > 0 {
			content += fmt.Sprintf("%d / %d matches", v.currentSearchResult+1, len(v.searchResults))
		} else {
			content += "no matches found"
		}
	}

	return v.style.Render(content)
}

func (v *Viewport) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if v.currentMode == viewportModeSearch {
		if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
			v.Search(v.searchInput.Value())
			return v, nil
		}

		si, cmd := v.searchInput.Update(msg)
		v.searchInput = si
		return v, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "k":
			v.ScrollUp(1)

		case "j":
			v.ScrollDown(1)

		case "u":
			v.ScrollUp(v.GetHeight() / 2)

		case "d":
			v.ScrollDown(v.GetHeight() / 2)

		case "b":
			v.ScrollUp(v.GetHeight())

		case "f":
			v.ScrollDown(v.GetHeight())

		case "tab":
			v.ToggleCurrentBlock()

		case "g":
			v.GotoTop()

		case "G":
			v.GotoBottom()

		case "n":
			if v.currentMode == viewportModeSearchNavigation {
				v.GotoNextSearchResult()
			}

		case "N":
			if v.currentMode == viewportModeSearchNavigation {
				v.GotoPreviousSearchResult()
			}

		case "/":
			if v.currentMode == viewportModeNormal || v.currentMode == viewportModeSearchNavigation {
				v.EnableSearchMode()
			}

		case "ctrl+c":
			if v.currentMode != viewportModeNormal {
				v.EnableNormalMode()
			}
		}
	}

	return v, nil
}

func (v *Viewport) Init() tea.Cmd {
	return nil
}

func (v *Viewport) ToggleCurrentBlock() {
	i := v.CurrentBlockIndex()

	var newContent string
	newFolded := false
	if v.blocks[i].folded {
		newContent = v.widthAdjustedString(v.blocks[i].block.String())
	} else {
		newContent = v.widthAdjustedString(v.blocks[i].block.FoldedString())
		newFolded = true
	}

	newContentLines := strings.Split(newContent, "\n")

	linesPre := v.lines[:v.blocks[i].start]
	linesPost := append([]string{}, v.lines[v.blocks[i].end+1:]...)
	v.lines = append(linesPre, newContentLines...)
	v.lines = append(v.lines, linesPost...)

	lengthDiff := v.blocks[i].end - v.blocks[i].start + 1 - len(newContentLines)
	v.blocks[i].end -= lengthDiff
	v.blocks[i].folded = newFolded

	for i = i + 1; i < len(v.blocks); i++ {
		v.blocks[i].start -= lengthDiff
		v.blocks[i].end -= lengthDiff
	}

	v.totalLines -= lengthDiff
	v.currentLine = max(0, min(v.currentLine, v.totalLines-1))
}

func (v *Viewport) CurrentBlockIndex() int {
	i := 0

	for ; v.blocks[i].start > v.currentLine || v.blocks[i].end < v.currentLine; i++ {
	}

	return i
}

func (v *Viewport) SetSize(width, height int) {
	v.width = width
	v.height = height

	v.style = v.style.
		Width(width - v.style.GetBorderLeftSize() - v.style.GetBorderRightSize()).
		MaxHeight(height)

	v.currentBlockStyle = v.currentBlockStyle.Width(width - v.style.GetBorderLeftSize() - v.style.GetBorderRightSize())

	if v.width == 0 || v.height == 0 {
		return
	}

	v.totalLines = 0
	v.lines = []string{}
	for i, bst := range v.blocks {
		var s string
		if bst.folded {
			s = bst.block.FoldedString()
		} else {
			s = bst.block.String()
		}
		s = v.widthAdjustedString(s)

		blockHeight := lipgloss.Height(s)

		v.blocks[i] = blockState{
			block:  bst.block,
			start:  v.totalLines,
			end:    v.totalLines + blockHeight - 1,
			folded: bst.folded,
		}

		v.totalLines += blockHeight

		v.lines = append(v.lines, strings.Split(s, "\n")...)

		if v.totalLines != len(v.lines) {
			panic(fmt.Sprintf("totalLines != len(lines): %d != %d", v.totalLines, len(v.lines)))
		}
	}

	v.currentLine = max(0, min(v.currentLine, v.totalLines-1))
}

func (v *Viewport) GetWidth() int {
	return v.width
}

func (v *Viewport) GetHeight() int {
	return v.height
}

func (v *Viewport) ScrollPercent() float64 {
	if v.totalLines == 0 {
		return 1.0
	}

	return float64(v.currentLine+1) / float64(v.totalLines)
}

func (v *Viewport) GotoBottom() {
	v.currentLine = max(0, v.totalLines-v.availableLines())
}

func (v *Viewport) GotoTop() {
	v.currentLine = 0
}

func (v *Viewport) GotoLine(line int) {
	v.currentLine = max(0, min(v.totalLines-1, line))
}

func (v *Viewport) GetCurrentLine() int {
	return v.currentLine
}

func (v *Viewport) widthAdjustedString(s string) string {
	return lipgloss.NewStyle().
		Width(v.width - v.style.GetBorderLeftSize() - v.style.GetBorderRightSize()).
		Render(s)
}

func (v *Viewport) EnableBlockHighlight() {
	v.highlightCurrentBlock = true
}

func (v *Viewport) DisableBlockHighlight() {
	v.highlightCurrentBlock = false
}

func (v *Viewport) TotalLines() int {
	return v.totalLines
}

func (v *Viewport) SetStyle(style lipgloss.Style) {
	v.style = style
	v.SetSize(v.width, v.height)
}

func (v *Viewport) availableLines() int {
	return max(0, v.height-v.style.GetBorderTopSize()-v.style.GetBorderBottomSize())
}

func (v *Viewport) Clear() {
	v.blocks = []blockState{}
	v.lines = []string{}
	v.currentLine = 0
	v.totalLines = 0
}

func (v *Viewport) EnableNormalMode() {
	v.currentMode = viewportModeNormal
}

func (v *Viewport) EnableSearchMode() {
	v.searchInput = textinput.New()
	v.searchInput.Prompt = "search: "
	v.searchInput.Focus()

	v.currentMode = viewportModeSearch
}

func (v *Viewport) Search(s string) {
	// TODO: this does not work if a block is folded. Fix it
	result := []int{}

	for i, l := range v.lines {
		if strings.Contains(l, s) {
			result = append(result, i)
		}
	}

	v.searchResults = result
	v.currentMode = viewportModeSearchNavigation
	v.currentSearchResult = -1

	v.GotoNextSearchResult()
}

func (v *Viewport) GotoNextSearchResult() {
	if len(v.searchResults) > 0 {
		v.currentSearchResult = (v.currentSearchResult + 1) % len(v.searchResults)
		v.GotoLine(v.searchResults[v.currentSearchResult])
	}
}

func (v *Viewport) GotoPreviousSearchResult() {
	if len(v.searchResults) > 0 {
		v.currentSearchResult = (len(v.searchResults) + v.currentSearchResult - 1) % len(v.searchResults)
		v.GotoLine(v.searchResults[v.currentSearchResult])
	}
}

type ViewportOption func(*Viewport)

func ShowEmptyLines(yes bool) ViewportOption {
	return func(vp *Viewport) {
		vp.showEmptyLines = yes
	}
}
