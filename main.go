package main

import (
    "math/rand"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "termtype/internal/ui"
)

func main() {
    rand.Seed(time.Now().UnixNano())
    p := tea.NewProgram(ui.NewApp(), tea.WithAltScreen())
    _ = p.Start()
}
