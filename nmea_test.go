package nmealogger

import "testing"

func TestCalculateChecksum(t *testing.T) {
	cksum := CalculateChecksum("IIVLW,09390,N,000.0,N")
	if cksum != "50" {
		t.Fatalf("Incorrect checksum %v", cksum)
	}

	cksum = CalculateChecksum("IIMWV,127,R,21.8,N,A")
	if cksum != "1C" {
		t.Fatalf("Incorrect checksum %v", cksum)
	}
}

func TestHasValidChecksum(t *testing.T) {
	if HasValidChecksum("") {
		t.Fatal("Expected checksum to be invalid for empty string")
	}

	sentence := "$IIVLW,09390,N,000.0,N*50"

	if HasValidChecksum(sentence[1:]) {
		t.Fatal("Expected checksum to be invalid for sentence not starting with $")
	}
	if !HasValidChecksum(sentence) {
		t.Fatal("Expected checksum to be valid for sentence")
	}
}
