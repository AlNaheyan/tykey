package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type tickMsg time.Time

type TestModel struct {
	target []rune
	typed  []rune
	cursor int

	started   bool
	startTime time.Time
	endTime   time.Time
	duration  time.Duration
	remaining time.Duration
	done      bool

	// stats
	correct   int
	incorrect int

	// viewport size (for centering)
	width  int
	height int
}

func NewTestModel() TestModel {
	return NewTestModelWith("the quick brown fox jumps over the lazy dog", 15*time.Second)
}

func NewTestModelWith(target string, d time.Duration) TestModel {
	if d < 0 {
		d = 0
	}
	tm := TestModel{
		target:    []rune(target),
		duration:  d,
		remaining: d,
	}
	return tm
}

func (m TestModel) Init() tea.Cmd {
	return nil
}

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		if !m.started || m.done || m.duration == 0 {
			return m, nil
		}
		elapsed := time.Since(m.startTime)
		if elapsed >= m.duration {
			m.remaining = 0
			m.done = true
			m.endTime = time.Now()
			return m, nil
		}
		m.remaining = m.duration - elapsed
		return m, tick()

	case tea.KeyMsg:
		k := msg.String()

		// Global quit
		if k == "ctrl+c" {
			return m, tea.Quit
		}

		if m.done {
			// ignore keys while done; app handles navigation
			return m, nil
		}

		// Start on first keystroke
		if !m.started {
			m.started = true
			m.startTime = time.Now()
			m.remaining = m.duration
			if m.duration > 0 {
				return m, tick()
			}
			return m, nil
		}

		// Backspace handling (support multiple terminals/platforms)
		if msg.Type == tea.KeyBackspace || k == "backspace" || k == "ctrl+h" {
			if len(m.typed) > 0 {
				// adjust stats for the char being removed
				idx := len(m.typed) - 1
				if idx < len(m.target) {
					if m.typed[idx] == m.target[idx] {
						m.correct--
					} else {
						m.incorrect--
					}
				} else {
					// over-typed beyond target counts as incorrect
					m.incorrect--
				}
				m.typed = m.typed[:idx]
				if m.cursor > 0 {
					m.cursor--
				}
			}
			return m, nil
		}

		if msg.Type == tea.KeyEnter {
			// ignore enter during test to avoid accidental submit
			return m, nil
		}

		// Insert typed rune(s). For simplicity, handle single-rune keys.
		if r := msg.Runes; len(r) > 0 {
			ch := r[0]
			m.typed = append(m.typed, ch)
			if m.cursor < len(m.target) {
				if ch == m.target[m.cursor] {
					m.correct++
				} else {
					m.incorrect++
				}
			} else {
				// over-typing beyond target counts as incorrect
				m.incorrect++
			}
			m.cursor++
		}

		// For untimed mode, auto-complete when all chars are typed
		if m.duration == 0 && m.cursor >= len(m.target) {
			m.done = true
			m.endTime = time.Now()
		}
		return m, nil
	}
	return m, nil
}

