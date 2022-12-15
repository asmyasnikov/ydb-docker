package writer

import (
	"bytes"
	"io"
)

type prefixedWriter struct {
	backend io.Writer
	prefix  string

	streamStarted  bool
	lineIncomplete bool
}

func Prefixed(prefix string, writer io.Writer) *prefixedWriter {
	return &prefixedWriter{
		backend: writer,
		prefix:  prefix,
	}
}

func (writer *prefixedWriter) Write(data []byte) (int, error) {
	var (
		reader         = bytes.NewBuffer(data)
		eofEncountered = false
	)

	for !eofEncountered {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return 0, err
			}

			eofEncountered = true
		}

		if line == "" {
			continue
		}

		if !writer.streamStarted || !writer.lineIncomplete {
			line = writer.prefix + line

			writer.streamStarted = true
		}

		writer.lineIncomplete = eofEncountered

		_, err = writer.backend.Write([]byte(line))
		if err != nil {
			return 0, err
		}
	}

	return len(data), nil
}
