package gui

import (
	"context"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
	window       *gtk.Window
	builder      *gtk.Builder
	generateTab  *GenerateTab
	verifyTab    *VerifyTab
	icon         *gdk.Pixbuf
	ctx          context.Context
	settings     *settings.Settings
	notebook     *gtk.Notebook
	windowState  settings.WindowState
	normalWidth  int // размер окна в нормальном состоянии
	normalHeight int
	normalX      int // позиция окна в нормальном состоянии
	normalY      int
}

func Run(path string) error {
	gtk.Init(nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := &App{ctx: ctx}

	if err := app.initUI(); err != nil {
		return fmt.Errorf("failed to initialize UI: %w", err)
	}

	app.window.Show()
	app.showFlatpakWarningIfNeeded()

	app.fillTabAndSwitch(path)

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

	a.windowState = a.settings.Window.WindowState

	a.generateTab = NewGenerateTab(a.ctx, a.builder, a.window, a.settings)
	a.verifyTab = NewVerifyTab(a.ctx, a.builder, a.window, a.settings)

	a.notebook = getNotebook(a.builder, "notebook")

	a.applyTabOrder()
	a.applyCurrentPage()
	a.restoreWindowGeometry()

	a.connectTabReorderHandler()
	a.connectTabSwitchHandler()
	a.connectWindowEvents()
	a.connectDragAndDropHandler()

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

func (a *App) restoreWindowGeometry() {
	if a.settings.Window.RestoreMode == settings.RestoreModeDefault {
		return
	}

	if (a.settings.Window.RestoreMode == settings.RestoreModeSize ||
		a.settings.Window.RestoreMode == settings.RestoreModeAll) &&
		(a.settings.Window.Width > 0 && a.settings.Window.Height > 0) {
		a.window.Resize(a.settings.Window.Width, a.settings.Window.Height)
		a.normalWidth = a.settings.Window.Width
		a.normalHeight = a.settings.Window.Height
	}

	if (a.settings.Window.RestoreMode == settings.RestoreModePosition ||
		a.settings.Window.RestoreMode == settings.RestoreModeAll) &&
		(a.settings.Window.X >= 0 || a.settings.Window.Y >= 0) {
		a.window.Move(a.settings.Window.X, a.settings.Window.Y)
		a.normalX = a.settings.Window.X
		a.normalY = a.settings.Window.Y
	}

	if a.settings.Window.RestoreMode == settings.RestoreModeSize ||
		a.settings.Window.RestoreMode == settings.RestoreModeAll {
		switch a.settings.Window.WindowState {
		case settings.WindowStateMaximized:
			a.window.Maximize()
		case settings.WindowStateFullscreen:
			a.window.Fullscreen()
		}
	}
}

func (a *App) saveWindowGeometry() {
	if a.windowState == settings.WindowStateNormal {
		width, height := a.window.GetSize()
		x, y := a.window.GetPosition()

		a.normalWidth = width
		a.normalHeight = height

		a.normalX = x
		a.normalY = y
	}

	if a.settings.Window.RestoreMode == settings.RestoreModeSize ||
		a.settings.Window.RestoreMode == settings.RestoreModeAll {
		a.settings.Window.Width = a.normalWidth
		a.settings.Window.Height = a.normalHeight

		state := a.windowState
		a.settings.Window.WindowState = state
	}

	if a.settings.Window.RestoreMode == settings.RestoreModePosition ||
		a.settings.Window.RestoreMode == settings.RestoreModeAll {
		a.settings.Window.X = a.normalX
		a.settings.Window.Y = a.normalY
	}

	if err := a.settings.Save(); err != nil {
		log.Error().Err(err).Msg("Failed to save window geometry")
	}
}

func (a *App) connectWindowEvents() {
	a.window.Connect("delete-event", func() {
		a.saveWindowGeometry()
	})

	a.window.Connect("window-state-event", func(_ *gtk.Window, event *gdk.Event) {
		wsEvent := gdk.EventWindowStateNewFromEvent(event)
		if wsEvent == nil {
			return
		}

		newState := wsEvent.NewWindowState()

		switch {
		case newState&gdk.WINDOW_STATE_FULLSCREEN != 0:
			a.windowState = settings.WindowStateFullscreen
		case newState&gdk.WINDOW_STATE_MAXIMIZED != 0:
			a.windowState = settings.WindowStateMaximized
		default:
			a.windowState = settings.WindowStateNormal
		}
	})

	a.window.Connect("configure-event", func() {
		if a.windowState == settings.WindowStateNormal {
			width, height := a.window.GetSize()
			x, y := a.window.GetPosition()

			a.normalWidth = width
			a.normalHeight = height

			a.normalX = x
			a.normalY = y
		}
	})
}

func uriToFilePath(uri string) (string, error) {
	uri = strings.TrimRight(strings.TrimSpace(uri), "\r\n")

	if uri == "" {
		return "", fmt.Errorf("empty URI")
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("failed to parse URI: %w", err)
	}

	if parsedURL.Scheme != "file" {
		return "", fmt.Errorf("unsupported URI scheme: %s", parsedURL.Scheme)
	}

	path := parsedURL.Path

	if runtime.GOOS == "windows" {
		if len(path) > 2 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}

		path = filepath.FromSlash(path)
	}

	decodedPath, err := url.PathUnescape(path)
	if err != nil {
		return "", fmt.Errorf("failed to unescape path: %w", err)
	}

	return decodedPath, nil
}

func (a *App) connectDragAndDropHandler() {
	targetEntry, err := gtk.TargetEntryNew("text/uri-list", gtk.TARGET_OTHER_APP, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create drag and drop target entry")
		return
	}

	a.window.DragDestSet(gtk.DEST_DEFAULT_ALL, []gtk.TargetEntry{*targetEntry}, gdk.ACTION_COPY)

	a.window.Connect("drag-data-received", func(
		window *gtk.Window,
		ctx *gdk.DragContext,
		x, y int,
		data *gtk.SelectionData,
		info uint,
		time uint,
	) {
		bytes := data.GetData()
		content := string(bytes)

		lines := strings.Split(strings.TrimSpace(content), "\n")
		if len(lines) == 0 || lines[0] == "" {
			log.Warn().Msg("No valid URIs in drag and drop data")
			return
		}

		filePath, err := uriToFilePath(lines[0])
		if err != nil {
			log.Warn().Msg("Failed to convert URI to file path")
			return
		}

		a.fillTabAndSwitch(filePath)
	})
}
