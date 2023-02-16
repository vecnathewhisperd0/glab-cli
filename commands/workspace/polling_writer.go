package workspace

import (
	"fmt"
	"io"
	"time"

	"github.com/gosuri/uilive"
)

// pollingWriter periodically writes to a nested StdOut io.Writer using a data generating function
// that is passed to runRenderLoop.
//
// This was written with the goal of supporting live updates / watch features in the terminal
type pollingWriter struct {
	pollingInterval time.Duration // polling interval in seconds
	writer          *uilive.Writer
}

func newPollingWriter(stdOut io.Writer, pollingInterval time.Duration) *pollingWriter {
	writer := uilive.New()
	writer.Out = stdOut
	return &pollingWriter{
		pollingInterval: pollingInterval,
		writer:          writer,
	}
}

// runRenderLoop is a blocking call that periodically invokes dataGenerator passed and renders the terminal in case of no errors
// When dataGenerator returns an error, runRenderLoop will cease execution and return the error
func (w *pollingWriter) runRenderLoop(dataGenerator func() (string, error)) error {
	ticker := time.NewTicker(w.pollingInterval)
	defer ticker.Stop()

	var lastRendered *string
	for {
		toRender, err := dataGenerator()
		if err != nil {
			return err
		}

		if lastRendered == nil || *lastRendered != toRender {
			fmt.Fprint(w.writer, toRender)

			// calling Flush() clears the previous output and re-renders the terminal with the latest output
			w.writer.Flush()

			lastRendered = &toRender
		}

		<-ticker.C
	}
}
