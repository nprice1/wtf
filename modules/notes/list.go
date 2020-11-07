package notes

type List struct {
	Items []*Item

	selected int
}

/* -------------------- Exported Functions -------------------- */

func (list *List) Add(text string) {
	item := Item{
		Text: text,
	}

	list.Items = append([]*Item{&item}, list.Items...)
}

func (list *List) Delete() {
	list.Items = append(list.Items[:list.selected], list.Items[list.selected+1:]...)
	list.Prev()
}

func (list *List) Next() {
	list.selected = list.selected + 1
	if list.selected >= len(list.Items) {
		list.selected = 0
	}
}

func (list *List) GetItems() []*Item {
	return list.Items
}

func (list *List) LongestLine() int {
	maxLen := 0

	for _, item := range list.Items {
		if len(item.Text) > maxLen {
			maxLen = len(item.Text)
		}
	}

	return maxLen
}

func (list *List) Prev() {
	list.selected = list.selected - 1
	if list.selected < 0 {
		list.selected = len(list.Items) - 1
	}
}

func (list *List) Selected() *Item {
	if list.isUnselectable() {
		return nil
	}

	return list.Items[list.selected]
}

func (list *List) SetSelectedByItem(selectableItem *Item) {
	for idx, item := range list.Items {
		if item == selectableItem {
			list.selected = idx
			break
		}
	}
}

func (list *List) Unselect() {
	list.selected = -1
}

func (list *List) Update(text string) {
	item := list.Selected()

	if item == nil {
		return
	}

	item.Text = text
}

/* -------------------- Sort Interface -------------------- */

func (list *List) Len() int {
	return len(list.Items)
}

func (list *List) Less(i, j int) bool {
	return list.Items[i].Text < list.Items[j].Text
}

func (list *List) Swap(i, j int) {
	list.Items[i], list.Items[j] = list.Items[j], list.Items[i]
}

/* -------------------- Unexported Functions -------------------- */

func (list *List) isSelectable() bool {
	return list.selected >= 0 && list.selected < len(list.Items)
}

func (list *List) isUnselectable() bool {
	return !list.isSelectable()
}
