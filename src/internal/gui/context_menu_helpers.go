package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

type ContextMenuProvider struct {
	treeView  *gtk.TreeView
	listStore *gtk.ListStore
	menu      *gtk.Menu
}

func NewContextMenuProvider(treeView *gtk.TreeView, listStore *gtk.ListStore) *ContextMenuProvider {
	return &ContextMenuProvider{
		treeView:  treeView,
		listStore: listStore,
	}
}

func (p *ContextMenuProvider) ConnectRightClick(onShowMenu func()) {
	p.treeView.Connect("button-press-event", func(_ *gtk.TreeView, event *gdk.Event) bool {
		eventButton := gdk.EventButtonNewFromEvent(event)
		if eventButton.Button() != 3 {
			return false
		}

		path, _, _, _, ok := p.treeView.GetPathAtPos(int(eventButton.X()), int(eventButton.Y()))
		if !ok {
			return false
		}

		selection, err := p.treeView.GetSelection()
		if err != nil {
			return false
		}

		selection.SelectPath(path)

		if onShowMenu != nil {
			onShowMenu()
		}

		return true
	})
}

func (p *ContextMenuProvider) CreateMenu(columnLabels []string) {
	menu, _ := gtk.MenuNew()

	for i, label := range columnLabels {
		copyItem, _ := gtk.MenuItemNewWithLabel(fmt.Sprintf("Copy %s", label))
		copyItem.Connect("activate", func() {
			p.copyColumnValue(i)
		})
		menu.Append(copyItem)
	}

	menu.ShowAll()

	p.menu = menu
}

func (p *ContextMenuProvider) ShowMenu() {
	if p.menu == nil {
		return
	}

	p.menu.PopupAtPointer(nil)
}

func (p *ContextMenuProvider) copyColumnValue(colIndex int) {
	rowData, ok := getSelectedRowData(p.treeView, p.listStore)
	if !ok {
		return
	}

	if value, exists := rowData[colIndex]; exists {
		_ = copyToClipboard(value)
	}
}

func copyToClipboard(text string) error {
	clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if err != nil {
		return err
	}

	clipboard.SetText(text)
	clipboard.Store()

	return nil
}

func getSelectedRowData(treeView *gtk.TreeView, listStore *gtk.ListStore) (map[int]string, bool) {
	selection, err := treeView.GetSelection()
	if err != nil {
		return nil, false
	}

	_, iter, ok := selection.GetSelected()
	if !ok {
		return nil, false
	}

	columns := listStore.GetNColumns()
	rowData := make(map[int]string, columns)

	for i := range columns {
		value, err := listStore.GetValue(iter, i)
		if err != nil {
			continue
		}

		goVal, err := value.GoValue()
		if err != nil {
			continue
		}

		switch v := goVal.(type) {
		case string:
			rowData[i] = v
		case int, int64, uint, uint64:
			rowData[i] = fmt.Sprintf("%d", v)
		case float64:
			rowData[i] = fmt.Sprintf("%g", v)
		default:
			rowData[i] = fmt.Sprintf("%v", v)
		}
	}

	return rowData, true
}
