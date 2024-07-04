package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	inputFile := flag.String("inputFile", "", "Name of the input file")
	optionalStartTime := flag.String("startTime", "", "Start time of replay, format 2006-01-02T15:04:05, UTC time zone")
	flag.Parse()

	buf, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	var startTime time.Time
	if *optionalStartTime != "" {
		startTime, err = time.Parse("2006-01-02T15:04:05", *optionalStartTime)
		if err != nil {
			log.Fatalf("Error parsing start time: %v", err)
		}
	}

	listenAddr := "0.0.0.0:10110"
	log.Printf("Listening on %s", listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Error listening on port: %v", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}
		log.Printf("Connection accepted from %v", conn.RemoteAddr())

		go nmeaReplay(conn, buf, startTime)
	}
}

func nmeaReplay(conn net.Conn, buf []byte, startTime time.Time) {
	prevTime := time.Time{}
	totalBytes := 0
	for _, line := range strings.Split(string(buf), "\n") {
		timestamp, sentence, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}

		currTime, err := time.Parse("2006-01-02T15:04:05.9+00:00", timestamp)
		if err != nil {
			log.Printf("Error parsing timestamp: %v", err)
		}

		if startTime.After(currTime) {
			continue
		}

		if !prevTime.IsZero() {
			delta := currTime.Sub(prevTime)
			time.Sleep(delta)
		}

		bytes, err := conn.Write([]byte(sentence + "\n"))
		if err != nil {
			log.Printf("Error writing to %s: %v", conn.RemoteAddr(), err)
			conn.Close()
			return
		}

		log.Printf("sent to %s: [%s] @%v", conn.RemoteAddr(), sentence, currTime)

		totalBytes += bytes
		prevTime = currTime
	}
	log.Printf("%s finished, sent %d total bytes.", conn.RemoteAddr(), totalBytes)
}
