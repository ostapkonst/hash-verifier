package gui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"
)

func ShowError(parent *gtk.Window, title, message string) {
	dialog := gtk.MessageDialogNew(
		parent,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_ERROR,
		gtk.BUTTONS_OK,
		"%s", message,
	)
	defer dialog.Destroy()

	dialog.SetTitle(title)
	dialog.Run()
}

func SelectDirectoryDialog(parent *gtk.Window, title, folder string) (string, bool) {
	dialog, err := gtk.FileChooserDialogNewWith2Buttons(
		title,
		parent,
		gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER,
		"_Open",
		gtk.RESPONSE_ACCEPT,
		"_Cancel",
		gtk.RESPONSE_CANCEL,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create select directory dialog: %w", err))
	}

	defer dialog.Destroy()

	dialog.SetCurrentFolder(folder)

	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		dir := dialog.GetFilename()
		return dir, true
	}

	return "", false
}

func OpenFileDialog(parent *gtk.Window, title, path string) (string, bool) {
	dialog, err := gtk.FileChooserDialogNewWith2Buttons(
		title,
		parent,
		gtk.FILE_CHOOSER_ACTION_SAVE,
		"_Open",
		gtk.RESPONSE_ACCEPT,
		"_Cancel",
		gtk.RESPONSE_CANCEL,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create open file dialog: %w", err))
	}

	defer dialog.Destroy()

	folder, filename := splitPath(path)

	dialog.SetCurrentFolder(folder)
	dialog.SetCurrentName(filename)

	addFileFilters(dialog)

	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		file := dialog.GetFilename()
		return file, true
	}

	return "", false
}

func SaveFileDialog(parent *gtk.Window, title, path, ext string) (string, bool) {
	dialog, err := gtk.FileChooserDialogNewWith2Buttons(
		title,
		parent,
		gtk.FILE_CHOOSER_ACTION_SAVE,
		"_Save",
		gtk.RESPONSE_ACCEPT,
		"_Cancel",
		gtk.RESPONSE_CANCEL,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create save file dialog: %w", err))
	}

	defer dialog.Destroy()

	folder, filename := splitPath(path)

	dialog.SetCurrentFolder(folder)
	dialog.SetCurrentName(filename)

	if dialog.Run() == gtk.RESPONSE_ACCEPT {
		file := dialog.GetFilename()
		return file, true
	}

	return "", false
}

func addFileFilters(dialog *gtk.FileChooserDialog) {
	supportedFiles := [][]string{
		{"CRC-32", "*.sfv"},
		{"MD4", "*.md4"},
		{"MD5", "*.md5"},
		{"SHA1", "*.sha1"},
		{"SHA256", "*.sha256"},
		{"SHA384", "*.sha384"},
		{"SHA512", "*.sha512"},
		{"SHA3-256", "*.sha3-256"},
		{"SHA3-384", "*.sha3-384"},
		{"SHA3-512", "*.sha3-512"},
		{"BLAKE3", "*.blake3"},
	}

	filterAllSupported, _ := gtk.FileFilterNew()
	filterAllSupported.SetName(
		fmt.Sprintf("All Supported Files (%s)", strings.Join(func() []string {
			var result []string
			for _, pattern := range supportedFiles {
				result = append(result, pattern[1])
			}

			return result
		}(), ", ")),
	)

	for _, pattern := range supportedFiles {
		filterAllSupported.AddPattern(pattern[1])
		filterAllSupported.AddPattern(strings.ToUpper(pattern[1]))
	}

	dialog.AddFilter(filterAllSupported)

	for _, pattern := range supportedFiles {
		filter, _ := gtk.FileFilterNew()
		filter.SetName(fmt.Sprintf("%s (%s)", pattern[0], pattern[1]))
		filter.AddPattern(pattern[1])
		filter.AddPattern(strings.ToUpper(pattern[1]))
		dialog.AddFilter(filter)
	}

	filterAny, _ := gtk.FileFilterNew()
	filterAny.SetName("All Files (*.*)")
	filterAny.AddPattern("*")
	dialog.AddFilter(filterAny)

	dialog.SetFilter(filterAllSupported)
}

func changeFileExtension(filename, ext string) string {
	if filename == "" {
		return ""
	}

	return strings.TrimSuffix(filename, filepath.Ext(filename)) + ext
}

func splitPath(fullPath string) (directory, filename string) {
	if fullPath == "" {
		return "", ""
	}

	directory = filepath.Dir(fullPath)
	filename = filepath.Base(fullPath)

	return directory, filename
}

func genChecksumFilename(directory, ext string) string {
	return directory + ext
}
