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

	liner(lineCh)
	messagePipe(messageCh)
	go titler(titleCh)
	go moder(modeCh)

	select {} // Block forever.
}
