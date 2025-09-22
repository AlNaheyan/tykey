package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"termtype/internal/words"
)

type screen int

const (
	screenMenu screen = iota
	screenInputWords
	screenTyping
	screenResults
)

type AppModel struct {
	scr    screen
	sel    int // 0: timer, 1: words
	width  int
	height int

	// input for words mode
	input string

	// active typing test
	test    TestModel
	hasTest bool

	// last mode for retry
	lastMode  int // 0 timer, 1 words
	lastWords int

	// cached results
	stats Stats
}

func NewApp() AppModel {
	return AppModel{scr: screenMenu}
}

func (m AppModel) Init() tea.Cmd { return nil }

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.hasTest {
			// forward to test for centering
			next, _ := m.test.Update(msg)
			if tm, ok := next.(TestModel); ok {
				m.test = tm
			}
		}
		return m, nil
	case tickMsg:
		if m.scr == screenTyping {
			next, cmd := m.test.Update(msg)
			if tm, ok := next.(TestModel); ok {
				m.test = tm
			}
			if m.test.done {
				m.stats = m.test.Stats()
				m.scr = screenResults
			}
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		k := msg.String()
		switch m.scr {
		case screenMenu:
			switch k {
			case "up", "k":
				if m.sel > 0 {
					m.sel--
				}
			case "down", "j":
				if m.sel < 1 {
					m.sel++
				}
			case "enter":
				if m.sel == 0 {
					// timer mode: generate enough words and start 15s test. change to selected time TODO!!
					tgt, err := words.GenerateString(120)
					if err != nil {
						//  fallback target on error
						tgt = "the quick brown fox jumps over the lazy dog"
					}
					m.test = NewTestModelWith(tgt, 15*time.Second)
					m.hasTest = true
					m.scr = screenTyping
					m.lastMode = 0
				} else {
					m.input = ""
					m.scr = screenInputWords
				}
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil

		case screenInputWords:
			switch msg.Type {
			case tea.KeyBackspace:
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
			case tea.KeyEnter:
				n, _ := strconv.Atoi(m.input)
				if n <= 0 {
					n = 25
				}
				tgt, err := words.GenerateString(n)
				if err != nil {
					tgt = "the quick brown fox jumps over the lazy dog"
				}
				m.test = NewTestModelWith(tgt, 0)
				m.hasTest = true
				m.scr = screenTyping
				m.lastMode = 1
				m.lastWords = n
			default:
				if len(msg.Runes) > 0 {
					r := msg.Runes[0]
					if r >= '0' && r <= '9' {
						if len(m.input) < 4 { 
							m.input += string(r)
						}
					}
				} else if k == "esc" {
					m.scr = screenMenu
				}
			}
			return m, nil

		case screenTyping:
			// forward all events to the test model
			var cmd tea.Cmd
			var next tea.Model
			next, cmd = m.test.Update(msg)
			if tm, ok := next.(TestModel); ok {
				m.test = tm
			}
			// if completed, compute stats and go to results
			if m.test.done {
				m.stats = m.test.Stats()
				m.scr = screenResults
			}
			return m, cmd
		case screenResults:
			switch k {
			case "1":
				if m.lastMode == 0 {
					tgt, err := words.GenerateString(120)
					if err != nil {
						tgt = "the quick brown fox jumps over the lazy dog"
					}
					m.test = NewTestModelWith(tgt, 15*time.Second)
				} else {
					tgt, err := words.GenerateString(m.lastWords)
					if err != nil {
						tgt = "the quick brown fox jumps over the lazy dog"
					}
					m.test = NewTestModelWith(tgt, 0)
				}
				m.hasTest = true
				m.scr = screenTyping
			case "esc":
				return m, tea.Quit
			default:
				// ignore any other keys
				return m, nil
			}
		}
	}
	return m, nil
}

func center(width, height int, content string) string {
	if width <= 0 {
		width = 80
	}
	lines := strings.Split(content, "\n")
	// horizontal center each line
	b := strings.Builder{}
	for i, ln := range lines {
		pad := 0
		if width > 0 {
			pad = (width - visualLen(ln)) / 2
			if pad < 0 {
				pad = 0
			}
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(ln)
	}
	body := b.String()
	top := 0
	if height > 0 {
		top = (height - len(lines)) / 2
		if top < 0 {
			top = 0
		}
	}
	return strings.Repeat("\n", top) + body
}

func (m AppModel) View() string {
	switch m.scr {
	case screenMenu:
		items := []string{"Timer (15s)", "Words (custom)"}
		var b strings.Builder
		// main text colored
		b.WriteString(AccentBold)
		b.WriteString("Select mode:")
		b.WriteString(AnsiReset)
		b.WriteString("\n\n")
		for i, it := range items {
			prefix := "  "
			if i == m.sel {
				prefix = "> "
			}
			if i == m.sel {
				b.WriteString(AccentBold)
				b.WriteString(prefix)
				b.WriteString(it)
				b.WriteString(AnsiReset)
			} else {
				b.WriteString(prefix)
				b.WriteString(it)
			}
			b.WriteByte('\n')
		}
		b.WriteString("\n")
		b.WriteString(AnsiDim)
		b.WriteString("↑/↓ or j/k to move, Enter to select, q to quit")
		b.WriteString(AnsiReset)
		return center(m.width, m.height, b.String())

	case screenInputWords:
		prompt := fmt.Sprintf("%sHow many words?%s %s", AccentBold, AnsiReset, m.input)
		hint := "\nEnter to start, Esc to cancel"
		return center(m.width, m.height, prompt+hint)

	case screenTyping:
		return m.test.View()
	case screenResults:
		s := m.stats
		header := AccentBold + "+------------------------------+\n|          RESULTS             |\n+------------------------------+" + AnsiReset
		body := fmt.Sprintf("%s\nTime: %.1fs\nWPM (gross): %.1f\nAccuracy: %.1f%% (%d/%d)\nWPM (net): %.1f\n\n[1] Restart  [esc] Quit",
			header, s.DurationSec, s.GrossWPM, s.AccuracyPct, s.Correct, s.TotalTyped, s.NetWPM)
		return center(m.width, m.height, body)
	}
	return ""
}
