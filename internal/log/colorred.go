package log

import (
	"io"
	"os"
	"sync"
)

type Colour string

var (
	resetColour = Colour("\u001b[0m")
	colours     = []Colour{
		Colour("\u001b[38;5;61m"),
		Colour("\u001b[38;5;34m"),
		Colour("\u001b[38;5;178m"),
		Colour("\u001b[38;5;208m"),
		Colour("\u001b[38;5;166m"),
	}
)

func NextColour() (c Colour) {
	if len(colours) == 0 {
		return c
	}
	c = colours[0]
	colours = colours[1:]
	return c
}

type coloredWriter struct {
	mu     sync.Mutex
	colour Colour
	writer io.Writer
}

func (w *coloredWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err = w.writer.Write([]byte(string(w.colour) + string(p) + string(resetColour)))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func Colored(colour Colour) (out io.Writer, err io.Writer) {
	return &coloredWriter{colour: colour, writer: os.Stdout}, &coloredWriter{colour: colour, writer: os.Stderr}
}
