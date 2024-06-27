#!/bin/sh

deploy_local() {
	go build -ldflags="-s -w" -o i3title \
		&& killall -q i3title

	cp ./i3title ~/.config/i3status/modules/i3title/i3title
	cp ./config.json ~/.config/i3status/modules/i3title/config.json
}

deploy_local
