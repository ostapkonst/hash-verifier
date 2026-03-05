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

type GenerateTab struct {
	builder *gtk.Builder
	window  *gtk.Window

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	entryDir             *gtk.Entry
	btnStart             *gtk.Button
	btnStop              *gtk.Button
	btnBrowseDir         *gtk.Button
	listStore            *gtk.ListStore
	entryChecksum        *gtk.Entry
	btnSaveChk           *gtk.Button
	cmbTxtAlgorithm      *gtk.ComboBoxText
	chkBtnFollowSymlinks *gtk.CheckButton
	chkBtnSortPaths      *gtk.CheckButton

	labelProcessedV  *gtk.Label
	labelWithErrorsV *gtk.Label
	labelPendingV    *gtk.Label
	labelSpeedV      *gtk.Label

	labelCurrFileV   *gtk.Label
	totalProgress    *gtk.ProgressBar
	currFileProgress *gtk.ProgressBar
}

func NewGenerateTab(ctx context.Context, builder *gtk.Builder, window *gtk.Window) *GenerateTab {
	tab := &GenerateTab{
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

func (t *GenerateTab) Fill(path string) {
	t.entryDir.SetText(path)
	extension := t.cmbTxtAlgorithm.GetActiveID()
	t.entryChecksum.SetText(genChecksumFilename(path, extension))
}

func (t *GenerateTab) getWidgets() {
	t.entryDir = getEntry(t.builder, "entry_gen_dir")
	t.btnStart = getButton(t.builder, "btn_start_generate")
	t.btnStop = getButton(t.builder, "btn_stop_generate")
	t.btnBrowseDir = getButton(t.builder, "btn_browse_gen_dir")
	t.listStore = getListStore(t.builder, "liststore_generate")
	t.entryChecksum = getEntry(t.builder, "entry_gen_checksum")
	t.btnSaveChk = getButton(t.builder, "btn_save_gen_checksum")
	t.cmbTxtAlgorithm = getComboBoxText(t.builder, "cmb_gen_algorithm")
	t.chkBtnFollowSymlinks = getCheckButton(t.builder, "chk_gen_follow_symlinks")
	t.chkBtnSortPaths = getCheckButton(t.builder, "chk_gen_sort_paths")

	t.totalProgress = getProgressBar(t.builder, "progress_gen_total")
	t.currFileProgress = getProgressBar(t.builder, "progress_gen_curr_file")
}

func (t *GenerateTab) getLabels() {
	t.labelProcessedV = getLabel(t.builder, "label_gen_processed_value")
	t.labelWithErrorsV = getLabel(t.builder, "label_gen_with_errors_value")
	t.labelPendingV = getLabel(t.builder, "label_gen_pending_value")
	t.labelSpeedV = getLabel(t.builder, "label_gen_speed_value")

	t.labelCurrFileV = getLabel(t.builder, "label_gen_curr_file_value")
}

func (t *GenerateTab) setupHandlers() {
	t.btnBrowseDir.Connect("clicked", func() {
		path, _ := t.entryDir.GetText()
		if dir, ok := SelectDirectoryDialog(t.window, "Select Source Directory", path); ok {
			t.entryDir.SetText(dir)

			extension := t.cmbTxtAlgorithm.GetActiveID()
			if checksumPath, _ := t.entryChecksum.GetText(); checksumPath == "" {
				t.entryChecksum.SetText(genChecksumFilename(dir, extension))
			}
		}
	})

	onAlgorithmChanged := func() {
		extension := t.cmbTxtAlgorithm.GetActiveID()
		path, _ := t.entryChecksum.GetText()
		file := changeFileExtension(path, extension)
		t.entryChecksum.SetText(file)
	}

	t.btnSaveChk.Connect("clicked", func() {
		extension := t.cmbTxtAlgorithm.GetActiveID()
		checksumPath, _ := t.entryChecksum.GetText()

		if file, ok := SaveFileDialog(t.window, "Save Checksum File", checksumPath, extension); ok {
			t.entryChecksum.SetText(file)

			if _, err := checksum.AlgorithmFromExtension(file); err != nil {
				onAlgorithmChanged()
			}
		}
	})

	t.entryChecksum.Connect("changed", func() {
		checksumPath, _ := t.entryChecksum.GetText()
		if algo, err := checksum.AlgorithmFromExtension(checksumPath); err == nil {
			t.cmbTxtAlgorithm.SetActiveID(algo.Extension())
		}
	})

	t.entryChecksum.Connect("focus_out_event", func() {
		checksumPath, _ := t.entryChecksum.GetText()
		if _, err := checksum.AlgorithmFromExtension(checksumPath); err != nil {
			onAlgorithmChanged()
		}
	})

	t.btnStart.Connect("clicked", t.onStart)
	t.btnStop.Connect("clicked", t.onStop)
	t.cmbTxtAlgorithm.Connect("changed", onAlgorithmChanged)
}

func (t *GenerateTab) onStart() {
	inputDir, _ := t.entryDir.GetText()
	outputFile, _ := t.entryChecksum.GetText()

	inputDir = filepath.Clean(inputDir)
	outputFile = filepath.Clean(outputFile)

	t.listStore.Clear()
	t.activateStopState()

	ctx, cancel := context.WithCancel(t.ctx)
	t.cancel = cancel

	cfg := action.GenerateStreamingConfig{
		InputDir:            inputDir,
		OutputFile:          outputFile,
		FollowSymbolicLinks: t.chkBtnFollowSymlinks.GetActive(),
		SortPaths:           t.chkBtnSortPaths.GetActive(),
	}

	results, err := action.GenerateChecksumsStreamingToFile(ctx, cfg)
	if err != nil {
		ShowError(t.window, "Error", fmt.Sprintf("Failed to start generation: %v", err))
		cancel()
		t.cancel = nil
		t.setStartState()

		return
	}

	log.Info().
		Str("input_dir", inputDir).
		Str("output_file", outputFile).
		Msg("Starting checksum generation")

	t.wg.Add(1)

	var hasError error

	lastStats := checksum.GeneratorStats{}

	go func() {
		defer t.wg.Done()

		for res := range results {
			if res.IsProgressUpdate {
				glib.IdleAdd(func() {
					lastStats = res.Stats
					t.updateStats(lastStats)
				})

				if res.Err != nil {
					hasError = res.Err
					break
				}

				continue
			}

			glib.IdleAdd(func() {
				iter := t.listStore.Append()
				_ = t.listStore.SetValue(iter, 0, res.Result.RelPath)
				_ = t.listStore.SetValue(iter, 1, bytesize.New(float64(res.Result.ReadBytes)).String())

				_ = t.listStore.SetValue(iter, 2, res.Result.Hash)
				if res.Result.Err != nil {
					_ = t.listStore.SetValue(iter, 3, unwrap.UnwrapAndNormalize(res.Result.Err))
				}

				lastStats = res.Stats
				t.updateStats(lastStats)
			})
		}
	}()

	go func() {
		t.wg.Wait()

		func() {
			if hasError != nil {
				if errors.Is(hasError, context.Canceled) {
					log.Warn().Msg("Checksum generation canceled")
					return
				}

				log.Error().Err(hasError).Msg("Failed to generate checksums")
				glib.IdleAdd(func() {
					ShowError(t.window, "Error", fmt.Sprintf("Failed to generate checksums: %v", hasError))
				})

				return
			}

			log.Info().
				Int("processed", lastStats.Processed).
				Int("pending", lastStats.Pending()).
				Int("with_errors", lastStats.WithErrors).
				Int("total_files", lastStats.TotalFiles).
				Msg("Checksum generation stats")

			log.Info().
				Str("file", outputFile).
				Msg("Checksum generation completed")
		}()

		glib.IdleAdd(func() {
			t.cancel()
			t.setStartState()
			t.cancel = nil
		})
	}()
}

func (t *GenerateTab) onStop() {
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *GenerateTab) activateStopState() {
	t.btnStart.SetSensitive(false)
	t.btnStop.SetSensitive(true)

	t.btnBrowseDir.SetSensitive(false)
	t.btnSaveChk.SetSensitive(false)
	t.entryDir.SetSensitive(false)
	t.entryChecksum.SetSensitive(false)
	t.cmbTxtAlgorithm.SetSensitive(false)
	t.chkBtnFollowSymlinks.SetSensitive(false)
	t.chkBtnSortPaths.SetSensitive(false)
}

func (t *GenerateTab) setStartState() {
	t.btnStart.SetSensitive(true)
	t.btnStop.SetSensitive(false)

	t.btnBrowseDir.SetSensitive(true)
	t.btnSaveChk.SetSensitive(true)
	t.entryDir.SetSensitive(true)
	t.entryChecksum.SetSensitive(true)
	t.cmbTxtAlgorithm.SetSensitive(true)
	t.chkBtnFollowSymlinks.SetSensitive(true)
	t.chkBtnSortPaths.SetSensitive(true)
}

func (t *GenerateTab) updateStats(stats checksum.GeneratorStats) {
	t.labelProcessedV.SetText(fmt.Sprintf("%d of %d files", stats.Processed, stats.TotalFiles))
	t.labelWithErrorsV.SetText(fmt.Sprintf("%d of %d files", stats.WithErrors, stats.TotalFiles))
	t.labelPendingV.SetText(fmt.Sprintf("%d of %d files", stats.Pending(), stats.TotalFiles))
	t.labelSpeedV.SetText(bytesize.New(stats.Speed).String() + "/s")

	t.labelCurrFileV.SetText(stats.CurrentFileOrStatus)
	t.totalProgress.SetFraction(stats.TotalProgress())
	t.currFileProgress.SetFraction(stats.FileHashingProgress)
}

func (t *GenerateTab) Wait() {
	t.wg.Wait()
}
