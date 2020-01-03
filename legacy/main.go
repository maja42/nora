package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"golang.org/x/mobile/event/key"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

var (
	touchX float32
	touchY float32
)

func main() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigc
		buf := make([]byte, 1<<16)
		runtime.Stack(buf, true)
		fmt.Printf("%s", buf)
		os.Exit(0)
	}()

	application := &Application{}
	app.Main(func(a app.App) {
		//var glctx gl.Context
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ := e.DrawContext.(gl.Context)
					application.Start(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					application.Stop()
					os.Exit(0) // TODO: it seems that app.Main runs in an endless loop and isn't supposed to be stopped
				}
			case size.Event:
				sz = e
				touchX = float32(sz.WidthPx / 2)
				touchY = float32(sz.HeightPx / 2)
				application.Resize(sz.WidthPx, sz.HeightPx)
			case paint.Event:
				application.Paint()
				a.Publish()           // swap buffer
				a.Send(paint.Event{}) // trigger next paint event
			case touch.Event:
				touchX = e.X
				touchY = e.Y
			case key.Event:
				if e.Rune == 'q' || e.Code == key.CodeEscape {
					application.Stop()
					os.Exit(0)
				}
				//application.OnKey(e.Rune, e.Code, e.Modifiers, e.Direction)
			}
		}
	})
}
