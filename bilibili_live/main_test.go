package main

import (
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	insertLive(22323445, "毬亚Maria")
}

func TestStart(t *testing.T) {
	initDB()
	startListen()
	time.Sleep(200 * time.Second)
}
