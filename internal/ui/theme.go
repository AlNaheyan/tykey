package ui

import "fmt"

//  ANSI colors
const (
    AnsiReset     = "\x1b[0m"
    AnsiDim       = "\x1b[2m"
    AnsiBold      = "\x1b[1m"
    AnsiUnderline = "\x1b[4m"
    AnsiWhite     = "\x1b[97m"
    AnsiGreen     = "\x1b[32m"
    AnsiRed       = "\x1b[31m"
)

// ansi color
func FgTrueColor(r, g, b int) string {
    return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

// main theme color (#524a6b)
var Accent = FgTrueColor(82, 74, 107)
var AccentBold = AnsiBold + Accent

