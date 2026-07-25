package main

import (
	"log"

	"github.com/a-pav/i3title/internal/config"
	"github.com/a-pav/i3title/internal/report"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	log.SetPrefix("i3title: ")
	log.SetFlags(log.Lmsgprefix)

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := report.Run(cfg); err != nil {
		return err
	}

	select {} // Block forever.
}
