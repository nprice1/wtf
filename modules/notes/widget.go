package notes

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gdamore/tcell"
	"github.com/olebedev/config"
	"github.com/rivo/tview"
	"github.com/wtfutil/wtf/cfg"
	"github.com/wtfutil/wtf/logger"
	"github.com/wtfutil/wtf/utils"
	"github.com/wtfutil/wtf/view"
)

// Config is a pointer to the global config object
var Config *config.Config

const HelpText = `
 Keyboard commands for Notes:

   /: Show/hide this help window
   j: Select the next notes file in the list
   k: Select the previous notes file in the list
   n: Create a new notes file
   o: Open the notes file in your editor
   r: Renamte the notes file

   arrow down: Select the next item in the list
   arrow up:   Select the previous item in the list

   ctrl-d: Delete the selected note file

   return: Edit selected item
`

type Widget struct {
	view.TextWidget

	app      *tview.Application
	filePath string
	list     *List
	pages    *tview.Pages
	settings *Settings
}

func NewWidget(app *tview.Application, pages *tview.Pages, settings *Settings) *Widget {
	widget := Widget{
		TextWidget: view.NewTextWidget(app, settings.common),

		app:      app,
		filePath: settings.folder,
		list:     &List{selected: -1},
		pages:    pages,
		settings: settings,
	}

	widget.init()
	widget.View.SetInputCapture(widget.keyboardIntercept)

	return &widget
}

/* -------------------- Exported Functions -------------------- */

func (widget *Widget) Refresh() {
	widget.load()
	widget.display()
}

func (widget *Widget) SetList(newList *List) {
	widget.list = newList
}

/* -------------------- Unexported Functions -------------------- */

func (widget *Widget) init() {
	_, err := cfg.CreateFile(widget.filePath)
	if err != nil {
		panic(err)
	}
}

func (widget *Widget) keyboardIntercept(event *tcell.EventKey) *tcell.EventKey {
	switch string(event.Rune()) {
	case "/":
		widget.showHelp()
		return nil
	case "j":
		// Select the next item down
		widget.list.Next()
		widget.display()
		return nil
	case "k":
		// Select the next item up
		widget.list.Prev()
		widget.display()
		return nil
	case "n":
		// Add a new item
		widget.newItem()
		return nil
	case "o":
		// Open the file
		utils.OpenFile(widget.filePath)
		return nil
	case "r":
		// Rename the file
		widget.renameItem()
		return nil
	}

	switch event.Key() {
	case tcell.KeyCtrlD:
		// Delete the selected item
		widget.deleteFile()
		widget.list.Delete()
		widget.display()
		return nil
	case tcell.KeyDown:
		// Select the next item down
		widget.list.Next()
		widget.display()
		return nil
	case tcell.KeyEnter:
		widget.editItem()
		return nil
	case tcell.KeyUp:
		// Select the next item up
		widget.list.Prev()
		widget.display()
		return nil
	default:
		// Pass it along
		return event
	}
}

// edit opens a modal dialog that permits editing the text of the currently-selected note
func (widget *Widget) editItem() {
	if widget.list.Selected() == nil {
		return
	}

	confDir, _ := cfg.WtfConfigDir()
	filePath := fmt.Sprintf("%s/%s/%s", confDir, widget.filePath, widget.list.Selected().Text)
	runCommand := exec.Command("tab", "vim "+filePath)

	err := runCommand.Start()
	if err != nil {
		logger.Log(fmt.Sprintf("Failed to run command: %v", err))
	}
}

// Loads the todo list from Yaml file
func (widget *Widget) load() {
	confDir, _ := cfg.WtfConfigDir()
	filePath := fmt.Sprintf("%s/%s/", confDir, widget.filePath)

	files, _ := ioutil.ReadDir(filePath)
	for _, file := range files {
		widget.list.Add(file.Name())
	}
}

