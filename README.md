<!-- <p align="center">
  <img width="321" height="26" alt="demo" src="https://github.com/user-attachments/assets/acb34afb-f877-4708-aff7-b727314b234e" />
</p> -->

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
<img width="460" height="26" alt="demo.ticker.gif" src="https://github.com/user-attachments/assets/495ee85f-ed26-459e-ac6b-a8722db537ef" />

*Not sure what you'd use it for? [See it in practice ↓](#in-practice)*

<details>
	<summary><strong>video</strong> (Best viewed on desktop in fullscreen)</summary>

https://github.com/user-attachments/assets/7846a025-db57-4143-86a4-0313c08b74e7
</details>


## Features
 - **Zero-Configuration Defaults** – Shows the active window title and i3 mode out of the box.
 - **Fully Customizable** – Accepts arbitrary text via standard input or command-line arguments, letting you display system metrics, weather, API results, custom scripts output, or anything else.
 - **Real-time Updates** – Instantly refreshes on workspace switches, window focus changes, background workspace activity, i3 mode toggles, and new piped data.
 - **Background Workspace Awareness** – Reports on work being done on other workspaces, keeping you informed of background activity without needing to switch views.
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

### From release
```sh
curl -fsSL https://raw.githubusercontent.com/a-pav/i3title/refs/heads/master/scripts/install.sh | sh
```
*(By default, these install the binaries to `~/.local/bin`. Make sure this directory is in your `$PATH`.)*

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
window manager from the command line and scripts. `i3toast` is the same concept,
but for sending messages to `i3title`. (In fact, it was originally named
`i3title-msg` for this reason.)

Its most basic use is sending one-shot, "toast-like" messages from your scripts
that appear on your bar and disappear after a timeout (default is 4 seconds.)

### One-shot messages (Toasts)

- Send a quick, temporary update from a script or keybind — via `-m`:

    ```sh
    # This could be from a volume adjustment script, informing you about the change.
    i3toast -m "🔊 ${volume}%" -f "<span font='bold italic 15'>%s</span>"
    ```

- Or by piping a command's output:
    ```sh
    bindsym $super+w exec curl https://wttr.in/?format=2 | i3toast \
        -f "<span bgcolor='green' fgcolor='black' font='bold'> W:☂️ </span> %s"
    ```
    <img width="450" height="26" alt="demo.i3toast.gif" src="https://github.com/user-attachments/assets/08800cd3-b164-4ff7-9fd6-98a5454731ad" />

### Piping continuous data

- Show the current date and time for 7 seconds, then disappear instantly.

    ```bash
    for i in {1..7}; do
        date '+%A, %-d %B %Y --- %Y-%m-%d %H:%M:%S'
        sleep 1
    done |
        i3toast -t -1,0 -f "<span font='14'>📅</span> <span color='#CFD8DC'>%s</span>"
    ```
    <img width="753" height="25" alt="Screenshot from 2026-08-08 15-08-57" src="https://github.com/user-attachments/assets/1151ff80-aa15-4910-a277-c27271b80e58" />

  - **Pro tip:** Displaying information on-demand like this means you can disable those
corresponding status modules in your i3status config. This frees up space, giving
`i3title` more room to fit longer data and window titles.

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

        ### Tag the output like i3toast:<flag>[:<payload>] to adjust the stream.
        ### Each tag takes effect immediately and applies to subsequent messages:

        echo "i3toast:-f:Now in bold: <b>%s</b>"  # Format
        echo "i3toast:-a:center"                  # Align
        echo "i3toast:-T:3"                       # Ticker mode

        ### See STDIN in usage (i3toast -h) for the flags you can pass as tags.

    } | i3toast

    ```


<details>
    <summary><b>Full usage and options:</b> <code>i3toast -h</code></summary>

```text

NAME
      i3toast - send messages to i3title

SYNOPSIS
      i3toast [OPTIONS]
      STDIN | i3toast [OPTIONS]

