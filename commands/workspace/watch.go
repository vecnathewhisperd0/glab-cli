package workspace

import (
	"fmt"
	"io"
	"time"

	"github.com/gosuri/uilive"
)

// todo: add docs describing this
type watchWriter struct {
	pollingInterval time.Duration // polling interval in seconds
	writer          *uilive.Writer
}

func newWatchWriter(nested io.Writer, pollingInterval time.Duration) *watchWriter {
	writer := uilive.New()
	writer.Out = nested
	return &watchWriter{
		pollingInterval: pollingInterval,
		writer:          writer,
	}
}

// todo: describe behavior esp wrt errors and how that will stop the render loop
func (w *watchWriter) runRenderLoop(dataGenerator func() (string, error)) error {
	w.writer.Start()
	defer w.writer.Flush()

	ticker := time.NewTicker(w.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			toRender, err := dataGenerator()
			if err != nil {
				return err
			}

			fmt.Fprint(w.writer, toRender)
		}
	}
}
