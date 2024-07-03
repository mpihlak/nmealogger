package nmealogger

import (
	"fmt"
	"strings"
)

// HasValidChecksum tests that the provided NMEA sentence contains a valid
// checksum. The sentence is of the form "$IIMWV,129,R,22.5,N,A*1C" and the
// checksum here is "1C" (the part after '*').
func HasValidChecksum(sentence string) bool {
	if !strings.HasPrefix(sentence, "$") {
		return false
	}

	data, providedChecksum, ok := strings.Cut(sentence, "*")
	if !ok {
		return false
	}

	return CalculateChecksum(data[1:]) == providedChecksum
}

// CalculateChecksum calculates the XOR checksum for an NMEA sentence. It assumes
// that the checksum part and leading $ are already stripped from the sentence.
func CalculateChecksum(strippedSentence string) string {
	var checksum byte
	for _, c := range strippedSentence {
		checksum ^= byte(c)
	}

	return fmt.Sprintf("%02X", checksum)
}
