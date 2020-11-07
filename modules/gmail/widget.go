package gmail

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gdamore/tcell"
	"github.com/olebedev/config"
	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/logger"
	"github.com/wtfutil/wtf/utils"
	"github.com/wtfutil/wtf/view"
)

// Config is a pointer to the global config object
var Config *config.Config

const HelpText = `
  Keyboard commands for Gmail:

    /: Show/hide this help window
	r: Refresh the data
	a: Archive the message
	d: Move the message to the trash
	,:  Previous message
	.: Next message
	
	enter: Open currently selected message using Chrome
`

type Widget struct {
	view.TextWidget

	app      *tview.Application
	pages    *tview.Pages
	settings *Settings

	Client *GmailClient

	Messages []*GmailMessage
	Idx      int
}

func NewWidget(app *tview.Application, pages *tview.Pages, settings *Settings) *Widget {
	widget := Widget{
		TextWidget: view.NewTextWidget(app, settings.common),

		Client:   NewClient(settings),
		app:      app,
		pages:    pages,
		settings: settings,
		Idx:      0,
	}

	widget.Messages, _ = widget.Client.Fetch()

	widget.View.SetScrollable(true)
	widget.View.SetInputCapture(widget.keyboardIntercept)

	return &widget
}

/* -------------------- Exported Functions -------------------- */

func (widget *Widget) Refresh() {
	widget.View.SetText("Updating...")

	widget.Messages, _ = widget.Client.Fetch()

	if len(widget.Messages) > 0 {
		widget.Idx = len(widget.Messages) - 1
		widget.showNotification()
	} else {
		widget.Idx = 0
	}

	widget.display()
	widget.deleteOldMessages()
}

func (widget *Widget) Next() {
	widget.Idx = widget.Idx + 1
	if widget.Idx == len(widget.Messages) {
		widget.Idx = 0
	}

	widget.display()
}

func (widget *Widget) Prev() {
	widget.Idx = widget.Idx - 1
	if widget.Idx < 0 {
		widget.Idx = len(widget.Messages) - 1
	}

	widget.display()
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) currentMessage() *GmailMessage {
	if len(widget.Messages) == 0 {
		return nil
	}

	return widget.Messages[widget.Idx]
}

func (widget *Widget) keyboardIntercept(event *tcell.EventKey) *tcell.EventKey {
	switch string(event.Rune()) {
	case "/":
		widget.showHelp()
		return nil
	case "r":
		widget.Refresh()
		return nil
	case "a":
		widget.archiveMessage(widget.currentMessage())
		widget.Refresh()
		return nil
	case "d":
		widget.deleteMessage(widget.currentMessage())
		widget.Refresh()
		return nil
	case ".":
		widget.Prev()
		return nil
	case ",":
		widget.Next()
		return nil
	}

	switch event.Key() {
	case tcell.KeyEnter:
		widget.openMessage()
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

func (widget *Widget) openMessage() {
	currentMessage := widget.currentMessage()
	runCommand := exec.Command("/usr/bin/open", "-a", "/Applications/Google Chrome.app", fmt.Sprintf("https://mail.google.com/mail/#inbox/%s", currentMessage.id))

	err := runCommand.Start()
	if err != nil {
		fmt.Println(fmt.Sprintf("\n[red]FAILED TO OPEN MAIL %v", err))
	}
}

func (widget *Widget) showNotification() {
	currentMessage := widget.currentMessage()

	if currentMessage == nil {
		return
	}

	alertCommand := exec.Command("alert", currentMessage.snippet, currentMessage.from)

	err := alertCommand.Start()
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to run alert command: %v", err))
	}
}

func (widget *Widget) getMessageDir() string {
	folder, _ := utils.ExpandHomeDir(widget.settings.messageFolder)
	return folder
}

func (widget *Widget) archiveMessage(message *GmailMessage) {
	filePath := widget.getMessageDir() + message.id + ".html"
	err := os.Remove(filePath)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to delete message file: %v", err))
	}
	widget.Client.Archive(widget.currentMessage())
}

func (widget *Widget) deleteMessage(message *GmailMessage) {
	filePath := widget.getMessageDir() + message.id + ".html"
	err := os.Remove(filePath)
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to delete message file: %v", err))
	}
	widget.Client.Trash(widget.currentMessage())
}

func (widget *Widget) deleteOldMessages() {
	dirRead, err := os.Open(widget.getMessageDir())
	if err != nil {
		logger.Log(fmt.Sprintf("Error opening directory: %v", err))
	}

	dirFiles, err := dirRead.Readdir(0)
	if err != nil {
		logger.Log(fmt.Sprintf("Error getting all files: %v", err))
	}

	expectedFiles := make(map[string]struct{}, len(widget.Messages))
	for _, message := range widget.Messages {
		expectedFiles[message.id+".html"] = struct{}{}
	}

	// Loop over the directory's files.
	for index := range dirFiles {
		file := dirFiles[index]

		// Get name of file and its full path.
		fileName := file.Name()
		_, contains := expectedFiles[fileName]
		if contains {
			continue
		}
		fullPath := widget.getMessageDir() + fileName

		// Remove the file.
		os.Remove(fullPath)
		logger.Log(fmt.Sprintf("Removed file: %s", fullPath))
	}
}
