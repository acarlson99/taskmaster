package main

import (
	"fmt"
	"time"

	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
)

const (
	lw = 40

	ih = 3
)

func runGocui(procs ProcessMap, p ProcChans) {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		logger.Println("Failed to create a GUI:", err)
		return
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)
	//keybind
	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	if err != nil {
		logger.Println("Could not set key binding:", err)
		return
	}
	err = g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, input)
	if err != nil {
		logger.Println("Cannot bind the enter key:", err)
	}

	tw, th := g.Size()
	//list
	lv, err := g.SetView("list", 0, 0, lw, th-1)

	if err != nil && err != gocui.ErrUnknownView {
		logger.Println("Failed to create main view:", err)
		return
	}
	lv.Title = "List"
	lv.FgColor = gocui.ColorCyan
	//output
	ov, err := g.SetView("output", lw+1, 0, tw-1, th-ih-1)
	if err != nil && err != gocui.ErrUnknownView {
		logger.Println("Failed to create output view:", err)
		return
	}
	ov.Title = "Output"
	ov.FgColor = gocui.ColorGreen

	ov.Autoscroll = true
	_, err = fmt.Fprintln(ov, "Press Ctrl-c to quit")
	if err != nil {
		logger.Println("Failed to print into output view:", err)
	}
	//input
	iv, err := g.SetView("input", lw+1, th-ih, tw-1, th-1)
	if err != nil && err != gocui.ErrUnknownView {
		logger.Println("Failed to create input view:", err)
		return
	}
	iv.Title = "Input"
	iv.FgColor = gocui.ColorYellow

	iv.Editable = true
	err = iv.SetCursor(0, 0)
	if err != nil {
		logger.Println("Failed to set cursor:", err)
		return
	}

	go test(g, &procs)

	_, err = g.SetCurrentView("input")
	if err != nil {
		logger.Println("Cannot set focus to input view:", err)
	}

	err = g.MainLoop()
	logger.Println("Main loop has finished:", err)
}

func test(g *gocui.Gui, procs *ProcessMap) {
	tmp := 0
	for {
		select {
		case <-time.After(500 * time.Millisecond):
			g.Update(func(g *gocui.Gui) error {
				v, err := g.View("list")
				if err != nil {
					return err
				}
				v.Clear()
				fmt.Fprintln(v, procs)
				tmp++
				return nil
			})
		}
	}
}

func input(g *gocui.Gui, v *gocui.View) error {
	iv, e := g.View("input")
	if e != nil {
		logger.Println("Cannot get output view:", e)
		return e
	}

	iv.Rewind()

	ov, e := g.View("output")
	if e != nil {
		logger.Println("Cannot get output view:", e)
		return e
	}

	_, e = fmt.Fprint(ov, iv.Buffer())
	if e != nil {
		logger.Println("Cannot print to output view:", e)
	}

	iv.Clear()

	e = iv.SetCursor(0, 0)
	if e != nil {
		logger.Println("Failed to set cursor:", e)
	}
	return e
}

func layout(g *gocui.Gui) error {

	tw, th := g.Size()

	_, err := g.SetView("list", 0, 0, lw, th-1)
	if err != nil {
		return errors.Wrap(err, "Cannot update list view")
	}
	_, err = g.SetView("output", lw+1, 0, tw-1, th-ih-1)
	if err != nil {
		return errors.Wrap(err, "Cannot update output view")
	}
	_, err = g.SetView("input", lw+1, th-ih, tw-1, th-1)
	if err != nil {
		return errors.Wrap(err, "Cannot update input view.")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
