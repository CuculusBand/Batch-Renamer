package main

import (
	"fmt"
	"image/color"
	"path/filepath"
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
	App    fyne.App
	Window fyne.Window

	StatusLabel *widget.Label
	FilePath    *PathDisplay

	ThemeButton *widget.Button

	PreviewTable          *widget.Table
	PreviewTableContainer *container.Scroll
	DarkMode              bool
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
		App:    app,
		Window: window,

		DarkMode: isDark, // Save the dark mode preference
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
	title := widget.NewLabel("<Folder Creator>")
	// Title and theme button layout
	TitleContainer := container.NewHBox(
		title,
		layout.NewSpacer(),
		aboutButton,
		a.ThemeButton,
	)

	// Create scrollable path displays
	a.FilePath = CreatePathDisplay(a.Window)
	// Refresh the colors of the path displays based on the theme
	a.FilePath.RefreshColor(a.DarkMode)
	// Set default width
	a.FilePath.UpdatePathDisplayWidth(a.Window)
	// Display paths using two containers
	fileInfo := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Table file:	"),
			a.FilePath.Container,
		),
	)

	// Create buttons
	fileSelectButton := widget.NewButton("Select File", a.SelectTableFile)
	targetSelectButton := widget.NewButton("Target Path", a.SelectDestination)
	clearButton := widget.NewButton("Clear", a.ClearAll)
	createButton := widget.NewButton("Create", a.GenerateFolders)
	exitButton := widget.NewButton("Exit", func() { a.App.Quit() })
	// Button layout
	buttonRow := container.NewHBox(
		fileSelectButton,
		targetSelectButton,
		layout.NewSpacer(),
		clearButton,
		createButton,
		exitButton,
	)

	// Create status Lables
	a.StatusLabel = widget.NewLabel("Ready")
	a.StatusLabel.Wrapping = fyne.TextWrapWord

	// Create preview table
	a.PreviewTable = a.InitializeTable()
	a.PreviewTableContainer = container.NewScroll(a.PreviewTable)

	// Create the main content layout
	contentContainer := container.NewBorder(
		container.NewVBox(
			TitleContainer,
			widget.NewSeparator(),
			fileInfo,
			widget.NewSeparator(),
			buttonRow,
			widget.NewSeparator(),
			widget.NewLabel("Preview:"),
		),
		a.StatusLabel,
		nil,
		nil,
		a.PreviewTableContainer,
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

// Select a file to load table data
func (a *MainApp) SelectTableFile() {
	dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		// Check file type and handle errors
		if err != nil {
			a.StatusLabel.SetText("Wrong file: " + err.Error())
			return
		}
		if reader == nil {
			return
		}
		// Handle the file path
		FilePath := reader.URI().Path()
		if runtime.GOOS == "windows" {
			// Remove leading slash for Windows paths
			if len(FilePath) > 2 && FilePath[0] == '/' && FilePath[2] == ':' {
				FilePath = FilePath[1:]
			}
			// Replace forward slashes with backslashes for Windows compatibility
			FilePath = strings.ReplaceAll(FilePath, "/", "\\")
		}
		// Set the file path to the label
		a.FilePath.Text.Text = FilePath
		a.FilePath.Text.Refresh()
		a.StatusLabel.SetText("Loading...")
		// Load the file
		if err := a.Processor.LoadFile(FilePath); err != nil {
			a.StatusLabel.SetText("Failed to load: " + err.Error())
			return
		}
		// Ensure the container is using the new table
		a.PreviewTable = a.InitializeTable() // Load new data
		a.PreviewTableContainer.Content = a.PreviewTable
		a.AutoUpdateColumnWidths() // Update the table columns
		a.ResetTableScroll()       // Reset the table scrollbar
		a.PreviewTableContainer.Refresh()
		a.StatusLabel.SetText(fmt.Sprintf("All data loaded: %d rows", len(a.Processor.TableData)))
	}, a.Window).Show()
}

