#!/bin/sh

# Check available dependency updates:
# 	go list -m -u all

main() {
	if ! go build -ldflags="-s -w" -o i3title; then
		return
	fi

	if [ ! "$(realpath "$HOME/.local/bin/i3title")" = "$(realpath ./i3title)" ]; then
		ln -sf "$(realpath ./i3title)" "$HOME/.local/bin/i3title"
	fi

	killall -q i3title

	sleep 0.2
	i3-msg restart
}

main
