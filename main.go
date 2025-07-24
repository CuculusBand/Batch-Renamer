package main

import (
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// Create the application
	MyApp := app.NewWithID("Batch Renamer")
	// Load and Set the custom font file
	//customFont := fyne.NewStaticResource("NotoSans", LoadFont("fonts/NotoSans-SemiBold.ttf"))
	MyApp.Settings().SetTheme(&appTheme{regularFont: AppFont})
	// Set the application icon
	icon, err := fyne.LoadResourceFromPath("GO-BatchRenamer-icon.png")
	if err != nil {
		log.Fatal("Faile to load icon:", err)
	}
	MyApp.SetIcon(icon)

	// Create the window
	MainWindow := MyApp.NewWindow("Batch Renamer")
	MainWindow.Resize(fyne.NewSize(600, 850))
	MainWindow.SetFixedSize(false)
	MainWindow.CenterOnScreen()
	app := InitializeApp(MyApp, MainWindow)
	time.Sleep(50 * time.Millisecond)
	// Create UI
	app.MakeUI()
	// Run the application
	MainWindow.ShowAndRun()
}
