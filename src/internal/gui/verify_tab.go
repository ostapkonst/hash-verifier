package gui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/inhies/go-bytesize"
	"github.com/rs/zerolog/log"

	"github.com/ostapkonst/HashVerifier/internal/action"
	"github.com/ostapkonst/HashVerifier/internal/checksum"
	"github.com/ostapkonst/HashVerifier/internal/settings"
	"github.com/ostapkonst/HashVerifier/utils/unwrap"
)

type VerifyTab struct {
	builder *gtk.Builder
	window  *gtk.Window

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	settings     *settings.Settings
	columnConfig *ColumnConfig

	entryChecksum       *gtk.Entry
	btnStart            *gtk.Button
	btnStop             *gtk.Button
	btnBrowseChk        *gtk.Button
	treeValidate        *gtk.TreeView
	listStore           *gtk.ListStore
	chkBoxVerifyOnOpen  *gtk.CheckButton
	contextMenuProvider *ContextMenuProvider

	labelMatchV      *gtk.Label
	labelMismatchV   *gtk.Label
	labelUnreadableV *gtk.Label
	labelPendingV    *gtk.Label
	labelSpeedV      *gtk.Label

	labelCurrFileV   *gtk.Label
	totalProgress    *gtk.ProgressBar
	currFileProgress *gtk.ProgressBar
}

func NewVerifyTab(ctx context.Context, builder *gtk.Builder, window *gtk.Window, settings *settings.Settings) *VerifyTab {
	tab := &VerifyTab{
		builder:      builder,
		window:       window,
		ctx:          ctx,
		settings:     settings,
		columnConfig: NewVerifyColumnConfig(),
	}

	tab.getWidgets()
	tab.getLabels()

	tab.contextMenuProvider = NewContextMenuProvider(tab.treeValidate, tab.listStore)

	tab.applySettingsToUI()
	tab.setStartState()

	tab.setupHandlers()

	return tab
}

func (t *VerifyTab) Fill(path string) {
	t.entryChecksum.SetText(path)

	if t.chkBoxVerifyOnOpen.GetActive() {
		t.onStart()
	}
}

func (t *VerifyTab) getWidgets() {
	t.entryChecksum = getEntry(t.builder, "entry_val_checksum")
	t.btnStart = getButton(t.builder, "btn_start_validate")
	t.btnStop = getButton(t.builder, "btn_stop_validate")
	t.btnBrowseChk = getButton(t.builder, "btn_browse_val_checksum")
	t.treeValidate = getTreeView(t.builder, "tree_validate")
	t.listStore = getListStore(t.builder, "liststore_validate")
	t.chkBoxVerifyOnOpen = getCheckButton(t.builder, "chk_val_verify_on_open")

	t.totalProgress = getProgressBar(t.builder, "progress_val_total")
	t.currFileProgress = getProgressBar(t.builder, "progress_val_curr_file")
}

func (t *VerifyTab) getLabels() {
	t.labelMatchV = getLabel(t.builder, "label_val_match_value")
	t.labelMismatchV = getLabel(t.builder, "label_val_mismatch_value")
	t.labelUnreadableV = getLabel(t.builder, "label_val_unreadable_value")
	t.labelPendingV = getLabel(t.builder, "label_val_pending_value")
	t.labelSpeedV = getLabel(t.builder, "label_val_speed_value")

	t.labelCurrFileV = getLabel(t.builder, "label_val_curr_file_value")
}

func (t *VerifyTab) setupHandlers() {
	t.btnBrowseChk.Connect("clicked", func() {
		path, _ := t.entryChecksum.GetText()
		if file, ok := OpenFileDialog(t.window, "Select Checksum File", path); ok {
			t.entryChecksum.SetText(file)

			if t.chkBoxVerifyOnOpen.GetActive() {
				t.onStart()
			}
		}
	})

	t.btnStart.Connect("clicked", t.onStart)
	t.btnStop.Connect("clicked", t.onStop)

	t.chkBoxVerifyOnOpen.Connect("toggled", func() {
		if err := t.saveSettings(); err != nil {
			log.Error().Err(err).Msg("Failed to save settings")
		}
	})
	t.treeValidate.Connect("columns-changed", func() {
		if err := t.saveSettings(); err != nil {
			log.Error().Err(err).Msg("Failed to save settings")
		}
	})

	t.setupContextMenu()

	columns := t.treeValidate.GetColumns()
	for l := columns; l != nil; l = l.Next() {
		if col, ok := l.Data().(*gtk.TreeViewColumn); ok {
			col.Connect("clicked", func() {
				if err := t.saveSettings(); err != nil {
					log.Error().Err(err).Msg("Failed to save settings")
				}
			})
		}
	}
}

