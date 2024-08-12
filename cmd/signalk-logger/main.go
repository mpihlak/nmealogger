package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// At most this amount of data in the logfile. Keep reasonably small to improve data
	// freshness (uploader only considers rotated files)
	FileRotationInterval = 5 * time.Minute
	// Ask SignalK to report measurements at this interval
	SignalKReportingIntervalMs = 1 * time.Second
	// If no data has been received within this time write the record anyway
	SignalKMissingDataTimeout = 2 * time.Second
	// Drop data that is older than the stale threshold
	SkipStaleDataThreshold = 15 * time.Second
)

func main() {
	logDirectory := flag.String("logDir", "data", "Directory where log files will be stored")
	signalK := flag.String("signalk-addr", "localhost:3000", "SignalK hostport")
	flag.Parse()

	log.Printf("Starting SignalK logger: log directory = %s, signalK = %s", *logDirectory, *signalK)

	if err := os.MkdirAll(*logDirectory, os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	u := url.URL{
		Scheme:   "ws",
		Host:     *signalK,
		Path:     "/signalk/v1/stream",
		RawQuery: "subscribe=none",
	}

	log.Printf("Connecting to %s", u.String())

	for {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("Error connecting to SignalK: %v", err)
			log.Printf("Retrying ...")
			time.Sleep(1 * time.Second)
			continue
		}

		log.Println("Connected to SignalK, start processing")
		processMessages(conn, *logDirectory)
	}
}

type Topic struct {
	Path   string `json:"path"`
	Period int    `json:"period"`
}

type Subscriptions struct {
	Context   string  `json:"context"`
	Subscribe []Topic `json:"subscribe"`
}

type SignalKMessage struct {
	Context string `json:"context"`
	Updates []struct {
		Source struct {
			Label          string `json:"label"`
			Type           string `json:"type"`
			Pgn            int    `json:"pgn"`
			Src            string `json:"src"`
			DeviceInstance int    `json:"deviceInstance"`
		} `json:"source"`
		SourceRef string    `json:"$source"`
		Timestamp time.Time `json:"timestamp"`
		Values    []struct {
			Path  string      `json:"path"`
			Value interface{} `json:"value"`
		} `json:"values"`
	} `json:"updates"`
}

func processMessages(c *websocket.Conn, logDirectory string) error {
	defer c.Close()

	_, helloMsg, err := c.ReadMessage()
	if err != nil {
		log.Printf("Error reading Hello message from signalK: %v", err)
		return err
	}
	log.Printf("Got hello: %v\n", string(helloMsg))
	log.Printf("Subscribing ...")

	subscriptions := Subscriptions{
		Context: "vessels.self",
		Subscribe: []Topic{
			{Path: "environment.depth.belowTransducer", Period: 1000},
			{Path: "environment.water.temperature", Period: 1000},
			{Path: "environment.wind.angleApparent", Period: 1000},
			{Path: "environment.wind.speedApparent", Period: 1000},
			{Path: "navigation.courseOverGroundTrue", Period: 1000},
			{Path: "navigation.datetime", Period: 1000},
			{Path: "navigation.headingMagnetic", Period: 1000},
			{Path: "navigation.magneticVariation", Period: 1000},
			{Path: "navigation.rateOfTurn", Period: 1000},
			{Path: "navigation.speedOverGround", Period: 1000},
			{Path: "navigation.speedThroughWater", Period: 1000},
			{Path: "navigation.attitude", Period: 1000},
			{Path: "navigation.position", Period: 1000},
		},
	}

	// If a measurement has multiple sources we need to choose which one to use.
	// ignoreSources map specifies values to drop from specific sources.
	ignoreSources := map[string]string{
		"environment.wind.angleApparent":  "can0.15",
		"environment.wind.speedApparent":  "can0.15",
		"navigation.courseOverGroundTrue": "can0.85",
		"navigation.datetime":             "can0.85",
		"navigation.headingMagnetic":      "can0.85",
		"navigation.magneticVariation":    "can0.85",
		"navigation.speedOverGround":      "can0.85",
		"navigation.position":             "can0.85",
	}

	requiredFields := []string{
		"environment.depth.belowTransducer",
		"environment.water.temperature",
		"environment.wind.angleApparent",
		"environment.wind.speedApparent",
		"navigation.courseOverGroundTrue",
		"navigation.headingMagnetic",
		"navigation.magneticVariation",
		"navigation.rateOfTurn",
		"navigation.speedOverGround",
		"navigation.speedThroughWater",
		"navigation.attitude.pitch",
		"navigation.attitude.yaw",
		"navigation.attitude.roll",
		"navigation.position.longitude",
		"navigation.position.latitude",
	}

	logWriter := NewSignalKLogWriter(logDirectory, requiredFields, SignalKMissingDataTimeout, FileRotationInterval)
	defer logWriter.Close()

	buf, err := json.Marshal(subscriptions)
	if err != nil {
		log.Fatalf("Error marshaling subscriptions to JSON: %v", err)
	}

	if err := c.WriteMessage(websocket.TextMessage, buf); err != nil {
		log.Printf("Error subscribing to SignalK: %v", err)
		return err
	}

	log.Printf("Reading messages ...")
	record := NewRecord()
	for {
		_, buf, err := c.ReadMessage()
		if err != nil {
			log.Printf("Error reading from SignalK: %v", err)
			return err
		}

		var message SignalKMessage
		err = json.Unmarshal(buf, &message)
		if err != nil {
			log.Printf("Error unmarshalling SignalK message [%v]: %v", string(buf), err)
			continue
		}

		for _, update := range message.Updates {
			for _, value := range update.Values {
				if ignoreSources[value.Path] == update.SourceRef {
					continue
				}
				if time.Now().Sub(update.Timestamp) > SkipStaleDataThreshold {
					log.Printf("Ignoring stale field: %s %v", value.Path, update.Timestamp)
					continue
				}
				if val, ok := value.Value.(float64); ok {
					record.AddValue(update.Timestamp, value.Path, val)
				} else if val, ok := value.Value.(string); ok {
					// log.Printf("Ignoring string value: %v=%v", value.Path, val)
					// Ignore string values
				} else if valueMap, ok := value.Value.(map[string]interface{}); ok {
					for k, v := range valueMap {
						if val, ok := v.(float64); ok {
							recordKey := fmt.Sprintf("%v.%v", value.Path, k)
							record.AddValue(update.Timestamp, recordKey, val)
						} else {
							log.Printf("Ignoring unknown map value: %v.%v=%v", value.Path, k, val)
						}
					}
				} else {
					log.Printf("Ignoring unknown value: %v=%v", value.Path, val)
				}
			}
		}

		logWriter.AddRecord(record)
	}
}
