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

	"github.com/ostapkonst/hash-verifier/internal/action"
	"github.com/ostapkonst/hash-verifier/internal/checksum"
	"github.com/ostapkonst/hash-verifier/utils/unwrap"
)

type VerifyTab struct {
	builder *gtk.Builder
	window  *gtk.Window

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	entryChecksum *gtk.Entry
	btnStart      *gtk.Button
	btnStop       *gtk.Button
	btnBrowseChk  *gtk.Button
	listStore     *gtk.ListStore

	labelMatchV      *gtk.Label
	labelMismatchV   *gtk.Label
	labelUnreadableV *gtk.Label
	labelPendingV    *gtk.Label

	labelCurrFileV   *gtk.Label
	totalProgress    *gtk.ProgressBar
	currFileProgress *gtk.ProgressBar
}

func NewVerifyTab(ctx context.Context, builder *gtk.Builder, window *gtk.Window) *VerifyTab {
	tab := &VerifyTab{
		builder: builder,
		window:  window,
		ctx:     ctx,
	}

	tab.getWidgets()
	tab.getLabels()

	tab.setupHandlers()
	tab.setStartState()

	return tab
}

func (t *VerifyTab) Fill(path string) {
	t.entryChecksum.SetText(path)
}

func (t *VerifyTab) getWidgets() {
	t.entryChecksum = getEntry(t.builder, "entry_val_checksum")
	t.btnStart = getButton(t.builder, "btn_start_validate")
	t.btnStop = getButton(t.builder, "btn_stop_validate")
	t.btnBrowseChk = getButton(t.builder, "btn_browse_val_checksum")
	t.listStore = getListStore(t.builder, "liststore_validate")

	t.totalProgress = getProgressBar(t.builder, "progress_val_total")
	t.currFileProgress = getProgressBar(t.builder, "progress_val_curr_file")
}

func (t *VerifyTab) getLabels() {
	t.labelMatchV = getLabel(t.builder, "label_val_match_value")
	t.labelMismatchV = getLabel(t.builder, "label_val_mismatch_value")
	t.labelUnreadableV = getLabel(t.builder, "label_val_unreadable_value")
	t.labelPendingV = getLabel(t.builder, "label_val_pending_value")

	t.labelCurrFileV = getLabel(t.builder, "label_val_curr_file_value")
}

func (t *VerifyTab) setupHandlers() {
	t.btnBrowseChk.Connect("clicked", func() {
		path, _ := t.entryChecksum.GetText()
		if file, ok := OpenFileDialog(t.window, "Select Checksum File", path); ok {
			t.entryChecksum.SetText(file)
		}
	})

	t.btnStart.Connect("clicked", t.onStart)
	t.btnStop.Connect("clicked", t.onStop)
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

	go func() {
		defer t.wg.Done()

		var hasError error

		for res := range results {
			if res.IsProgressUpdate {
				glib.IdleAdd(func() {
					t.updateStats(res.Stats)
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
				iter := t.listStore.Append()
				_ = t.listStore.SetValue(iter, 0, res.Result.Path)
				_ = t.listStore.SetValue(iter, 1, bytesize.New(float64(res.Result.ReadBytes)).String())
				_ = t.listStore.SetValue(iter, 2, res.Result.Status.String())
				_ = t.listStore.SetValue(iter, 3, res.Result.ActualHash)
				_ = t.listStore.SetValue(iter, 4, res.Result.ExpectedHash)

				_ = t.listStore.SetValue(iter, 5, colorOfStatus)
				if res.Result.Err != nil {
					_ = t.listStore.SetValue(iter, 6, unwrap.UnwrapAndNormalize(res.Result.Err))
				}

				t.updateStats(res.Stats)
			})
		}

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

		log.Info().Msg("Verification completed")
	}()

	go func() {
		t.wg.Wait()
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
	t.btnStart.SetSensitive(false)
	t.btnStop.SetSensitive(true)

	t.btnBrowseChk.SetSensitive(false)
	t.entryChecksum.SetSensitive(false)
}

func (t *VerifyTab) setStartState() {
	t.btnStart.SetSensitive(true)
	t.btnStop.SetSensitive(false)

	t.btnBrowseChk.SetSensitive(true)
	t.entryChecksum.SetSensitive(true)
}

func (t *VerifyTab) updateStats(stats checksum.VerifierStats) {
	t.labelMatchV.SetText(fmt.Sprintf("%d of %d files", stats.Matched, stats.TotalFiles))
	t.labelMismatchV.SetText(fmt.Sprintf("%d of %d files", stats.Mismatch, stats.TotalFiles))
	t.labelUnreadableV.SetText(fmt.Sprintf("%d of %d files", stats.Unreadable, stats.TotalFiles))
	t.labelPendingV.SetText(fmt.Sprintf("%d of %d files", stats.Pending(), stats.TotalFiles))

	t.labelCurrFileV.SetText(stats.CurrentFileOrStatus)
	t.totalProgress.SetFraction(stats.TotalProgress())
	t.currFileProgress.SetFraction(stats.FileHashingProgress)
}

func (t *VerifyTab) Wait() {
	t.wg.Wait()
}
