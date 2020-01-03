package main

//
//import (
//	"fmt"
//
//	"golang.org/x/mobile/event/paint"
//
//	"golang.org/x/mobile/event/size"
//
//	"golang.org/x/exp/shiny/driver/gldriver"
//	"golang.org/x/exp/shiny/screen"
//	"golang.org/x/mobile/event/lifecycle"
//	"golang.org/x/mobile/gl"
//)
//
//func main() {
//	gldriver.Main(func(s screen.Screen) {
//		w, err := s.NewWindow(nil)
//		if err != nil {
//			//handleError(err)
//			return
//		}
//		defer w.Release()
//
//		application := &Application{}
//		for {
//			switch e := w.NextEvent().(type) {
//			case size.Event:
//			case paint.Event:
//				application.Paint()
//			case lifecycle.Event:
//				ctx := e.DrawContext
//				fmt.Printf("CTX = %#v\n", ctx)
//				//glCtx := ctx.(gl.Context)
//				//glCtx.ClearColor(1, 0, 0, 1)
//				//glCtx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
//
//				if e.To == lifecycle.StageDead {
//					fmt.Printf("END\n")
//					return
//				}
//
//				switch e.Crosses(lifecycle.StageVisible) {
//				case lifecycle.CrossOn:
//					glctx, _ := e.DrawContext.(gl.Context)
//					application.Start(glctx)
//					//a.Send(paint.Event{})
//				case lifecycle.CrossOff:
//					application.Stop()
//					w.Release()
//					//os.Exit(0) // TODO: it seems that app.Main runs in an endless loop and isn't supposed to be stopped
//				}
//
//				//etc/
//				//case mouse.Event:
//				//	te := touch.Event{
//				//		X: e.X,
//				//		Y: e.Y,
//				//	}
//				//	switch e.Direction {
//				//	case mouse.DirNone:
//				//		te.Type = touch.TypeMove
//				//	case mouse.DirPress:
//				//		te.Type = touch.TypeBegin
//				//	case mouse.DirRelease:
//				//		te.Type = touch.TypeEnd
//				//	}
//				//	fmt.Printf("%v\n", te)
//				//return te
//				//default:
//				//	fmt.Printf("event: %v\n", e)
//				//case etc:
//				//	etc
//			}
//		}
//	})
//}