func (m TestModel) View() string {
	const (
		reset = "\x1b[0m"
		dim   = "\x1b[2m"
		bold  = "\x1b[1m"
		green = "\x1b[32m"
		red   = "\x1b[31m"
		ul    = "\x1b[4m"
		white = "\x1b[97m"
	)

	out := make([]byte, 0, len(m.target)*8)

	caretIdx := len(m.typed)
	// optional soft wrapping based on width
	maxW := m.width
	if maxW <= 0 {
		maxW = 80
	}
	maxW -= 4
	if maxW < 20 {
		maxW = 20
	}
	col := 0
	lastSpace := -1
	newlineAt := make([]bool, len(m.target))
	for i, ch := range m.target {
		if ch == ' ' {
			lastSpace = i
		}
		if col >= maxW && lastSpace >= 0 {
			newlineAt[lastSpace] = true
			col = i - lastSpace - 1
			lastSpace = -1
		}
		col++
	}

	for i, ch := range m.target {
		if i > 0 && newlineAt[i-1] {
			out = append(out, '\n')
		}
		if i < len(m.typed) {
			// already typed
			if m.typed[i] == ch {
				out = append(out, []byte(bold+green)...)
				out = append(out, string(ch)...)
				out = append(out, []byte(reset)...)
			} else {
				out = append(out, []byte(bold+red)...)
				out = append(out, string(ch)...)
				out = append(out, []byte(reset)...)
			}
		} else if i == caretIdx && !m.done {
			out = append(out, []byte(bold+ul+white)...)
			out = append(out, string(ch)...)
			out = append(out, []byte(reset)...)
		} else {
			out = append(out, []byte(bold+dim+white)...)
			out = append(out, string(ch)...)
			out = append(out, []byte(reset)...)
		}
	}

	// live stats line
	elapsed := 0.0
	if m.started {
		if m.done {
			if m.duration > 0 {
				elapsed = m.duration.Seconds()
			} else if !m.startTime.IsZero() && !m.endTime.IsZero() {
				elapsed = m.endTime.Sub(m.startTime).Seconds()
			}
		} else {
			elapsed = time.Since(m.startTime).Seconds()
		}
	}
	minutes := elapsed / 60.0
	wpm := 0.0
	if minutes > 0 {
		wpm = float64(len(m.typed)) / 5.0 / minutes
	}
	acc := 0.0
	if len(m.typed) > 0 {
		acc = float64(m.correct) / float64(len(m.typed)) * 100.0
	}
	// time to display: remaining for timed tests, elapsed for untimed
	timeSec := elapsed
	if m.duration > 0 {
		timeSec = 0
		if m.started && !m.done {
			timeSec = m.remaining.Seconds()
		} else if !m.started {
			timeSec = m.duration.Seconds()
		} else {
			timeSec = 0
		}
	}
	statLine := fmt.Sprintf("WPM %4.1f | ACC %5.1f%% | ERR %d | TIME %4.1fs", wpm, acc, m.incorrect, timeSec)

	// Center per-line horizontally and vertically; show text and stats
	raw := string(out)
	parts := strings.Split(raw, "\n")
	// add a blank line and the stats during the test
	parts = append(parts, "")
	parts = append(parts, statLine)
	for i, ln := range parts {
		l := 0
		if m.width > 0 {
			l = (m.width - visualLen(ln)) / 2
			if l < 0 {
				l = 0
			}
		}
		parts[i] = strings.Repeat(" ", l) + ln
	}
	body := strings.Join(parts, "\n")
	top := 0
	if m.height > 0 {
		top = (m.height - len(parts)) / 2
		if top < 0 {
			top = 0
		}
	}
	padTop := strings.Repeat("\n", top)
	return padTop + body
}

// visualLen counts printable bytes, ignoring ANSI escapes.
func visualLen(s string) int {
	n := 0
	inEsc := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inEsc {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				inEsc = false
			}
			continue
		}
		if c == 0x1b { // ESC
			inEsc = true
			continue
		}
		if c == '\n' {
			continue
		}
		n++
	}
	return n
}

// computes final metrics for the run.
type Stats struct {
	DurationSec float64
	GrossWPM    float64
	NetWPM      float64
	AccuracyPct float64
	Correct     int
	TotalTyped  int
}

func (m TestModel) Stats() Stats {
	totalSec := 0.0
	if m.duration > 0 {
		totalSec = m.duration.Seconds()
	} else if !m.startTime.IsZero() && !m.endTime.IsZero() {
		totalSec = m.endTime.Sub(m.startTime).Seconds()
	}
	minutes := 0.0
	if totalSec > 0 {
		minutes = totalSec / 60.0
	}
	gross := 0.0
	if minutes > 0 {
		gross = float64(len(m.typed)) / 5.0 / minutes
	}
	acc := 0.0
	if len(m.typed) > 0 {
		acc = float64(m.correct) / float64(len(m.typed)) * 100.0
	}
	return Stats{
		DurationSec: totalSec,
		GrossWPM:    gross,
		NetWPM:      gross * (acc / 100.0),
		AccuracyPct: acc,
		Correct:     m.correct,
		TotalTyped:  len(m.typed),
	}
}
