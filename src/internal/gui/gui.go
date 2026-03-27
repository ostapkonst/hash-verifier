package gui

import (
	"context"
	"embed"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"

	"github.com/ostapkonst/HashVerifier/internal/checksum"
	"github.com/ostapkonst/HashVerifier/internal/settings"
	"github.com/ostapkonst/HashVerifier/utils/gracer"
)

//go:embed glade/*
var assets embed.FS

const (
	uiMain    = "glade/main.glade"
	uiFavIcon = "glade/favicon.ico"
)

type App struct {
	window         *gtk.Window
	builder        *gtk.Builder
	generateTab    *GenerateTab
	verifyTab      *VerifyTab
	icon           *gdk.Pixbuf
	ctx            context.Context
	settings       *settings.Settings
	notebook       *gtk.Notebook
	showDetails    *gtk.ToggleButton
	previousHeight int
}

func Run(path string) error {
	gtk.Init(nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &App{ctx: ctx}

	if err := app.initUI(); err != nil {
		return fmt.Errorf("failed to initialize UI: %w", err)
	}

	app.showFlatpakWarningIfNeeded()
	app.fillTabAndSwitch(path)
	app.window.Show()

	gracer.AddCallback(func() error {
		cancel()

		app.generateTab.Wait()
		app.verifyTab.Wait()

		gtk.MainQuit()

		return nil
	})

	go gtk.Main()

	return gracer.Wait()
}

func (a *App) fillTabAndSwitch(path string) {
	cleanPath := filepath.Clean(path)

	if cleanPath == "." {
		return
	}

	fileInfo, err := os.Stat(cleanPath)
	if err == nil && fileInfo.IsDir() {
		a.generateTab.Fill(cleanPath)
		a.applySelectedPage(a.getTabNumberByName("generate"))

		return
	}

	if _, err = checksum.AlgorithmFromExtension(cleanPath); err != nil {
		a.generateTab.Fill(cleanPath)
		a.applySelectedPage(a.getTabNumberByName("generate"))

		return
	}

	a.verifyTab.Fill(cleanPath)
	a.applySelectedPage(a.getTabNumberByName("verify"))
}

func (a *App) initUI() error {
	builder, err := getMainForm()
	if err != nil {
		return fmt.Errorf("failed to get main form: %w", err)
	}

	a.builder = builder

	favIcon, err := getMainIcon()
	if err != nil {
		return fmt.Errorf("failed to get main icon: %w", err)
	}

	a.icon = favIcon

	window, err := getMainWindow(builder)
	if err != nil {
		return fmt.Errorf("failed to get main window: %w", err)
	}

	a.window = window

	window.SetIcon(favIcon)

	window.Connect("destroy", func() {
		gracer.GracefulShutdown()
	})

	if err := a.connectAboutButton(); err != nil {
		return fmt.Errorf("failed to connect about button: %w", err)
	}

	a.settings, err = settings.Load()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load settings, using defaults")

		a.settings = settings.DefaultSettings()
	}

	a.generateTab = NewGenerateTab(a.ctx, a.builder, a.window, a.settings)
	a.verifyTab = NewVerifyTab(a.ctx, a.builder, a.window, a.settings)

	a.notebook = getNotebook(a.builder, "notebook")
	a.showDetails = getToggleButton(a.builder, "show_details")

	a.applyTabOrder()
	a.applyCurrentPage()
	a.applyShowDetails()

	a.connectTabReorderHandler()
	a.connectTabSwitchHandler()
	a.connectShowDetailsHandler()

	return nil
}

func (a *App) connectAboutButton() error {
	obj, err := a.builder.GetObject("main_about")
	if err != nil {
		return fmt.Errorf("failed to get about button: %w", err)
	}

	menuItem, ok := obj.(*gtk.Button)
	if !ok {
		return fmt.Errorf("object is not a GtkButton")
	}

	menuItem.Connect("clicked", func() {
		ShowAboutDialog(a.window, a.icon)
	})

	return nil
}

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

func getScrolledWindow(builder *gtk.Builder, id string) *gtk.ScrolledWindow {
	scrolled, err := func() (*gtk.ScrolledWindow, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get scrolled window %s: %w", id, err)
		}

		sw, ok := obj.(*gtk.ScrolledWindow)
		if !ok {
			return nil, fmt.Errorf("object %s is not a ScrolledWindow", id)
		}

		return sw, nil
	}()
	if err != nil {
		panic(err)
	}

	return scrolled
}

func getToggleButton(builder *gtk.Builder, id string) *gtk.ToggleButton {
	button, err := func() (*gtk.ToggleButton, error) {
		obj, err := builder.GetObject(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get toggle button %s: %w", id, err)
		}

		btn, ok := obj.(*gtk.ToggleButton)
		if !ok {
			return nil, fmt.Errorf("object %s is not a ToggleButton", id)
		}

		return btn, nil
	}()
	if err != nil {
		panic(err)
	}

	return button
}

