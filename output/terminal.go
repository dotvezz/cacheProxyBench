package output

import (
	"fmt"
	"os"

	"github.com/charmbracelet/x/term"
	"github.com/dotvezz/cacheProxyBench/benchmark"
)

func ToTerminal(rs []benchmark.Result) error {
	width, height, err := term.GetSize(os.Stdout.Fd())

	if err != nil {
		return fmt.Errorf("failed to get terminal dimensions: %w", err)
	}

	xResolution := len(rs) / width

	return nil
}
