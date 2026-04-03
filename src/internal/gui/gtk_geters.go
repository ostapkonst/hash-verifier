package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func getMainIcon() (*gdk.Pixbuf, error) {
	uiContent, err := assets.ReadFile(uiFavIcon)
	if err != nil {
		return nil, fmt.Errorf("failed to read UI file: %w", err)
	}

	pixbuf, err := gdk.PixbufNewFromDataOnly(uiContent)
	if err != nil {
		return nil, fmt.Errorf("failed to create pixbuf: %w", err)
	}

	return pixbuf, nil
}

func getMainForm() (*gtk.Builder, error) {
	builder, err := gtk.BuilderNew()
	if err != nil {
		return nil, fmt.Errorf("failed to create builder: %w", err)
	}

	uiContent, err := assets.ReadFile(uiMain)
	if err != nil {
		return nil, fmt.Errorf("failed to read UI file: %w", err)
	}

	if err := builder.AddFromString(string(uiContent)); err != nil {
		return nil, fmt.Errorf("failed to parse UI: %w", err)
	}

	return builder, nil
}

func getMainWindow(builder *gtk.Builder) (*gtk.Window, error) {
	obj, err := builder.GetObject("main_window")
	if err != nil {
		return nil, fmt.Errorf("failed to get main window: %w", err)
	}

	window, ok := obj.(*gtk.Window)
	if !ok {
		return nil, fmt.Errorf("object is not a GtkWindow")
	}

	return window, nil
}

func getButton(builder *gtk.Builder, id string) *gtk.Button {
	button, err := func() (*gtk.Button, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get button %s: %w", id, err)
		}

		button, ok := obj.(*gtk.Button)
		if !ok {
			return nil, fmt.Errorf("object %s is not a Button", id)
		}

		return button, nil
	}()
	if err != nil {
		panic(err)
	}

	return button
}

func getEntry(builder *gtk.Builder, id string) *gtk.Entry {
	entry, err := func() (*gtk.Entry, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get entry %s: %w", id, err)
		}

		entry, ok := obj.(*gtk.Entry)
		if !ok {
			return nil, fmt.Errorf("object %s is not an Entry", id)
		}

		return entry, nil
	}()
	if err != nil {
		panic(err)
	}

	return entry
}

func getListStore(builder *gtk.Builder, id string) *gtk.ListStore {
	listStore, err := func() (*gtk.ListStore, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get list store %s: %w", id, err)
		}

		store, ok := obj.(*gtk.ListStore)
		if !ok {
			return nil, fmt.Errorf("object %s is not a ListStore", id)
		}

		return store, nil
	}()
	if err != nil {
		panic(err)
	}

	return listStore
}

func getComboBoxText(builder *gtk.Builder, id string) *gtk.ComboBoxText {
	comboBox, err := func() (*gtk.ComboBoxText, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get combo box text %s: %w", id, err)
		}

		combo, ok := obj.(*gtk.ComboBoxText)
		if !ok {
			return nil, fmt.Errorf("object %s is not a ComboBoxText", id)
		}

		return combo, nil
	}()
	if err != nil {
		panic(err)
	}

	return comboBox
}

func getLabel(builder *gtk.Builder, id string) *gtk.Label {
	label, err := func() (*gtk.Label, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get label %s: %w", id, err)
		}

		label, ok := obj.(*gtk.Label)
		if !ok {
			return nil, fmt.Errorf("object %s is not a Label", id)
		}

		return label, nil
	}()
	if err != nil {
		panic(err)
	}

	return label
}

func getNotebook(builder *gtk.Builder, id string) *gtk.Notebook {
	notebook, err := func() (*gtk.Notebook, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get notebook %s: %w", id, err)
		}

		notebook, ok := obj.(*gtk.Notebook)
		if !ok {
			return nil, fmt.Errorf("object %s is not a Notebook", id)
		}

		return notebook, nil
	}()
	if err != nil {
		panic(err)
	}

	return notebook
}

func getProgressBar(builder *gtk.Builder, id string) *gtk.ProgressBar {
	bar, err := func() (*gtk.ProgressBar, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get progress bar %s: %w", id, err)
		}

		bar, ok := obj.(*gtk.ProgressBar)
		if !ok {
			return nil, fmt.Errorf("object %s is not a ProgressBar", id)
		}

		return bar, nil
	}()
	if err != nil {
		panic(err)
	}

	return bar
}

func getCheckButton(builder *gtk.Builder, id string) *gtk.CheckButton {
	button, err := func() (*gtk.CheckButton, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get check button %s: %w", id, err)
		}

		button, ok := obj.(*gtk.CheckButton)
		if !ok {
			return nil, fmt.Errorf("object %s is not a CheckButton", id)
		}

		return button, nil
	}()
	if err != nil {
		panic(err)
	}

	return button
}

func getTreeView(builder *gtk.Builder, id string) *gtk.TreeView {
	tree, err := func() (*gtk.TreeView, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get tree view %s: %w", id, err)
		}

		treeView, ok := obj.(*gtk.TreeView)
		if !ok {
			return nil, fmt.Errorf("object %s is not a TreeView", id)
		}

		return treeView, nil
	}()
	if err != nil {
		panic(err)
	}

	return tree
}

func getGrid(builder *gtk.Builder, id string) *gtk.Grid {
	grid, err := func() (*gtk.Grid, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get grid %s: %w", id, err)
		}

		grid, ok := obj.(*gtk.Grid)
		if !ok {
			return nil, fmt.Errorf("object %s is not a Grid", id)
		}

		return grid, nil
	}()
	if err != nil {
		panic(err)
	}

	return grid
}
