package notes

import (
	"fmt"

	"github.com/wtfutil/wtf/utils"
)

const checkWidth = 4

func (widget *Widget) display() {
	str := ""
	newList := List{selected: -1}

	selectedItem := widget.list.Selected()
	maxLineLen := widget.list.LongestLine()

	for _, item := range widget.list.GetItems() {
		str = str + widget.formattedItemLine(item, selectedItem, maxLineLen)
		newList.Items = append(newList.Items, item)
	}

	newList.SetSelectedByItem(widget.list.Selected())
	widget.SetList(&newList)

	widget.View.Clear()
	widget.View.SetText(fmt.Sprintf("%s", str))
}

func (widget *Widget) formattedItemLine(item *Item, selectedItem *Item, maxLen int) string {
	foreColor, backColor := "white", ""

	if widget.View.HasFocus() && (item == selectedItem) {
		foreColor = ""
		backColor = "white"
	}

	str := fmt.Sprintf(
		"[%s:%s] - %s [white]",
		foreColor,
		backColor,
		item.Text,
	)

	str = str + utils.RowPadding((checkWidth+len(item.Text)), (checkWidth+maxLen)) + "\n"

	return str
}
