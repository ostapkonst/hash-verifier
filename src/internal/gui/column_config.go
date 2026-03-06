package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

type ColumnConfig struct {
	titleToName map[string]string
}

func NewGenerateColumnConfig() *ColumnConfig {
	return &ColumnConfig{
		titleToName: map[string]string{
			"Path": "path",
			"Size": "size",
			"Hash": "hash",
			"Note": "note",
		},
	}
}

func NewVerifyColumnConfig() *ColumnConfig {
	return &ColumnConfig{
		titleToName: map[string]string{
			"Path":          "path",
			"Size":          "size",
			"Status":        "status",
			"Hash":          "hash",
			"Expected Hash": "expected_hash",
			"Note":          "note",
		},
	}
}

func (c *ColumnConfig) GetColumnOrder(treeView *gtk.TreeView) []string {
	columns := treeView.GetColumns()
	result := make([]string, 0)

	for l := columns; l != nil; l = l.Next() {
		col, ok := l.Data().(*gtk.TreeViewColumn)
		if !ok {
			continue
		}

		name := c.getColumnTitle(col)
		if name != "" {
			result = append(result, name)
		}
	}

	return result
}

func (c *ColumnConfig) ApplyColumnOrder(treeView *gtk.TreeView, order []string) {
	if len(order) == 0 {
		return
	}

	columns := treeView.GetColumns()
	columnMap := make(map[string]*gtk.TreeViewColumn)

	for l := columns; l != nil; l = l.Next() {
		col, ok := l.Data().(*gtk.TreeViewColumn)
		if !ok {
			continue
		}

		name := c.getColumnTitle(col)
		if name != "" {
			columnMap[name] = col
		}
	}

	for i := len(order) - 1; i >= 0; i-- {
		name := order[i]
		if col, ok := columnMap[name]; ok {
			treeView.MoveColumnAfter(col, nil)
		}
	}
}

func (c *ColumnConfig) getColumnTitle(col *gtk.TreeViewColumn) string {
	title := col.GetTitle()

	if name, ok := c.titleToName[title]; ok {
		return name
	}

	return ""
}
