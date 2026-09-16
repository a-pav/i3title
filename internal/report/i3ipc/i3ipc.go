package i3ipc

import (
	"fmt"

	"github.com/a-pav/i3title/internal/bbuf"

	"go.i3wm.org/i3/v4"
)

// Subscribe to title and mode events via i3 Go package.
func Subscribe(titleCh, modeCh chan<- []byte, i3Done <-chan struct{}, errCh chan<- error) {
	var recvr *i3.EventReceiver
	if modeCh == nil {
		recvr = i3.Subscribe(i3.WindowEventType) // Only window events
	} else {
		recvr = i3.Subscribe(i3.WindowEventType, i3.ModeEventType)
	}

	for recvr.Next() {
		switch ev := recvr.Event().(type) {
		case *i3.WindowEvent:
			switch ev.Change {
			case "title", "focus":
				titleCh <- bbuf.AsBytes(ev.Container.WindowProperties.Title)
				<-i3Done
			}
		case *i3.ModeEvent:
			modeCh <- bbuf.AsBytes(ev.Change)
			<-i3Done
		}
	}

	errCh <- fmt.Errorf("no more i3 events: %v", recvr.Close())
}
