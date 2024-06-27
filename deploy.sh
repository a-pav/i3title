#!/bin/sh

deploy_local() {
	go build -ldflags="-s -w" -o i3title \
		&& killall -q i3title

	sleep 0.2
	cp ./i3title ~/.config/i3status/modules/i3title/i3title
	cp ./config.json ~/.config/i3status/modules/i3title/config.json

	sleep 0.2
	i3-msg restart
}

deploy_local
