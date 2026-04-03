package app

import (
	"context"
	"fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/rs/zerolog/log"

	"github.com/ostapkonst/HashVerifier/internal/gui/tabs"
	"github.com/ostapkonst/HashVerifier/internal/gui/widgets"
	"github.com/ostapkonst/HashVerifier/internal/settings"
	"github.com/ostapkonst/HashVerifier/utils/gracer"
)

type App struct {
	window       *gtk.Window
	builder      *gtk.Builder
	generateTab  *tabs.GenerateTab
	verifyTab    *tabs.VerifyTab
	icon         *gdk.Pixbuf
	ctx          context.Context
	settings     *settings.Settings
	notebook     *gtk.Notebook
	tabManager   *TabManager
	windowGeom   *WindowGeometry
	pathResolver *PathResolver
	dragAndDrop  *DragAndDrop
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
	pathType, resolvedPath := a.pathResolver.Resolve(path)
	switch pathType {
	case PathTypeDirectory:
		a.generateTab.Fill(resolvedPath)
		a.tabManager.ApplySelectedPage(a.tabManager.GetTabNumberByName("generate"))
	case PathTypeChecksumFile:
		a.verifyTab.Fill(resolvedPath)
		a.tabManager.ApplySelectedPage(a.tabManager.GetTabNumberByName("verify"))
	}
}

func (a *App) initUI() error {
	builder, err := widgets.GetMainForm()
	if err != nil {
		return fmt.Errorf("failed to get main form: %w", err)
	}

	a.builder = builder

	favIcon, err := widgets.GetMainIcon()
	if err != nil {
		return fmt.Errorf("failed to get main icon: %w", err)
	}

	a.icon = favIcon

	window, err := widgets.GetMainWindow(builder)
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

	a.generateTab = tabs.NewGenerateTab(a.ctx, a.builder, a.window, a.settings)
	a.verifyTab = tabs.NewVerifyTab(a.ctx, a.builder, a.window, a.settings)
	a.notebook = widgets.GetNotebook(a.builder, "notebook")
	a.tabManager = NewTabManager(a.notebook, a.window, a.settings)
	a.windowGeom = NewWindowGeometry(a.window, a.settings)
	a.pathResolver = NewPathResolver()
	a.dragAndDrop = NewDragAndDrop(a.window, a.pathResolver, a.fillTabAndSwitch)
	a.tabManager.ApplyTabOrder()
	a.tabManager.ApplyCurrentPage()
	a.windowGeom.Restore()
	a.tabManager.ConnectReorderHandler()
	a.tabManager.ConnectSwitchHandler()
	a.windowGeom.ConnectEvents()
	a.dragAndDrop.Setup()

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
		widgets.ShowAboutDialog(a.window, a.icon)
	})

	return nil
}

func (a *App) showFlatpakWarningIfNeeded() {
	if a.settings.Flatpak.SuppressSandboxWarning {
		return
	}

	if !widgets.IsRunningInFlatpak() {
		return
	}

	suppress := widgets.ShowFlatpakSandboxWarningDialog(a.window)
	if !suppress {
		return
	}

	a.settings.Flatpak.SuppressSandboxWarning = true
	if err := a.settings.Save(); err != nil {
		log.Error().Err(err).Msg("Failed to save Flatpak warning suppression setting")
	}
}