// Select a destination folder to create new folders
func (a *MainApp) SelectDestination() {
	dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
		if err != nil {
			a.StatusLabel.SetText("Wrong target path: " + err.Error())
			return
		}
		if list == nil {
			return
		}
		a.Processor.DestPath = list.Path()

		a.StatusLabel.SetText("Selected target path: " + filepath.Base(a.Processor.DestPath))
	}, a.Window).Show()
}

// Clear all content in the table
func (a *MainApp) ClearAll() {
	// Reset FilePath and DestPath
	a.FilePath.Text.Text = "No Selection"
	a.FilePath.Text.Refresh()
	a.ResetPathScroll()
	// Reset table
	a.PreviewTable = a.InitializeTable()
	// Update table container
	a.PreviewTableContainer.Content = a.PreviewTable
	// Reset scrollbar of table container
	a.ResetTableScroll()
	// Update status
	a.StatusLabel.SetText("All content cleared")
	// Cleanup ram
	a.Cleanup()
	// Reset Processor
	a.Processor = NewFileProcessor()
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

// Create table
func (a *MainApp) InitializeTable() *widget.Table {
	return widget.NewTable(
		func() (int, int) {
			if a.Processor == nil || len(a.Processor.TableData) == 0 {
				return 0, 0 // Check data in the Processor
			}
			return len(a.Processor.TableData), len(a.Processor.TableData[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{})
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			if a.Processor != nil &&
				len(a.Processor.TableData) > i.Row &&
				len(a.Processor.TableData[i.Row]) > i.Col {
				label.SetText(a.Processor.TableData[i.Row][i.Col])
			} else {
				label.SetText("")
			}
		},
	)
}

// Generate folders and update the status label
func (a *MainApp) GenerateFolders() {
	// Ensure a file is selected
	if a.Processor.TableFilePath == "" {
		a.StatusLabel.SetText("Select a file first!")
		return
	}
	// Ensure a destination path is selected
	if a.Processor.DestPath == "" {
		a.StatusLabel.SetText("Select a target path first!")
		return
	}
	// Ensure there is data to process
	if len(a.Processor.TableData) == 0 {
		a.StatusLabel.SetText("No available data!")
		return
	}
	// Call the method to batch create folders
	// returning the number of successes and any error encountered
	successCount, err := a.Processor.GenerateFolders()
	if err != nil {
		a.StatusLabel.SetText("Error: " + err.Error())
		return
	}
	a.PreviewTable.Refresh()
	a.StatusLabel.SetText(fmt.Sprintf("Sucessfully created %d folder(s)", successCount))
}

// Update PathDisplay width based on the window size
func (pd *PathDisplay) UpdatePathDisplayWidth(window fyne.Window) {
	winWidth := window.Canvas().Size().Width
	minWidth := float32(300)
	targetWidth := winWidth * 0.8
	targetWidth = max(minWidth, targetWidth)
	pd.Container.SetMinSize(fyne.NewSize(targetWidth, 45))
}

// Adjusts the column widths based on the content
func (a *MainApp) AutoUpdateColumnWidths() {
	minWidth := float32(80)
	padding := float32(20)
	if len(a.Processor.TableData) == 0 {
		a.PreviewTable.SetColumnWidth(0, float32(minWidth)) // If no data, set a default width
		return
	}
	numCols := len(a.Processor.TableData[0]) // Get the number of columns
	// Extract each column and update its width
	for col := 0; col < numCols; col++ {
		maxLen := float32(0) // No length limit
		// Extract each row in the column
		for row := 0; row < len(a.Processor.TableData); row++ {
			cellText := a.Processor.TableData[row][col]
			// Use MeasureText to calculate the width of the cell
			cellSize := fyne.MeasureText(cellText, theme.TextSize(), fyne.TextStyle{})
			// Update the maximum width of the cell
			if cellSize.Width > maxLen {
				maxLen = cellSize.Width
			}
		}
		// Add padding to the maximum length, ensure the width is larger than the minWidth
		width := max(maxLen+padding, minWidth)
		// Update column width
		a.PreviewTable.SetColumnWidth(col, width)
	}
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
	a.DestPath.RefreshColor(a.DarkMode)
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
