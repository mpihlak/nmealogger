package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"

	nmealogger "github.com/mpihlak/go-nmealogger"
)

const (
	FileRotationInterval   = 5 * time.Minute
	StatsReportingInterval = 60 * time.Second
)

func main() {
	logDirectory := flag.String("logDir", "data", "Directory where log files will be stored")
	kplex := flag.String("kplex", "127.0.0.1:10110", "Kplex server hostport")
	flag.Parse()

	log.Printf("Starting NMEA logger: log directory = %s, kplex = %s", *logDirectory, *kplex)

	if err := os.MkdirAll(*logDirectory, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	for {
		conn, err := net.Dial("tcp", *kplex)
		if err != nil {
			log.Printf("Error connecting to kplex: %v", err)
			log.Printf("Retrying ...")
			time.Sleep(1 * time.Second)
			continue
		}

		log.Printf("Connected to kplex, start processing messages")
		processMessages(conn, *logDirectory)
	}
}

func processMessages(conn net.Conn, outputDirectory string) {
	reader := bufio.NewReader(conn)

	logWriter := NewNMEALogWriter(outputDirectory, FileRotationInterval)
	defer logWriter.Close()

	statsLastReported := time.Now()
	messagesProcessed := 0
	messagesSkipped := 0

	for {
		if time.Since(statsLastReported) > StatsReportingInterval {
			log.Printf("%d sentences logged, %d skipped", messagesProcessed, messagesSkipped)
			statsLastReported = time.Now()
			messagesProcessed = 0
			messagesSkipped = 0
		}

		data, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from kplex: %v", err)
			return
		}

		sentence := strings.TrimRight(string(data), "\r\n")
		if !nmealogger.HasValidChecksum(sentence) {
			log.Printf("Skipping sentence with invalid checksum: [%s]", sentence)
			messagesSkipped += 1
			continue
		}

		if err := logWriter.Write(sentence); err != nil {
			log.Printf("Error writing log entry: %v", err)
			return
		}

		messagesProcessed += 1
	}
}
