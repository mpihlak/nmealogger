package main

import "testing"

func TestRewritePosition(t *testing.T) {
	sentence := "$GPGLL,5928.09769,N,02449.29616,E,180019.00,A,A*6E"
	expected := "$GPGLL,5928.098,N,02449.296,E,180019,A,A"
	result := rewritePosition(sentence)
	if result != expected {
		t.Fatalf("Expected [%s], got [%s]", expected, result)
	}
}

func TestRewriteNavigationInfo(t *testing.T) {
	sentence := "$GPRMC,180021.00,A,5928.09760,N,02449.29615,E,0.031,,090724,,,A*7C"
	expected := "$GPRMC,180021,A,5928.098,N,02449.296,E,00.0,000,090724,00,E,A"
	result := rewriteNavigationInfo(sentence)
	if result != expected {
		t.Fatalf("Expected [%s], got [%s]", expected, result)
	}
}
