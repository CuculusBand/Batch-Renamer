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
	FolderPath             *PathDisplay // Display the folder path
	FilterEntry            *widget.Entry
	OriginalTable          *widget.Table
	OriginalTableContainer *container.Scroll
	PreviewTable           *widget.Table
	PreviewTableContainer  *container.Scroll
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
	filterBox := container.NewBorder(nil, nil, filterLabel, nil, container.NewHScroll(a.FilterEntry))

	// Create operations area
	// Create a label for operations
	operationsLabel := widget.NewLabel("Operations:")

	// Entry for prefix
	a.PrefixEntry = widget.NewEntry()
	// Create a radio group for prefix operations
	prefixLabel := widget.NewLabel("Prefix:")
	a.PrefixRadio = widget.NewRadioGroup([]string{"None", "Add", "Remove"}, nil)
	a.PrefixRadio.Horizontal = true // Make the radio buttons horizontal
	// Set container for the prefix operations
	a.PrefixContainer = container.NewVBox(
		container.NewHBox(prefixLabel, a.PrefixRadio),
		a.PrefixEntry,
	)
	// Set the onChanged function for the prefix radio group
	a.PrefixRadio.OnChanged = func(selected string) {
		if selected == "" {
			a.PrefixRadio.SetSelected(a.Processor.PrefixMode)
			return
		} // Avoid situation where selected is empty
		a.Processor.PrefixMode = selected
		if selected == "None" {
			a.PrefixEntry.Hide()      // Hide the entry if "None" is selected
			a.PrefixEntry.SetText("") // Clear the entry text
			a.Processor.PrefixValue = ""
		} else {
			a.PrefixEntry.Show()                         // Show the entry for adding/removing prefix
			a.Processor.PrefixValue = a.PrefixEntry.Text // Update the prefix value in the processor
		}
		a.PrefixContainer.Refresh()
	}
	// Set the default selection for prefix radio group
	a.PrefixRadio.SetSelected("None")
	// Update value when prefix entry changes
	a.PrefixEntry.OnChanged = func(value string) {
		a.Processor.PrefixValue = value
	}

	// Entry for suffix
	a.SuffixEntry = widget.NewEntry()
	// Create a radio group for suffix operations
	suffixLabel := widget.NewLabel("Suffix:")
	// Create a radio group for suffix operations
	a.SuffixRadio = widget.NewRadioGroup([]string{"None", "Add", "Remove"}, nil)
	a.SuffixRadio.Horizontal = true // Make the radio buttons horizontal
	// Set container for the suffix operations
	a.SuffixContainer = container.NewVBox(
		container.NewHBox(suffixLabel, a.SuffixRadio),
		a.SuffixEntry,
	)
	// Set the onChanged function for the suffix radio group
	a.SuffixRadio.OnChanged = func(selected string) {
		if selected == "" {
			a.SuffixRadio.SetSelected(a.Processor.SuffixMode)
			return
		} // Avoid situation where selected is empty
		a.Processor.SuffixMode = selected
		if selected == "None" {
			a.SuffixEntry.Hide()      // Hide the entry if "None" is selected
			a.SuffixEntry.SetText("") // Clear the entry text
			a.Processor.SuffixValue = ""
		} else {
			a.SuffixEntry.Show()                         // Show the entry for adding/removing suffix
			a.Processor.SuffixValue = a.SuffixEntry.Text // Update the suffix value in the processor
		}
		if a.SuffixContainer != nil {
			a.SuffixContainer.Refresh()
		}
	}
	// Set the default selection for suffix radio group
	a.SuffixRadio.SetSelected("None")
	// Update value when suffix entry changes
	a.SuffixEntry.OnChanged = func(value string) {
		a.Processor.SuffixValue = value
	}

	// Entry for extension
	a.ExtensionEntry = widget.NewEntry()
	// Create a horizontal box for the new extension
	extLabel := widget.NewLabel("Extension:")
	// Create a radio group for extension operations
	a.ExtensionRadio = widget.NewRadioGroup([]string{"None", "Change"}, nil)
	a.ExtensionRadio.Horizontal = true // Make the radio buttons horizontal
	// Set container for the extension operations
	a.ExtensionContainer = container.NewVBox(
		container.NewHBox(extLabel, a.ExtensionRadio),
		a.ExtensionEntry,
	)
	// Set the onChanged function for the extension radio group
	a.ExtensionRadio.OnChanged = func(selected string) {
		if selected == "" {
			a.ExtensionRadio.SetSelected(a.Processor.ExtensionMode)
			return
		} // Avoid situation where selected is empty
		a.Processor.ExtensionMode = selected
		if selected == "Change" {
			a.ExtensionEntry.Show()      // Show the entry for changing extension
			a.ExtensionEntry.SetText("") // Clear the entry text
			a.Processor.ExtensionValue = ""
		} else {
			a.ExtensionEntry.Hide()
			a.Processor.ExtensionValue = a.ExtensionEntry.Text // Update the extension value in the processor
		}
		if a.ExtensionContainer != nil {
			a.ExtensionContainer.Refresh()
		}
	}
	// Set the default selection for suffix radio group
	a.ExtensionRadio.SetSelected("None")
	a.ExtensionEntry.SetPlaceHolder("e.g. txt (without dot)")
	// Update value when extension entry changes
	a.ExtensionEntry.OnChanged = func(value string) {
		a.Processor.ExtensionValue = value
	}

	// Combine all operation boxes into a vertical box
	operationsBox := container.NewVBox(
		operationsLabel,
		a.PrefixContainer,
		a.SuffixContainer,
		a.ExtensionContainer,
	)

	// Create a table to display the original files
	a.OriginalTable = widget.NewTable(
		func() (rows int, cols int) {
			if a.Processor == nil || a.Processor.FilteredFiles == nil {
				return 0, 1
			}
			return len(a.Processor.FilteredFiles), 1
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(tid widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if a.Processor != nil && a.Processor.FilteredFiles != nil && tid.Row < len(a.Processor.FilteredFiles) {
				label.SetText(a.Processor.FilteredFiles[tid.Row].Name())
			}
		},
	)

	// Create a table to display the processed files
	a.PreviewTable = widget.NewTable(
		func() (rows int, cols int) {
			if a.Processor == nil || a.Processor.NewNames == nil {
				return 0, 1
			}
			return len(a.Processor.NewNames), 1
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(tid widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if a.Processor != nil && a.Processor.NewNames != nil && tid.Row < len(a.Processor.NewNames) {
				label.SetText(a.Processor.NewNames[tid.Row])
			}
		},
	)

	// Create container, Set the column width for both tables
	a.OriginalTable = a.InitializeTable()
	a.OriginalTable.SetColumnWidth(0, 300)
	a.OriginalTableContainer = container.NewScroll(a.OriginalTable)
	a.OriginalTableContainer.SetMinSize(fyne.NewSize(300, 400))
	a.PreviewTable = a.InitializeTable()
	a.PreviewTable.SetColumnWidth(0, 300)
	a.PreviewTableContainer = container.NewScroll(a.PreviewTable)
	a.PreviewTableContainer.SetMinSize(fyne.NewSize(300, 400))
	a.ResetTableScroll()

	// Combine the original table and preview table into a horizontal box
	listsContainer := container.NewHBox(
		container.NewBorder(
			widget.NewLabel("Original Files:"),
			nil, nil, nil,
			a.OriginalTableContainer,
		),
		container.NewBorder(
			widget.NewLabel("Preview New Names:"),
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
	renameButton := widget.NewButton("Rename Files", a.RunRenameProcess)
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
				// Update FolderPath width
				//a.FolderPath.UpdatePathDisplayWidth(a.Window)
				lastSize = currentSize
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// FilterFiles filters the files based on the specified extension
func (a *MainApp) FilterFiles() {
	a.Processor.FilterFiles()
	a.OriginalTable.Refresh()
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
		a.OriginalTable.Refresh()
		a.OriginalTableContainer.Refresh()
		a.StatusLabel.SetText(fmt.Sprintf("Loaded %d files", len(a.Processor.Files)))
	}, a.Window).Show()
}

// Generate new names for the files and update the preview table
func (a *MainApp) PreviewChanges() {
	if a.Processor.FolderPath == "" {
		a.StatusLabel.SetText("Select a folder first!")
		return
	}
	// Check if there are any files to rename
	if len(a.Processor.FilteredFiles) == 0 {
		a.StatusLabel.SetText("No files to rename!")
		return
	}
	// Get the values from the entries
	a.Processor.PrefixValue = a.PrefixEntry.Text
	a.Processor.SuffixValue = a.SuffixEntry.Text
	a.Processor.ExtensionValue = a.ExtensionEntry.Text

	a.Processor.GenerateNewNames()
	a.PreviewTable.Refresh()
	a.PreviewTableContainer.Refresh()
	a.StatusLabel.SetText(fmt.Sprintf("Preview generated, %d files", len(a.Processor.NewNames)))
}

// Rename files based on the generated new names
func (a *MainApp) RunRenameProcess() {
	if a.Processor.FolderPath == "" {
		a.StatusLabel.SetText("Select a folder first!")
		return
	}

	if a.Processor.NewNames == nil {
		a.StatusLabel.SetText("Generate preview first!")
		return
	}

	successCount, err := a.Processor.RenameFiles()
	if err != nil {
		a.StatusLabel.SetText("Error: " + err.Error())
		return
	}
	a.OriginalTable.Refresh()
	newPreviewTable := a.InitializeTable()
	a.PreviewTable = newPreviewTable
	a.PreviewTableContainer.Content = newPreviewTable
	a.PreviewTableContainer.Refresh()
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
	a.OriginalTable = a.InitializeTable()
	a.PreviewTable = a.InitializeTable()
	a.OriginalTable.Refresh()
	a.PreviewTable.Refresh()
	a.OriginalTableContainer.Content = a.OriginalTable
	a.PreviewTableContainer.Content = a.PreviewTable
	a.OriginalTableContainer.Refresh()
	a.PreviewTableContainer.Refresh()
	a.PrefixContainer.Refresh()
	a.SuffixContainer.Refresh()
	a.ExtensionContainer.Refresh()
	// Reset scrollbar of table container
	//a.ResetTableScroll()
	// Update status
	a.StatusLabel.SetText("Cleared all settings and selections")
	// Cleanup ram
	a.Cleanup()
}

// Reset scrollbar of PathDisplay
func (a *MainApp) ResetPathScroll() {
	if a.FolderPath != nil {
		a.FolderPath.Container.Offset = fyne.Position{X: 0, Y: 0}
		a.FolderPath.Container.Refresh()
	}
}

// Reset scrollbar of table
func (a *MainApp) ResetTableScroll() {
	if a.OriginalTableContainer != nil {
		a.OriginalTableContainer.ScrollToTop()
		a.OriginalTableContainer.Offset = fyne.Position{X: 0, Y: 0}
		a.OriginalTableContainer.Refresh()
	}
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

// Create table
func (a *MainApp) InitializeTable() *widget.Table {
	return widget.NewTable(
		func() (int, int) {
			if a.Processor == nil || len(a.Processor.NewNames) == 0 {
				return 0, 1 // Check data in the Processor
			}
			return len(a.Processor.NewNames), 1
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if a.Processor != nil &&
				len(a.Processor.NewNames) > i.Row {
				label.SetText(a.Processor.NewNames[i.Row])
			} else {
				label.SetText("")
			}
		},
	)
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
	if a.DarkMode {
		a.ThemeButton.SetText("☀️") // Show sun icon if dark mode is enabled
	} else {
		a.ThemeButton.SetText("🌙") // Show moon icon if dark mode is disabled
	}
	// Update PathDisplays's colors
	a.FolderPath.RefreshColor(a.DarkMode)
	a.Window.Content().Refresh() // Immediately refresh window content
	runtime.GC()                 // Cleanup ram
}

// Cleanup ram
func (a *MainApp) Cleanup() {
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