func (t *VerifyTab) onStart() {
	checksumFile, _ := t.entryChecksum.GetText()

	checksumFile = filepath.Clean(checksumFile)

	t.listStore.Clear()
	t.activateStopState()

	ctx, cancel := context.WithCancel(t.ctx)
	t.cancel = cancel

	results, err := action.VerifyChecksumsStreaming(ctx, checksumFile)
	if err != nil {
		ShowError(t.window, "Error", fmt.Sprintf("Failed to start verification: %v", err))
		cancel()
		t.cancel = nil
		t.setStartState()

		return
	}

	log.Info().
		Str("checksum_file", checksumFile).
		Msg("Starting verification")

	t.wg.Add(1)

	var hasError error

	lastState := checksum.VerifierStats{}
	currentIdx := int64(0)

	go func() {
		defer t.wg.Done()

		for res := range results {
			if res.IsProgressUpdate {
				glib.IdleAdd(func() {
					lastState = res.Stats
					t.updateStats(lastState)
				})

				if res.Err != nil {
					hasError = res.Err
					break
				}

				continue
			}

			var colorOfStatus string

			switch res.Result.Status {
			case checksum.HashMatched:
				colorOfStatus = "green"
			case checksum.HashMismatch:
				colorOfStatus = "firebrick1"
			default:
				colorOfStatus = "dark orange"
			}

			glib.IdleAdd(func() {
				currentIdx += 1

				iter := t.listStore.Append()
				_ = t.listStore.SetValue(iter, 0, currentIdx)
				_ = t.listStore.SetValue(iter, 1, res.Result.Path)
				_ = t.listStore.SetValue(iter, 2, bytesize.New(float64(res.Result.ReadBytes)).String())
				_ = t.listStore.SetValue(iter, 3, res.Result.Status.String())
				_ = t.listStore.SetValue(iter, 4, res.Result.ActualHash)
				_ = t.listStore.SetValue(iter, 5, res.Result.ExpectedHash)

				if res.Result.Err != nil {
					_ = t.listStore.SetValue(iter, 6, unwrap.UnwrapAndNormalize(res.Result.Err))
				}

				_ = t.listStore.SetValue(iter, 7, colorOfStatus)
				_ = t.listStore.SetValue(iter, 8, res.Result.ReadBytes)

				lastState = res.Stats
				t.updateStats(lastState)
			})
		}
	}()

	go func() {
		t.wg.Wait()

		func() {
			if hasError != nil {
				if errors.Is(hasError, context.Canceled) {
					log.Warn().Msg("Verification canceled")
					return
				}

				log.Error().Err(hasError).Msg("Failed to verify checksums")
				glib.IdleAdd(func() {
					ShowError(t.window, "Error", fmt.Sprintf("Failed to verify checksums: %v", hasError))
				})

				return
			}

			log.Info().
				Int("matched", lastState.Matched).
				Int("mismatch", lastState.Mismatch).
				Int("unreadable", lastState.Unreadable).
				Int("pending", lastState.Pending()).
				Int("total_files", lastState.TotalFiles).
				Msg("Verification stats")

			log.Info().Msg("Verification completed")
		}()

		glib.IdleAdd(func() {
			cancel()
			t.cancel = nil
			t.setStartState()
		})
	}()
}

func (t *VerifyTab) onStop() {
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *VerifyTab) activateStopState() {
	t.btnStart.SetVisible(false)
	t.btnStop.SetVisible(true)

	t.btnBrowseChk.SetSensitive(false)
	t.entryChecksum.SetSensitive(false)
	t.chkBoxVerifyOnOpen.SetSensitive(false)
}

func (t *VerifyTab) setStartState() {
	t.btnStart.SetVisible(true)
	t.btnStop.SetVisible(false)

	t.btnBrowseChk.SetSensitive(true)
	t.entryChecksum.SetSensitive(true)
	t.chkBoxVerifyOnOpen.SetSensitive(true)
}

func (t *VerifyTab) updateStats(stats checksum.VerifierStats) {
	t.labelMatchV.SetText(fmt.Sprintf("%d of %d files", stats.Matched, stats.TotalFiles))
	t.labelMismatchV.SetText(fmt.Sprintf("%d of %d files", stats.Mismatch, stats.TotalFiles))
	t.labelUnreadableV.SetText(fmt.Sprintf("%d of %d files", stats.Unreadable, stats.TotalFiles))
	t.labelPendingV.SetText(fmt.Sprintf("%d of %d files", stats.Pending(), stats.TotalFiles))
	t.labelSpeedV.SetText(bytesize.New(stats.Speed).String() + "/s")

	t.labelCurrFileV.SetText(stats.CurrentFileOrStatus)
	t.totalProgress.SetFraction(stats.TotalProgress())
	t.currFileProgress.SetFraction(stats.FileHashingProgress)
}

func (t *VerifyTab) Wait() {
	t.wg.Wait()
}

func (t *VerifyTab) applySettingsToUI() {
	if t.settings == nil {
		return
	}

	t.chkBoxVerifyOnOpen.SetActive(t.settings.Verify.VerifyOnOpen)
	t.columnConfig.ApplyColumnOrder(t.treeValidate, t.settings.Verify.ColumnOrder)

	var sortOrder gtk.SortType
	if t.settings.Verify.SortOrder == settings.SortOrderDesc {
		sortOrder = gtk.SORT_DESCENDING
	} else {
		sortOrder = gtk.SORT_ASCENDING
	}

	t.columnConfig.ApplySortState(t.treeValidate, t.settings.Verify.SortColumn, sortOrder)
}

func (t *VerifyTab) saveSettings() error {
	if t.settings == nil ||
		t.window.InDestruction() {
		return nil
	}

	t.settings.Verify.VerifyOnOpen = t.chkBoxVerifyOnOpen.GetActive()
	t.settings.Verify.ColumnOrder = t.columnConfig.GetColumnOrder(t.treeValidate)

	sortColumn, sortOrder := t.columnConfig.GetSortState(t.treeValidate)

	t.settings.Verify.SortColumn = sortColumn
	if sortOrder == gtk.SORT_DESCENDING {
		t.settings.Verify.SortOrder = settings.SortOrderDesc
	} else {
		t.settings.Verify.SortOrder = settings.SortOrderAsc
	}

	return t.settings.Save()
}

func (t *VerifyTab) setupContextMenu() {
	columnLabels := []string{"Idx", "Path", "Size", "Status", "Hash", "Expected Hash", "Note"}

	t.contextMenuProvider.CreateMenu(columnLabels)
	t.contextMenuProvider.ConnectRightClick(func() {
		t.contextMenuProvider.ShowMenu()
	})
}
