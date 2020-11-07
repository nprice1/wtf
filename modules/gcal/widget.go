package gcal

import (
	"fmt"
	"os/exec"

	"github.com/gdamore/tcell"
	"github.com/olebedev/config"
	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/logger"
	"github.com/wtfutil/wtf/view"
	"google.golang.org/api/calendar/v3"
)

// Config is a pointer to the global config object
var Config *config.Config

const HelpText = `
  Keyboard commands for Gcal:

    /: Show/hide this help window
	r: Refresh the data
	,: Previous message
	.: Next message

	m: Open the event in meet
`

type Widget struct {
	view.TextWidget

	app      *tview.Application
	pages    *tview.Pages
	settings *Settings

	Events *calendar.Events
	Idx    int
}

func NewWidget(app *tview.Application, pages *tview.Pages, settings *Settings) *Widget {
	widget := Widget{
		TextWidget: view.NewTextWidget(app, settings.common),

		app:      app,
		pages:    pages,
		settings: settings,
		Idx:      0,
	}

	widget.Events, _ = widget.Fetch()
	widget.View.SetScrollable(true)
	widget.View.SetInputCapture(widget.keyboardIntercept)

	return &widget
}

/* -------------------- Exported Functions -------------------- */

func (widget *Widget) Refresh() {
	widget.Events, _ = widget.Fetch()

	widget.display()
}

func (widget *Widget) Next() {
	widget.Idx = widget.Idx + 1
	if widget.Idx == len(widget.Events.Items) {
		widget.Idx = 0
	}

	widget.display()
}

func (widget *Widget) Prev() {
	widget.Idx = widget.Idx - 1
	if widget.Idx < 0 {
		widget.Idx = len(widget.Events.Items) - 1
	}

	widget.display()
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) currentEvent() *calendar.Event {
	if len(widget.Events.Items) == 0 {
		return nil
	}

	if widget.Idx < 0 || widget.Idx >= len(widget.Events.Items) {
		return nil
	}

	return widget.Events.Items[widget.Idx]
}

func (widget *Widget) keyboardIntercept(event *tcell.EventKey) *tcell.EventKey {
	switch string(event.Rune()) {
	case "/":
		widget.showHelp()
		return nil
	case "r":
		widget.Refresh()
		return nil
	case "m":
		widget.openMeet()
		return nil
	case ".":
		widget.Next()
		return nil
	case ",":
		widget.Prev()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEnter:
		widget.openCalendar()
		return nil
	default:
		return event
	}
}

func (widget *Widget) showHelp() {
	closeFunc := func() {
		widget.pages.RemovePage("help")
		widget.app.SetFocus(widget.View)
	}

	modal := view.NewBillboardModal(HelpText, closeFunc)

	widget.pages.AddPage("help", modal, false, true)
	widget.app.SetFocus(modal)
}

func (widget *Widget) openMeet() {
	currentEvent := widget.currentEvent()
	runCommand := exec.Command("/usr/bin/open", "-a", "/Applications/Google Chrome.app", currentEvent.HangoutLink)

	err := runCommand.Start()
	if err != nil {
		fmt.Println(fmt.Sprintf("\n[red]FAILED TO OPEN EVENT %v", err))
	}
}

func (widget *Widget) openCalendar() {
	runCommand := exec.Command("/usr/bin/open", "-a", "/Applications/Google Chrome.app", "https://calendar.google.com")

	err := runCommand.Start()
	if err != nil {
		fmt.Println(fmt.Sprintf("\n[red]FAILED TO OPEN EVENT %v", err))
	}
}

func (widget *Widget) showNotification(event *calendar.Event) {
	runCommand := exec.Command("alert", event.Summary, "Starting soon")

	err := runCommand.Start()
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to run alert command: %v", err))
	}
}
