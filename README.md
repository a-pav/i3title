<p align="center">
  <img width="321" height="26" alt="demo" src="https://github.com/user-attachments/assets/acb34afb-f877-4708-aff7-b727314b234e" />
</p>


# i3title

> **Fill the empty black space in your status bar with useful information.**

`i3title` is a lightweight companion tool for the [i3 window manager](https://i3wm.org/).
It seamlessly plugs into the gap left between your workspace buttons and the status area,
turning that unused dead space into a dynamic, highly customizable information panel.

By default, `i3title` displays the active window title and the current i3 mode
(e.g., "resize" or "move"). More importantly, you can pipe arbitrary data into it,
format it on the fly, and trigger instant updates—all without touching your configurations or restarting i3.

---

## Demo



https://github.com/user-attachments/assets/7846a025-db57-4143-86a4-0313c08b74e7


## Features
 - **Zero Configuration Defaults** – Shows the active window title out of the box.
 - **Fully Customizable** – Accepts arbitrary text via standard input or command-line arguments, letting you display system metrics, weather, custom scripts output, or anything else.
 - **Background Workspace Awareness** – Reports on work being done on other workspaces, keeping you informed of background activity without needing to switch views.
 - **Real-time Updates** – Instantly refreshes on workspace switches, window focus changes, background workspace activity, i3 mode toggles, and new piped data.
 - **i3bar Native** – Designed specifically for i3bar, respecting your existing configurations.
 - **Lightweight & Fast** – Written in Go, with a **minimal footprint**.

---

## Installation

### From source
```bash
git clone https://github.com/a-pav/i3title
cd i3title
make
make install
```
*(By default, this installs the binaries to `~/.local/bin`. Make sure this directory is in your `$PATH`.)*

---

## Configuration

### i3

Once the binary is in your `$PATH`, add the following to your `~/.config/i3/config` :

```sh
bar {
    status_command i3status | i3title
    # Recommended if you want to have i3title show you the current mode:
    binding_mode_indicator no

    # ... your other bar settings
}
```

### i3status

Add the following to your `~/.config/i3status/config` :

```sh
general {
        markup = "pango"

        # ... your other general settings
}
```

### i3title

See [config.example.json](config.example.json) for a fully documented example.
Each config key has a detailed comment explaining its purpose and usage.

**Restart i3** (`$mod+shift+r`). You'll now see the active window title as the first
status item after your workspace buttons. The first run might not be perfect depending
on your config, but once you find your ideal setup, you won't need to change it again.

---

## Displaying Your Data: Enter **`i3toast`**

As an i3 user, you are likely familiar with `i3-msg` for sending IPC messages to the
window manager from the command line and scripts. `i3toast` is the exact same concept,
but for sending messages to `i3title`. (In fact, it was originally named
`i3title-msg` for this reason.)

Its most basic use is sending one-shot, "toast-like" messages from your scripts
that appear on your bar and disappear after a timeout (default is 4 seconds.)

### One-shot messages (Toasts)

Send a quick, temporary update from a script or keybind. Perfect for volume or brightness adjustments:

```sh
# This could be from your volume adjustment script, informing you about the change.
i3toast -m "🔊 ${volume}%" -f "<span font='bold italic 15'>%s</span>"
```

### Piping continuous data

You can also pipe standard output directly into `i3toast` for real-time updates.
Common use cases include:

- Show the current date and time for 7 seconds, then disappear instantly.

    ```bash
    for i in {1..7}; do
        date '+%A, %-d %B %Y --- %Y-%m-%d %H:%M:%S'
        sleep 1
    done |
        i3toast -t 0 -f "<span font='14'>📅</span> <span color='#CFD8DC'>%s</span>"
    ```

  - **Pro tip:** Displaying information on-demand like this means you can disable those
corresponding status modules in your i3status config. This frees up space, giving
`i3title` more room to fit longer window titles and data.

- Print to the terminal, but keep the tail with you on whatever workspace you switch to:
    ```sh
    ./main | tee /dev/tty | i3toast
    ```
- Take full control:
    ```sh
    {
        ### Code your logic here and control the flow:

        # Whatever you print to stdout is considered a message.
        echo "This is a message!"

        ### You can tag the stdout like `i3toast:<cmd>[:<payload>]` to instruct
        ### i3toast itself mid-stream:

        # Set the offset
        echo "i3toast:off:10"
        # Set the format
        echo "i3toast:fmt:<span color='red'>PREFIX:</span> %s"
        # Set the trimming length
        echo "i3toast:trm:60"
        # Erase whatever is currently on the bar
        echo "i3toast:ers"

    } | i3toast

    ```


<details>
    <summary><b>Full usage and options</b></summary>

```
Usage: i3toast [OPTIONS]

Send messages to i3title.

OPTIONS:
  -m MESSAGE    Display MESSAGE.
  -f FORMAT     Format the message. The format should contain one %s.
  -o OFFSET     Add OFFSET spaces before the message.
  -T TRIM       Trim the message to TRIM characters.
  -t TIMEOUT    Keep the message for TIMEOUT (default: 4 seconds).
                Use -1 to keep it until explicitly erased or overwritten.
  -i            Mark the message as important. Important messages are queued
                and displayed one at a time. Transient (unimportant) messages
                cannot interrupt them while they are running.
  -R            Send the message in raw mode.
  -w            Print the configured i3title maximum width.
  -e            Erase the current message.
  -E            Clear the queue and erase the current message.
  -h            Print usage.

STDIN:
  When input is provided on stdin, each line is treated as a separate message.

  You can tag the output like i3toast:<cmd>[:<payload>] to instruct
  i3toast itself mid-stream:

      i3toast:ers           Erase the current message.
      i3toast:off:OFFSET    Set the message offset.
      i3toast:fmt:FORMAT    Set the message format.
      i3toast:trm:TRIM      Set the message trim width.

RAW MODE:
  By default, i3title trims messages to the configured max_width and sanitizes
  them for safe display in i3bar. Raw mode (-R) disables both safeguards and
  passes the message directly to i3bar. Use it only when you have complete
  control over the data and can ensure that it is safe to pass through unchanged.

  Raw mode allows FORMAT to be included directly in the MESSAGE:

    i3toast -R -m '<b>Hello</b>'

  Without -R, the markup is displayed literally; with -R, "Hello" is displayed
  in bold.

  Arbitrary data is (mostly) safe in normal mode because it is sanitized:

    head -c 100 /dev/urandom | i3toast

  But the same data passed through raw mode can immediately crash i3bar:

    head -c 100 /dev/urandom | i3toast -R  # Don't do this

EXAMPLES:
  i3toast -m 'Hello!'
  echo 'Hello!' | i3toast
  i3toast -i -m 'Important message'
  i3toast -t 10 -m 'Shown for ten seconds'

```

</details>

---
## Acknowledgments

Special thanks to the i3 developers. Their commitment to simplicity, efficiency
and a robust IPC interface is what makes projects like this possible in the first place.

---

## License

GPL-3.0 license. See [LICENSE](LICENSE) for details.
