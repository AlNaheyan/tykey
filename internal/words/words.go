package words

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
	"sync"
)

var (
	bank    []string
	once    sync.Once
	loadErr error
)

func load() error {
	once.Do(func() {
		f, err := os.Open("data/wordbank.txt")
		if err != nil {
			loadErr = err
			return
		}
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			w := strings.TrimSpace(s.Text())
			if w != "" {
				bank = append(bank, w)
			}
		}
		if err := s.Err(); err != nil {
			loadErr = err
		}
		if len(bank) == 0 {
			// fallback to a tiny set if file empty
			bank = []string{"the", "quick", "brown", "fox", "jumps", "over", "the", "lazy", "dog"}
		}
	})
	return loadErr
}

// (clamped to the size of the wordbank).
func Generate(n int) ([]string, error) {
	if err := load(); err != nil {
		return nil, err
	}
	if n <= 0 {
		n = 1
	}
	if n > len(bank) {
		n = len(bank)
	}
	// sample indices without modifying the original slice
	idx := make([]int, len(bank))
	for i := range idx {
		idx[i] = i
	}
	rand.Shuffle(len(idx), func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = bank[idx[i]]
	}
	return out, nil
}

// GenerateString returns a space-joined string of n random words.
func GenerateString(n int) (string, error) {
	ws, err := Generate(n)
	if err != nil {
		return "", err
	}
	return strings.Join(ws, " "), nil
}
