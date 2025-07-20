package main

import (
	"fmt"
	"image/color"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MainApp holds the main application structure, including the app, window, and file processor
type MainApp struct {
	App         fyne.App
	Window      fyne.Window
	Processor   *RenamerProcessor
	StatusLabel *widget.Label
	ThemeButton *widget.Button
	DarkMode    bool

	// File selection and filtering
	FilePath              *PathDisplay
	FilterEntry           *widget.Entry
	FileList              *widget.List
	FileListContainer     *container.Scroll
	PreviewTable          *widget.Table
	PreviewTableContainer *container.Scroll
	// Radio groups for operations
	PrefixRadio    *widget.RadioGroup
	SuffixRadio    *widget.RadioGroup
	ExtensionRadio *widget.RadioGroup
	// Perfix, Suffix, and Extension entries
	PrefixEntry    *widget.Entry
	SuffixEntry    *widget.Entry
	ExtensionEntry *widget.Entry
	// Containers for operations
	PrefixContainer    *fyne.Container
	SuffixContainer    *fyne.Container
	ExtensionContainer *fyne.Container

	FolderPath *PathDisplay // Display the folder path
}

// PathDisplay shows the file or folder path in a scrollable text container
type PathDisplay struct {
	Text      *canvas.Text
	Container *container.Scroll
}

// InitializeApp holds the application and window instances along with a file processor
func InitializeApp(app fyne.App, window fyne.Window) *MainApp {
	isDark := app.Preferences().BoolWithFallback("dark_mode", false) // Check if dark mode is enabled in preferences
	return &MainApp{
		App:       app,
		Window:    window,
		Processor: &RenamerProcessor{},
		DarkMode:  isDark, // Save the dark mode preference
	}
}

// Sets up the UI for the application
func (a *MainApp) MakeUI() {

	// Create a rectangle to control the minimum size of the window
	bg := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	bg.SetMinSize(fyne.NewSize(600, 500))

	// Set the theme based on the dark mode preference when the app starts
	a.SetTheme(a.DarkMode)
	// Add theme control button, refreshes the theme when clicked
	// The button's style is based on the current theme
	if a.DarkMode {
		a.ThemeButton = widget.NewButton("☀️", a.ToggleTheme)
	} else {
		a.ThemeButton = widget.NewButton("🌙", a.ToggleTheme)
	}

	// Create about button
	aboutButton := widget.NewButton("About", func() { a.ShowAbout(a.Window) })

	// Set the title of the app
	title := widget.NewLabel("<Batch Renamer>")
	// Title and theme button layout
	TitleContainer := container.NewHBox(
		title,
		layout.NewSpacer(),
		aboutButton,
		a.ThemeButton,
	)

	// Create a scrollable container for the file path
	a.FolderPath = CreatePathDisplay(a.Window)
	a.FolderPath.RefreshColor(a.DarkMode)
	a.FolderPath.UpdatePathDisplayWidth(a.Window)
	// Create a horizontal box for the filter entry
	filterLabel := widget.NewLabel("Filter Extension:")
	a.FilterEntry = widget.NewEntry()
	a.FilterEntry.SetPlaceHolder("e.g. .txt;.jpg (leave empty for all)")
	a.FilterEntry.OnChanged = func(desiredExtension string) {
		a.Processor.FilterExt = desiredExtension // Update the filter extension in the processor
		a.FilterFiles()                          // Call the filter method when the entry changes
	}
	filterBox := container.NewHBox(filterLabel, a.FilterEntry)

	// Create operations area
	// Create a label for operations
	operationsLabel := widget.NewLabel("Operations:")

	// Create a radio group for prefix operations
	prefixLabel := widget.NewLabel("Prefix:")
	a.PrefixRadio = widget.NewRadioGroup([]string{"None", "Add", "Remove"}, func(selected string) {
		a.Processor.PrefixMode = selected
		if selected == "None" {
			a.PrefixEntry.Hide() // Hide the entry if "None" is selected
		} else {
			a.PrefixEntry.Show() // Show the entry for adding/removing prefix
		}
		a.PrefixContainer.Refresh()
	})
	a.PrefixRadio.Horizontal = false  // Make the radio buttons vertical
	a.PrefixRadio.SetSelected("None") // Default to "None"
	a.PrefixEntry = widget.NewEntry() // Entry for prefix
	// Set container for the prefix operations
	a.PrefixContainer = container.NewVBox(
		container.NewHBox(prefixLabel, a.PrefixRadio),
		a.PrefixEntry,
	)

	// Create a radio group for suffix operations
	suffixLabel := widget.NewLabel("Suffix:")
	a.SuffixRadio = widget.NewRadioGroup([]string{"None", "Add", "Remove"}, func(selected string) {
		a.Processor.SuffixMode = selected
		if selected == "None" {
			a.SuffixEntry.Hide() // Hide the entry if "None" is selected
		} else {
			a.SuffixEntry.Show() // Show the entry for adding/removing suffix
		}
		a.SuffixContainer.Refresh()
	})
	a.SuffixRadio.Horizontal = true
	a.SuffixRadio.SetSelected("None")
	a.SuffixEntry = widget.NewEntry()
	// Set container for the suffix operations
	a.SuffixContainer = container.NewVBox(
		container.NewHBox(suffixLabel, a.SuffixRadio),
		a.SuffixEntry,
	)

	// Create a horizontal box for the new extension
	extLabel := widget.NewLabel("Extension:")
	a.ExtensionRadio = widget.NewRadioGroup([]string{"None", "Change"}, func(selected string) {
		a.Processor.ExtensionMode = selected
		if selected == "Change" {
			a.ExtensionEntry.Show()
		} else {
			a.ExtensionEntry.Hide()
		}
		a.ExtensionContainer.Refresh()
	})
	a.ExtensionRadio.Horizontal = true
	a.ExtensionRadio.SetSelected("None")
	a.ExtensionEntry = widget.NewEntry()
	a.ExtensionEntry.SetPlaceHolder("e.g. txt (without dot)")
	// Set container for the extension operations
	a.ExtensionContainer = container.NewVBox(
		container.NewHBox(extLabel, a.ExtensionRadio),
		a.ExtensionEntry,
	)

	// Combine all operation boxes into a vertical box
	operationsBox := container.NewVBox(
		operationsLabel,
		a.PrefixContainer,
		a.SuffixContainer,
		a.ExtensionContainer,
	)

	// Create a list to display the files in the folder
	a.FileList = widget.NewList(
		func() int {
			if a.Processor == nil {
				return 0
			}
			return len(a.Processor.FilteredFiles)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if a.Processor != nil && i < len(a.Processor.FilteredFiles) {
				o.(*widget.Label).SetText(a.Processor.FilteredFiles[i].Name())
			}
		},
	)
	a.FileListContainer = container.NewScroll(a.FileList)
	a.FileListContainer.SetMinSize(fyne.NewSize(300, 300))

	// Create a table to display the file information
	a.PreviewTable = widget.NewTable(
		func() (int, int) {
			if a.Processor == nil || a.Processor.NewNames == nil {
				return 0, 2
			}
			return len(a.Processor.NewNames), 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(tid widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if tid.Col == 0 {
				if tid.Row < len(a.Processor.FilteredFiles) {
					label.SetText(a.Processor.FilteredFiles[tid.Row].Name())
				}
			} else {
				if tid.Row < len(a.Processor.NewNames) {
					label.SetText(a.Processor.NewNames[tid.Row])
				}
			}
		},
	)
	// Adjust table size
	a.PreviewTable.SetColumnWidth(0, 320)
	a.PreviewTable.SetColumnWidth(1, 320)
	a.PreviewTableContainer = container.NewScroll(a.PreviewTable)
	a.PreviewTableContainer.SetMinSize(fyne.NewSize(300, 300))

	// Combine the file list and preview table into a horizontal box
	listsContainer := container.NewHBox(
		container.NewBorder(
			widget.NewLabel("Original Files:"),
			nil, nil, nil,
			a.FileListContainer,
		),
		container.NewBorder(
			widget.NewLabel("Preview:"),
			nil, nil, nil,
			a.PreviewTableContainer,
		),
	)

	// Create status Lables
	a.StatusLabel = widget.NewLabel("Ready")
	a.StatusLabel.Wrapping = fyne.TextWrapWord

	// Create buttons
	folderButton := widget.NewButton("Select Folder", a.SelectFolder)
	previewButton := widget.NewButton("Preview", a.PreviewChanges)
	renameButton := widget.NewButton("Rename Files", a.RenameFiles)
	clearButton := widget.NewButton("Clear", a.ClearAll)
	exitButton := widget.NewButton("Exit", func() { a.App.Quit() })
	// Create a horizontal box for the buttons
	buttonRow := container.NewHBox(
		folderButton,
		previewButton,
		renameButton,
		layout.NewSpacer(),
		clearButton,
		exitButton,
	)

	// Create the main content layout
	contentContainer := container.NewBorder(
		container.NewVBox(
			TitleContainer,
			widget.NewSeparator(),
			container.NewVBox(
				container.NewHBox(widget.NewLabel("Folder:"), a.FolderPath.Container),
				filterBox,
				widget.NewSeparator(),
				operationsBox,
				widget.NewSeparator(),
			),
			buttonRow,
		),
		a.StatusLabel,
		nil,
		nil,
		listsContainer,
	)

	fullWindow := container.New(
		layout.NewStackLayout(),
		bg,
		contentContainer,
	)

	// Set the content
	a.Window.SetContent(fullWindow)

	// Update PathDisplays' width based on window size
	go func() {
		lastSize := a.Window.Canvas().Size()
		for {
			currentSize := a.Window.Canvas().Size()
			if currentSize != lastSize {
				a.FilePath.UpdatePathDisplayWidth(a.Window)
				lastSize = currentSize
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// FilterFiles filters the files based on the specified extension
func (a *MainApp) FilterFiles() {
	a.Processor.FilterFiles()
	a.FileList.Refresh()
	a.StatusLabel.SetText(fmt.Sprintf("Filtered: %d files", len(a.Processor.FilteredFiles)))
}

// Use canvas to display file paths
func CreatePathDisplay(window fyne.Window) *PathDisplay {
	// Set text first
	text := canvas.NewText("No Selection", color.Black)
	text.TextSize = 14
	text.TextStyle = fyne.TextStyle{Monospace: false, Bold: true}
	// Create a scrollable container for the text
	scroll := container.NewHScroll(text)
	// Get width of the window
	windowWidth := window.Canvas().Size().Width
	// Set min size for labels and add scrolls
	minWidth := float32(350)
	// Calculate target width
	targetWidth := windowWidth * 0.85
	scrollLength := max(targetWidth, minWidth)
	scroll.SetMinSize(fyne.NewSize(scrollLength, 45))
	return &PathDisplay{
		Text:      text,
		Container: scroll,
	}
}

// Refreshes PathDisplay's text color based on the theme
func (pd *PathDisplay) RefreshColor(isDark bool) {
	if isDark {
		pd.Text.Color = color.White // Use White for dark theme
	} else {
		pd.Text.Color = color.Black // Use Black for light theme
	}
	pd.Text.Refresh()
}

// Select a folder and update the PathDisplay
func (a *MainApp) SelectFolder() {
	dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			a.StatusLabel.SetText("Error: " + err.Error())
			return
		}
		if list == nil {
			return
		}

		path := list.Path()
		if runtime.GOOS == "windows" {
			if len(path) > 2 && path[0] == '/' && path[2] == ':' {
				path = path[1:]
			}
			path = strings.ReplaceAll(path, "/", "\\")
		}

		if err := a.Processor.LoadFiles(path); err != nil {
			a.StatusLabel.SetText("Error loading files: " + err.Error())
			return
		}

		a.FolderPath.Text.Text = path
		a.FolderPath.Text.Refresh()
		a.FileList.Refresh()
		a.StatusLabel.SetText(fmt.Sprintf("Loaded %d files", len(a.Processor.Files)))
	}, a.Window).Show()
}

// Generate new names for the files and update the preview table
func (a *MainApp) PreviewChanges() {
	if a.Processor.FolderPath == "" {
		a.StatusLabel.SetText("Select a folder first!")
		return
	}
	// Get the values from the entries
	a.Processor.PrefixValue = a.PrefixEntry.Text
	a.Processor.SuffixValue = a.SuffixEntry.Text
	a.Processor.ExtensionValue = a.ExtensionEntry.Text

	a.Processor.GenerateNewNames()
	a.PreviewTable.Refresh()
	a.StatusLabel.SetText("Preview generated")
}

// Rename files based on the generated new names
func (a *MainApp) RenameFiles() {
	if a.Processor.FolderPath == "" {
		a.StatusLabel.SetText("Select a folder first!")
		return
	}

	if a.Processor.NewNames == nil {
		a.StatusLabel.SetText("Preview changes first!")
		return
	}

	successCount, err := a.Processor.RenameFiles()
	if err != nil {
		a.StatusLabel.SetText("Error: " + err.Error())
		return
	}
	a.FileList.Refresh()
	a.PreviewTable.Refresh()
	a.StatusLabel.SetText(fmt.Sprintf("Successfully renamed %d files", successCount))
}

// Clear all content in the table
func (a *MainApp) ClearAll() {
	//Reset RenamerProcessor
	a.Processor = &RenamerProcessor{
		PrefixMode:    "None",
		SuffixMode:    "None",
		ExtensionMode: "None",
	}
	// Reset PathDisplay
	a.FolderPath.Text.Text = "No Folder Selected"
	a.FolderPath.Text.Refresh()
	a.FilterEntry.SetText("")
	// Reset radio buttons
	a.PrefixRadio.SetSelected("None")
	a.SuffixRadio.SetSelected("None")
	a.ExtensionRadio.SetSelected("None")
	// Reset entries
	a.PrefixEntry.SetText("")
	a.PrefixEntry.Hide()
	a.SuffixEntry.SetText("")
	a.SuffixEntry.Hide()
	a.ExtensionEntry.SetText("")
	a.ExtensionEntry.Hide()
	// Reset containers
	a.PrefixContainer.Refresh()
	a.SuffixContainer.Refresh()
	a.ExtensionContainer.Refresh()
	a.FileList.Refresh()
	a.PreviewTable.Refresh()
	// Reset scrollbar of table container
	a.ResetTableScroll()
	// Update status
	a.StatusLabel.SetText("Cleared all settings and selections")
	// Cleanup ram
	a.Cleanup()
	// Reset Processor
	a.FileList.Refresh()
	a.PreviewTable.Refresh()
}

// Reset scrollbar of PathDisplay
func (a *MainApp) ResetPathScroll() {
	if a.FilePath != nil {
		a.FilePath.Container.Offset = fyne.Position{X: 0, Y: 0}
		a.FilePath.Container.Refresh()
	}
}

// Reset scrollbar of table
func (a *MainApp) ResetTableScroll() {
	if a.PreviewTableContainer != nil {
		a.PreviewTableContainer.ScrollToTop()
		a.PreviewTableContainer.Offset = fyne.Position{X: 0, Y: 0}
		a.PreviewTableContainer.Refresh()
	}
}

// Update PathDisplay width based on the window size
func (pd *PathDisplay) UpdatePathDisplayWidth(window fyne.Window) {
	winWidth := window.Canvas().Size().Width
	minWidth := float32(300)
	targetWidth := winWidth * 0.8
	targetWidth = max(minWidth, targetWidth)
	pd.Container.SetMinSize(fyne.NewSize(targetWidth, 45))
}

// Toggle the theme between light and dark mode
func (a *MainApp) SetTheme(isDark bool) {
	a.DarkMode = isDark
	// Save the theme preference
	// Set Theme based on the button state
	a.App.Preferences().SetBool("dark_mode", isDark)
	if isDark {
		a.App.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.App.Settings().SetTheme(theme.LightTheme())
	}
}

// Toggle the theme when the button is clicked
func (a *MainApp) ToggleTheme() {
	a.SetTheme(!a.DarkMode)
	time.Sleep(150 * time.Millisecond)
	if a.DarkMode {
		a.ThemeButton.SetText("☀️") // Show sun icon if dark mode is enabled
	} else {
		a.ThemeButton.SetText("🌙") // Show moon icon if dark mode is disabled
	}
	// Update PathDisplays's colors
	a.FilePath.RefreshColor(a.DarkMode)
	runtime.GC() // Cleanup ram
	// Refresh window
	time.Sleep(100 * time.Millisecond)
	a.Window.Content().Refresh()
	runtime.GC() // Cleanup ram
}

// Cleanup ram
func (a *MainApp) Cleanup() {
	a.Processor = nil
	a.PreviewTable = nil
	runtime.GC()
}

// Show copyright
func (a *MainApp) ShowAbout(win fyne.Window) {
	aboutContent := `Batch-Renamer v1.0.0

© 2025 Cuculus Band
Licensed under the GNU GPL v3.0
Source: https://github.com/CuculusBand/Batch-Renamer
Uses Fyne GUI toolkit (© 2018-present The Fyne Authors) under BSD-3-Clause license`
	dialog.ShowInformation("About", aboutContent, win)
}
