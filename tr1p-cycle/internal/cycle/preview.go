package cycle

import (
	"image/color"
	"log"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
)

type Preview struct {
	content *fyne.Container
	window  fyne.Window
	visible bool
	mu      sync.Mutex
	cl      *CycleList
}

func NewPreview(app fyne.App, cl *CycleList) *Preview {
	log.Println("Creating new Preview")
	drv := app.Driver()
	if drv, ok := drv.(desktop.Driver); ok {
		w := drv.CreateSplashWindow()
		log.Println("Splash window created")
		w.RequestFocus()
		overlayColor := theme.Color(theme.ColorNameBackground)
		r, g, b, _ := overlayColor.RGBA()
		overlay := canvas.NewRectangle(color.NRGBA{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
			A: 230,
		})
		content := container.NewStack(overlay)
		w.SetContent(content)
		w.Resize(fyne.NewSize(500, 400))
		return &Preview{
			content: content,
			window:  w,
			visible: false,
			cl:      cl,
		}
	}
	log.Println("Failed to create Preview: driver does not support desktop")
	return nil
}

func (p *Preview) ShowPreview() {
	log.Println("ShowPreview called")
	if p == nil || p.window == nil {
		log.Println("Preview or window is nil")
		return
	}
	p.mu.Lock()
	p.visible = true
	p.mu.Unlock()
	p.updateContent()
	p.window.Show()
	p.window.RequestFocus()
}

func (p *Preview) HidePreview() {
	log.Println("HidePreview called")
	if p == nil || p.window == nil {
		log.Println("Preview or window is nil")
		return
	}
	p.mu.Lock()
	p.visible = false
	p.mu.Unlock()
	p.window.Hide()
	log.Println("Preview window hidden")
}

func (p *Preview) IsVisible() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.visible
}

func (p *Preview) updateContent() {
	items := p.cl.GetItems()
	currentItem := p.cl.GetCurrentItem()
	content := generatePreviewContent(items, currentItem)
	p.content.Objects = []fyne.CanvasObject{p.content.Objects[0], content}
	p.content.Refresh()
}

func generatePreviewContent(items []CycleItem, currentItem *CycleItem) *fyne.Container {
	if len(items) == 0 {
		fgColor := theme.Color(theme.ColorNameForeground)
		emptyText := canvas.NewText("The cycle list is empty.", fgColor)
		emptyText.TextSize = 18
		return container.NewCenter(emptyText)
	}

	listContainer := container.NewVBox()

	for i, item := range items {
		isActive := &item == currentItem
		isTopItem := i == 0

		row := container.NewHBox()

		// Add prefix for the active item
		prefix := canvas.NewText("  ", color.Transparent) // Two spaces for alignment
		if isActive {
			prefix = canvas.NewText("> ", theme.Color(theme.ColorNamePrimary))
		}
		prefix.TextSize = 20
		row.Add(prefix)

		// Add item details
		details := container.NewVBox()

		titleText := canvas.NewText(item.title, theme.Color(theme.ColorNameForeground))
		titleText.TextSize = 16
		if isActive {
			titleText.Color = theme.Color(theme.ColorNamePrimary)
			titleText.TextStyle = fyne.TextStyle{Bold: true}
		}
		if isTopItem {
			titleText.Color = color.NRGBA{R: 0, G: 0, B: 139, A: 255} // Dark Blue
			titleText.TextStyle = fyne.TextStyle{Bold: true}
		}
		details.Add(titleText)

		appNameText := canvas.NewText(item.appName, theme.Color(theme.ColorNamePlaceHolder))
		appNameText.TextSize = 12
		if isTopItem {
			appNameText.Color = color.NRGBA{R: 0, G: 0, B: 139, A: 255} // Dark Blue
		}
		details.Add(appNameText)

		row.Add(details)

		// Create a background for the entire row
		var background *canvas.Rectangle
		if isTopItem {
			background = canvas.NewRectangle(color.NRGBA{R: 144, G: 238, B: 144, A: 255}) // Light Green
		} else if isActive {
			background = canvas.NewRectangle(theme.Color(theme.ColorNameSelection))
		} else {
			background = canvas.NewRectangle(color.Transparent)
		}

		// Combine background and content
		rowContainer := container.NewStack(background, row)

		listContainer.Add(rowContainer)
		listContainer.Add(layout.NewSpacer()) // Add space between items
	}

	// Wrap the list in a scroll container
	scroll := container.NewScroll(listContainer)
	scroll.SetMinSize(fyne.NewSize(400, 300))

	// Align the scroll container to the left
	return container.NewHBox(scroll, layout.NewSpacer())
}
