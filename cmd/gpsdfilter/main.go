package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/mpihlak/go-nmealogger"
)

func main() {
	gpsdAddress := flag.String("gpsd", "localhost:2947", "gpsd address")
	listenAddress := flag.String("listen", "localhost:2948", "Local address to listen to")
	flag.Parse()

	log.Printf("Listening on %s", *listenAddress)
	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		log.Fatalf("Error listening on port: %v", err)
	}

	// Keep it simple and accept only a single connection at a time. Each client connection
	// gets their own gpsd connection.

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}
		log.Printf("Connection accepted from %v", conn.RemoteAddr())

		if err := handleClient(conn, *gpsdAddress); err != nil {
			log.Printf("Error handling client connection: %v", err)
			continue
		}
	}
}

func handleClient(client net.Conn, gpsdAddress string) error {
	defer client.Close()

	gpsd, err := connectGpsd(gpsdAddress)
	if err != nil {
		return fmt.Errorf("error connecting to gpsd: %w", err)
	}
	defer gpsd.Close()
	gpsdReader := bufio.NewReader(gpsd)

	for {
		data, err := gpsdReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading from gpsd: %w", err)
		}

		sentence := strings.TrimRight(string(data), "\r\n")
		outSentence := ""
		if strings.HasPrefix(sentence, "$GPGLL") {
			outSentence = rewritePosition(sentence)
		} else if strings.HasPrefix(sentence, "$GPRMC") {
			outSentence = rewriteNavigationInfo(sentence)
		} else {
			log.Printf("Ignoring sentence: [%s]", sentence)
		}

		if outSentence != "" {
			checksum := nmealogger.CalculateChecksum(outSentence[1:])
			outSentence = outSentence + "*" + checksum
			log.Printf("Publishing: [%s]", outSentence)
			if _, err := client.Write([]byte(outSentence + "\r\n")); err != nil {
				return fmt.Errorf("error writing to client: %w", err)
			}
		}
	}
}

func connectGpsd(gpsdAddress string) (net.Conn, error) {
	log.Printf("Connecting to gpsd at %s", gpsdAddress)

	conn, err := net.Dial("tcp", gpsdAddress)
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to gpsd, starting watch")

	_, err = conn.Write([]byte(`?WATCH={"enable":true,"nmea":true}`))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// rewritePosition takes the NMEA GLL sentence and rewrites it to be compatible with TackTick T122
// https://gpsd.gitlab.io/gpsd/NMEA.html#_gll_geographic_position_latitudelongitude
// Example: $GPGLL,5928.09769,N,02449.29616,E,180019.00,A,A*6E
// T122:    $IIGLL,5928.888,N,02449.039,E,142800,A,A*50
// Return:  $GPGLL,5928.097,N,02449.296,E,180019,A,A*6E
// Field changes:
// 1. Latitude truncate to 3 digits after comma
// 3. Longitude truncate to 3 digits after comma
// 5. Truncate time to integer
func rewritePosition(sentence string) string {
	pieces := strings.Split(sentence, ",")
	if len(pieces) != 8 {
		return ""
	}

	// Latitude ####.###
	pieces[1] = reformatFloat(pieces[1], "%08.03f")
	// Longitude #####.###
	pieces[3] = reformatFloat(pieces[3], "%09.03f")
	// Time
	pieces[5] = reformatFloat(pieces[5], "%.0f")

	// FAA indicator replaces checksum
	pieces[7] = "A"

	return strings.Join(pieces, ",")
}

// rewriteNavigationInfo takes the NMEA RMC sentence and rewrites it to be compatible with TackTick T122
// https://gpsd.gitlab.io/gpsd/NMEA.html#_rmc_recommended_minimum_navigation_information
// Example: $GPRMC,180021.00,A,5928.09760,N,02449.29615,E,0.031,,090724,,,A*7C
// T122:    $IIRMC,142800,A,5928.888,N,02449.039,E,04.7,144,060923,00,E,A*00
// Return:  $GPRMC,180021,A,5928.097,N,02449.296,E,00.0,###,090724,00,E,A*7C
// Field changes:
//  1. The time needs to be an integer
//  3. Latitude truncated to 3 digits after comma
//  5. Longitude truncated to 3 digits after comma
//  7. Speed is formatted as ##.#
//  8. Add Course Made Good (GPS heading). Maybe - check if it doesn't have this for moving boat
//     Reformat as integer.
//  10. Add magnetic variation degrees
//  11. Add magnetic variation direction (E/W)
func rewriteNavigationInfo(sentence string) string {
	pieces := strings.Split(sentence, ",")
	if len(pieces) != 13 {
		return ""
	}

	// Time
	pieces[1] = reformatFloat(pieces[1], "%.0f")
	// Latitude ####.###
	pieces[3] = reformatFloat(pieces[3], "%08.03f")
	// Longitude #####.###
	pieces[5] = reformatFloat(pieces[5], "%09.03f")
	// Speed
	pieces[7] = reformatFloat(pieces[7], "%04.01f")
	// Course Made Good
	pieces[8] = reformatFloat(pieces[8], "%.0f")
	if pieces[8] == "" {
		pieces[8] = "000"
	}

	// Add zero magnetic variation
	pieces[10] = "00"
	pieces[11] = "E"

	// $IIRMC,161700,A,5928.095,N,02449.291,E,00.0,051,110724,00,E,A*03
	// $GPRMC,161712,A,5928.099,N,02449.296,E,00.0,   ,110724,00,E*45

	// Nav status replaces checksum
	pieces[12] = "A"

	return strings.Join(pieces, ",")
}

func reformatFloat(strValue string, format string) string {
	if f, err := strconv.ParseFloat(strValue, 64); err != nil {
		return ""
	} else {
		return fmt.Sprintf(format, f)
	}
}
