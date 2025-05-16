package cesta

import (
	"log"
	"testing"
)

func TestCurrentBase34Date(t *testing.T) {
	days := nowUnixDays()
	date := CurrentBase34Date()
	log.Printf("unix days: %d", days)
	log.Printf("date code: %s", date)
}