OPTIONS
      -m, --message MESSAGE  Display MESSAGE.
      -f, --format FORMAT    Format the message. FORMAT must contain one %s.
      -a, --align ALIGN      Align the message: left, center, or right.
                             Overrides "align" in config.
      -o, --offset OFFSET    Add OFFSET spaces before the message.
      -M, --min-width MINWP  Set the minimum width (in pixels) for the message.
                             Overrides "min_width" in config.
      -W, --max-width MAXWC  Set the maximum width (in characters) for the message.
                             Text longer than this is trimmed.
                             Cannot exceed "max_width" in config.
      -t, --timeout TIMEOUT  Keep the message for TIMEOUT seconds (default: 4).
      -T, --ticker TICK      Enable TICKER mode and set TICK as the minimum display time
	                         (in seconds) per line.
      -R, --raw              Send the message in RAW MODE.
      -i, --important        Mark the message, or stream, as important.
                             Important messages are queued and displayed one at a time.
                             Transient (unimportant) messages cannot interrupt them while
                             they are running. A stream is important until it reaches EOF.
      -w, --widths           Print the configured i3title minimum and maximum widths.
                             Output can be used with eval like: eval $(i3title -w)
      -e, --erase            Erase the current message, as if TIMEOUT had elapsed.
      -E, --clear            Terminate all queued messages and clear the queue (for debugging).
      -h, --help             Print usage.
      -v, --version          Print version.


STDIN
  When input is provided on stdin, each line is treated as a separate message.

  You can tag the output like i3toast:<flag>[:<payload>] to adjust the
  characteristics of the current stream. Each tag takes effect immediately
  and applies to subsequent messages:

      i3toast:-e           Erase the current message.
      i3toast:-f:FORMAT    Set the stream format.
      i3toast:-a:ALIGN     Set the stream alignment.
      i3toast:-o:OFFSET    Set the stream offset.
      i3toast:-M:MINWIDTH  Set the stream minimum width.
      i3toast:-W:MAXWIDTH  Set the stream maximum width.
      i3toast:-T:TICK      Set the stream TICKER mode.

TIMEOUT
  A whole number of seconds (no fractions).

  TIMEOUT may be a single value or two comma-separated values:

      TIMEOUT
      TIMEOUT_MID_STREAM,TIMEOUT_AFTER_EOF

  A single value applies to both cases.

  The first value controls how long intermittent messages remain while the
  stream is still active. The second controls how long the final message
  remains after the input stream reaches EOF.

  Use -1 to keep the message indefinitely until it is explicitly erased
  or overwritten.

  TIMEOUT must be passed as a single argument. Do not include spaces around the
  comma unless you quote the entire argument:

      -t 8,2     # OK
      -t "8, 2"  # OK
      -t 8, 2    # Not OK

  An empty field uses the default timeout:

      -t ,0
          Default timeout while the stream is active, instant timeout after EOF.

      -t 3,
          Timeout after 3 seconds while the stream is active, default timeout
          after EOF.


  Examples:
      -t 4
          Keep the message(s) for 4 seconds.
      -t -1
          Keep the message(s) indefinitely.
      -t 2,30
          Keep the intermittent messages only for 2 seconds, but keep the final
          one for 30 seconds.
      -t -1,0
          Keep intermittent messages indefinitely, but instantly timeout after
          the final one.

TICKER
  -T, --ticker TICK enables ticker mode. TICK sets how long each line rests
  (in seconds, fractions allowed) before it starts rolling to the next. Since
  the stream may have its own pace, each line rests for at least TICK.

  When TICK exceeds TIMEOUT, the message disappears before the next line is shown.

  TICK may be two comma-separated values, TICK,TRANSITION, where TRANSITION
  is how long the roll itself takes (default: 0.1). Quoting rules match TIMEOUT.

  Ticker mode is enabled only when TICK > 0; 0 or an empty first field disables it.

RAW MODE
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

EXAMPLES
      i3toast -m "Hello"
      echo "Hello" | i3toast
      i3toast -i -m "Important message"
      i3toast -t 5 -o 10 -m "Shown for 5 seconds, offseted by 10 spaces"

```

</details>

---
## In Practice

<img width="460" height="26" alt="demo.usecases.gif" src="https://github.com/user-attachments/assets/44cf26c5-1e18-4dba-acf9-22dcf20c7e21" />

`i3title` cycling through different data sources — window title and mode from i3, plus weather and other script outputs. The quick succession is just for the demo; how long each message stays is up to you.

---
## Acknowledgments

Special thanks to the i3 developers. Their commitment to simplicity, efficiency
and a robust IPC interface is what makes projects like this possible in the first place.

---

## License

GPL-3.0 license. See [LICENSE](LICENSE) for details.
