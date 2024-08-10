package nmealogger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogWriter struct {
	lastRotationTime     time.Time
	fileRotationInterval time.Duration
	outputDirectory      string
	writer               io.WriteCloser
}

func NewLogWriter(outputDirectory string, fileRotationInterval time.Duration) *LogWriter {
	return &LogWriter{
		lastRotationTime:     time.Now(),
		fileRotationInterval: fileRotationInterval,
		outputDirectory:      outputDirectory,
		writer:               nil,
	}
}

func (lw *LogWriter) Write(sentence string) error {
	writer, err := lw.getWriter()
	if err != nil {
		return err
	}

	// TODO: Make time format a const
	entry := fmt.Sprintf("%s\t%s\n", time.Now().UTC().Format("2006-01-02T15:04:05.999-0700"), sentence)
	_, err = writer.Write([]byte(entry))

	return err
}

func (lw *LogWriter) Close() {
	if lw.writer != nil {
		if err := lw.writer.Close(); err != nil {
			log.Printf("Error closing active file: %v", err)
		}
		lw.writer = nil
	}
}

func (lw *LogWriter) getWriter() (io.Writer, error) {
	if time.Since(lw.lastRotationTime) > lw.fileRotationInterval {
		lw.Close()
		lw.lastRotationTime = time.Now()
	}

	if lw.writer == nil {
		fileName := fmt.Sprintf("nmea-%s.log", time.Now().UTC().Format("2006-01-02T150405"))
		pathName := filepath.Join(lw.outputDirectory, fileName)
		log.Printf("Writing to %s", pathName)

		var err error
		lw.writer, err = os.Create(pathName)
		if err != nil {
			return nil, fmt.Errorf("error opening %s for writing: %w", pathName, err)
		}
	}

	return lw.writer, nil
}
