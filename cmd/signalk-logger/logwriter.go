package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type SignalKLogWriter struct {
	lastRotationTime     time.Time
	fileRotationInterval time.Duration
	outputDirectory      string
	writer               io.WriteCloser
	csvWriter            *csv.Writer
	requiredFields       []string
	missingFieldsTimeout time.Duration
	lastWrite            time.Time
}

func NewSignalKLogWriter(
	outputDirectory string,
	requiredFields []string,
	missingFieldsTimeout time.Duration,
	fileRotationInterval time.Duration,
) *SignalKLogWriter {
	return &SignalKLogWriter{
		lastRotationTime:     time.Now(),
		fileRotationInterval: fileRotationInterval,
		outputDirectory:      outputDirectory,
		writer:               nil,
		requiredFields:       requiredFields,
		missingFieldsTimeout: missingFieldsTimeout,
		lastWrite:            time.Now(),
	}
}

func (lw *SignalKLogWriter) AddRecord(record *Record) error {
	forceWrite := false
	if !lw.lastWrite.IsZero() && lw.lastWrite.Before(time.Now().Add(-lw.missingFieldsTimeout)) {
		forceWrite = true
	}

	if record.HasRequiredFields(lw.requiredFields) || forceWrite {
		values := []string{time.Now().Format(time.RFC3339)}
		for _, field := range lw.requiredFields {
			strVal := ""
			if val, ok := record.Values[field]; ok {
				strVal = fmt.Sprintf("%f", val)
			}
			values = append(values, strVal)
		}

		csvWriter, err := lw.getWriter()
		if err != nil {
			return err
		}

		record.Clear()
		lw.lastWrite = time.Now()

		err = csvWriter.Write(values)
		csvWriter.Flush()
		return err
	}

	return nil
}

func (lw *SignalKLogWriter) Close() {
	if lw.writer != nil {
		lw.csvWriter.Flush()
		if err := lw.writer.Close(); err != nil {
			log.Printf("Error closing active file: %v", err)
		}
		lw.writer = nil
	}
}

func (lw *SignalKLogWriter) getWriter() (*csv.Writer, error) {
	if time.Since(lw.lastRotationTime) > lw.fileRotationInterval {
		lw.Close()
		lw.lastRotationTime = time.Now()
	}

	if lw.writer == nil {
		fileName := fmt.Sprintf("signalk-%s.log", time.Now().UTC().Format("2006-01-02T150405"))
		pathName := filepath.Join(lw.outputDirectory, fileName)
		log.Printf("Writing to %s", pathName)

		var err error
		lw.writer, err = os.Create(pathName)
		if err != nil {
			return nil, fmt.Errorf("error opening %s for writing: %w", pathName, err)
		}

		lw.csvWriter = csv.NewWriter(lw.writer)
		// TODO: Write the header with the record so that all writing is in one place
		// also that we don't have just files with headers
		header := append([]string{"time"}, lw.requiredFields...)
		if err := lw.csvWriter.Write(header); err != nil {
			// TODO: Close and reset the io.Writer
			return nil, fmt.Errorf("error writing CSV header: %w", err)
		}
		lw.csvWriter.Flush()
	}

	return lw.csvWriter, nil
}