func (a *App) getTabOrder() []string {
	var order []string

	nPages := a.notebook.GetNPages()

	for i := 0; i < nPages; i++ {
		child, err := a.notebook.GetNthPage(i)
		if err != nil {
			continue
		}

		widget, ok := child.(*gtk.Box)
		if !ok {
			continue
		}

		name, err := widget.GetName()
		if err == nil && name != "" {
			order = append(order, name)
		}
	}

	return order
}

func (a *App) applyTabOrder() {
	order := a.settings.Window.TabOrder
	if len(order) == 0 {
		return
	}

	nPages := a.notebook.GetNPages()
	pageMap := make(map[string]*gtk.Box)

	for i := 0; i < nPages; i++ {
		child, err := a.notebook.GetNthPage(i)
		if err != nil {
			continue
		}

		widget, ok := child.(*gtk.Box)
		if !ok {
			continue
		}

		name, err := widget.GetName()
		if err == nil && name != "" {
			pageMap[name] = widget
		}
	}

	for i, name := range order {
		if child, ok := pageMap[name]; ok {
			a.notebook.ReorderChild(child, i)
		}
	}
}

func (a *App) connectTabReorderHandler() {
	a.notebook.Connect("page-reordered", func() {
		if a.window.InDestruction() {
			return
		}

		a.settings.Window.TabOrder = a.getTabOrder()
		a.settings.Window.CurrentPage = a.notebook.GetCurrentPage()

		if err := a.settings.Save(); err != nil {
			log.Error().Err(err).Msg("Failed to save tab order")
		}
	})
}

func (a *App) applyCurrentPage() {
	a.applySelectedPage(a.settings.Window.CurrentPage)
}

func (a *App) applySelectedPage(page int) {
	if page >= 0 && page < a.notebook.GetNPages() {
		a.notebook.SetCurrentPage(page)
	}
}

func (a *App) connectTabSwitchHandler() {
	a.notebook.Connect(
		"switch-page",
		func(
			self any,
			page any,
			page_num uint, // пришлось делать так, т. к. GetCurrentPage() возвращает старое значение
		) {
			if a.window.InDestruction() {
				return
			}

			a.settings.Window.CurrentPage = int(page_num)
			if err := a.settings.Save(); err != nil {
				log.Error().Err(err).Msg("Failed to save current page")
			}
		},
	)
}

func (a *App) getTabNumberByName(name string) int {
	nPages := a.notebook.GetNPages()

	for i := 0; i < nPages; i++ {
		child, err := a.notebook.GetNthPage(i)
		if err != nil {
			continue
		}

		widget, ok := child.(*gtk.Box)
		if !ok {
			continue
		}

		widgetName, err := widget.GetName()
		if err == nil && widgetName == name {
			return i
		}
	}

	return -1
}

func (a *App) showFlatpakWarningIfNeeded() {
	if a.settings.Flatpak.SuppressSandboxWarning {
		return
	}

	if !isRunningInFlatpak() {
		return
	}

	suppress := ShowFlatpakSandboxWarningDialog(a.window)

	if !suppress {
		return
	}

	a.settings.Flatpak.SuppressSandboxWarning = true
	if err := a.settings.Save(); err != nil {
		log.Error().Err(err).Msg("Failed to save Flatpak warning suppression setting")
	}
}

func (a *App) applyShowDetails() {
	show := a.settings.Window.ShowDetails
	a.showDetails.SetActive(show)
	a.generateTab.SetDetailsVisible(show)
	a.verifyTab.SetDetailsVisible(show)

	_, currentHeight := a.window.GetSize()
	a.previousHeight = currentHeight

	geometry := gdk.Geometry{}

	if show {
		a.window.SetGeometryHints(nil, geometry, 0)
	} else {
		geometry.SetMinHeight(1)
		geometry.SetMaxHeight(1)
		geometry.SetMinWidth(1)
		geometry.SetMaxWidth(math.MaxInt32)
		a.window.SetGeometryHints(nil, geometry, gdk.HINT_MIN_SIZE|gdk.HINT_MAX_SIZE)
	}
}

func (a *App) connectShowDetailsHandler() {
	a.showDetails.Connect("toggled", func() {
		show := a.showDetails.GetActive()
		a.generateTab.SetDetailsVisible(show)
		a.verifyTab.SetDetailsVisible(show)

		geometry := gdk.Geometry{}

		if show {
			a.window.SetGeometryHints(nil, geometry, 0)
			currentWidth, _ := a.window.GetSize()
			a.window.Resize(currentWidth, a.previousHeight)
		} else {
			_, currentHeight := a.window.GetSize()
			a.previousHeight = currentHeight

			geometry.SetMinHeight(1)
			geometry.SetMaxHeight(1)
			geometry.SetMinWidth(1)
			geometry.SetMaxWidth(math.MaxInt32)
			a.window.SetGeometryHints(nil, geometry, gdk.HINT_MIN_SIZE|gdk.HINT_MAX_SIZE)
		}

		a.settings.Window.ShowDetails = show
		if err := a.settings.Save(); err != nil {
			log.Error().Err(err).Msg("Failed to save show details setting")
		}
	})
}