func (widget *Widget) newItem() {
	form := widget.modalForm("New:", "")

	saveFctn := func() {
		text := form.GetFormItem(0).(*tview.InputField).GetText()

		widget.list.Add(text)
		widget.persistNewFile(text)
		widget.pages.RemovePage("modal")
		widget.app.SetFocus(widget.View)
		widget.display()
	}

	widget.addButtons(form, saveFctn)
	widget.modalFocus(form)
}

func (widget *Widget) renameItem() {
	form := widget.modalForm("Rename:", "")

	saveFctn := func() {
		text := form.GetFormItem(0).(*tview.InputField).GetText()

		widget.renameFile(text)
		widget.list.Update(text)
		widget.pages.RemovePage("modal")
		widget.app.SetFocus(widget.View)
		widget.display()
	}

	widget.addButtons(form, saveFctn)
	widget.modalFocus(form)
}

// persist writes the todo list to Yaml file
func (widget *Widget) persistNewFile(filename string) {
	confDir, _ := cfg.WtfConfigDir()
	filePath := fmt.Sprintf("%s/%s/%s", confDir, widget.filePath, filename)

	_, err := os.Create(filePath)

	if err != nil {
		panic(err)
	}
}

func (widget *Widget) persistFileContent(content string) {
	filename := widget.list.Selected().Text
	confDir, _ := cfg.WtfConfigDir()
	filePath := fmt.Sprintf("%s/%s/%s", confDir, widget.filePath, filename)

	err := ioutil.WriteFile(filePath, []byte(content), 0644)

	if err != nil {
		panic(err)
	}
}

func (widget *Widget) deleteFile() {
	filename := widget.list.Selected().Text
	confDir, _ := cfg.WtfConfigDir()
	filePath := fmt.Sprintf("%s/%s/%s", confDir, widget.filePath, filename)

	err := os.Remove(filePath)

	if err != nil {
		logger.Log(fmt.Sprintf("Error deleting file: %+v", err))
		panic(err)
	}
}

func (widget *Widget) renameFile(text string) {
	filename := widget.list.Selected().Text
	confDir, _ := cfg.WtfConfigDir()
	filePath := fmt.Sprintf("%s/%s/%s", confDir, widget.filePath, filename)
	newFilePath := fmt.Sprintf("%s/%s/%s", confDir, widget.filePath, text)

	err := os.Rename(filePath, newFilePath)

	if err != nil {
		logger.Log(fmt.Sprintf("ERROR: %+v", err))
		panic(err)
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

/* -------------------- Modal Form -------------------- */

func (widget *Widget) addButtons(form *tview.Form, saveFctn func()) {
	widget.addSaveButton(form, saveFctn)
	widget.addCancelButton(form)
}

func (widget *Widget) addCancelButton(form *tview.Form) {
	cancelFn := func() {
		widget.pages.RemovePage("modal")
		widget.app.SetFocus(widget.View)
		widget.display()
	}

	form.AddButton("Cancel", cancelFn)
	form.SetCancelFunc(cancelFn)
}

func (widget *Widget) addSaveButton(form *tview.Form, fctn func()) {
	form.AddButton("Save", fctn)
}

func (widget *Widget) modalFocus(form *tview.Form) {
	frame := widget.modalFrame(form)
	widget.modalFrameFocus(frame)
}

func (widget *Widget) modalFrameFocus(frame *tview.Frame) {
	widget.pages.AddPage("modal", frame, false, true)
	widget.app.SetFocus(frame)
}

func (widget *Widget) modalForm(lbl, text string) *tview.Form {
	form := tview.NewForm().
		SetButtonsAlign(tview.AlignCenter).
		SetButtonTextColor(tview.Styles.PrimaryTextColor)

	form.AddInputField(lbl, text, 60, nil, nil)

	return form
}

func (widget *Widget) modalFrame(form *tview.Form) *tview.Frame {
	_, _, w, h := widget.View.GetInnerRect()

	frame := tview.NewFrame(form).SetBorders(0, 0, 0, 0, 0, 0)
	frame.SetBorder(true)
	frame.SetRect(w+20, h+2, 80, 7)

	return frame
}
