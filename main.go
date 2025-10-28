package main

import (
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	log.SetPrefix("i3title: ")
	log.SetFlags(log.Lmsgprefix)

	if err := loadConfig(); err != nil {
		return err
	}

	var (
		lineCh    = make(chan []byte)
		messageCh = make(chan []byte)
		titleCh   = make(chan string)
		modeCh    = make(chan string)
	)
	go reporter(lineCh, messageCh, titleCh, modeCh)

	emitLines(lineCh)
	emitMessages(messageCh)
	go emitTitles(titleCh)
	go emitModes(modeCh)

	select {} // Block forever.
}
