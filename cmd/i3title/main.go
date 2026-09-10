package main

import (
	"fmt"
	"log"
	"os"

	"github.com/a-pav/i3title/internal/config"
	"github.com/a-pav/i3title/internal/report"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-v" {
		fmt.Println("i3title", Version)
		return
	}

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

	return report.Run(cfg)
}
