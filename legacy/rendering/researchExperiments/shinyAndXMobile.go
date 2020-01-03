package main

//import (
//	"fmt"
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
//		for {
//			switch e := w.NextEvent().(type) {
//			case lifecycle.Event:
//				ctx := e.DrawContext
//				fmt.Printf("CTX = %#v\n", ctx)
//				glCtx := ctx.(gl.Context)
//				glCtx.ClearColor(1, 0, 0, 1)
//				glCtx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
//
//				if e.To == lifecycle.StageDead {
//					fmt.Printf("END\n")
//					return
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
