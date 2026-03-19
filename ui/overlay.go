package ui

import (
	"fmt"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Overlay struct {
	win   *gtk.ApplicationWindow
	timer *gtk.Label
	bars  []*gtk.Box
}

func NewOverlay(uiApp *App) *Overlay {
	win := gtk.NewApplicationWindow(uiApp.Application)
	win.SetDecorated(false)
	win.SetResizable(false)
	win.AddCSSClass("transparent-window")
	win.RemoveCSSClass("background")

	keyCtrl := gtk.NewEventControllerKey()
	keyCtrl.Connect("key-pressed", func(keyval uint, keycode uint, state gdk.ModifierType) bool {
		if keyval == 65293 || keyval == 65421 {
			if uiApp.OnStop != nil {
				go uiApp.OnStop()
			}
			return true
		}
		if keyval == 65307 {
			if uiApp.OnCancel != nil {
				go uiApp.OnCancel()
			}
			return true
		}
		return false
	})
	win.AddController(keyCtrl)

	// Main vertical layout
	mainBox := gtk.NewBox(gtk.OrientationVertical, 12)
	mainBox.SetHAlign(gtk.AlignCenter)
	mainBox.SetVAlign(gtk.AlignCenter)
	mainBox.SetMarginTop(20)
	mainBox.SetMarginBottom(20)
	mainBox.SetMarginStart(20)
	mainBox.SetMarginEnd(20)

	// The Squircle container (the colored box)
	squircle := gtk.NewBox(gtk.OrientationHorizontal, 0)
	squircle.AddCSSClass("squircle")
	squircle.SetHAlign(gtk.AlignCenter)
	squircle.SetVAlign(gtk.AlignCenter)
	squircle.SetSizeRequest(110, 110) // Fixed square size

	// Inner container for the 5 waveform bars
	waveBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	waveBox.SetHAlign(gtk.AlignCenter)
	waveBox.SetVAlign(gtk.AlignCenter)
	waveBox.SetHExpand(true)
	waveBox.SetVExpand(true)

	bars := make([]*gtk.Box, 5)
	for i := 0; i < 5; i++ {
		bar := gtk.NewBox(gtk.OrientationVertical, 0)
		bar.AddCSSClass("wave-bar")
		bar.SetSizeRequest(8, 10) // Initial small height
		bar.SetVAlign(gtk.AlignCenter)
		waveBox.Append(bar)
		bars[i] = bar
	}

	squircle.Append(waveBox)

	// Timer Label below the squircle
	timer := gtk.NewLabel("00:00")
	timer.AddCSSClass("timer-label")

	mainBox.Append(squircle)
	mainBox.Append(timer)

	win.SetChild(mainBox)

	// Custom CSS
	cssProvider := gtk.NewCSSProvider()
	cssProvider.LoadFromData(`
		window.transparent-window,
		window.transparent-window decoration {
			background-color: rgba(0,0,0,0);
			background-image: none;
			box-shadow: none;
			border: none;
		}
		.squircle {
			background-color: @theme_selected_bg_color; /* Uses your system accent color */
			border-radius: 36px; /* Smooth rounded squircle shape */
			box-shadow: 0 10px 30px rgba(0,0,0,0.3);
		}
		.wave-bar {
			background-color: #ffffff; /* White bars */
			border-radius: 4px;
			transition: all 100ms cubic-bezier(0.25, 0.46, 0.45, 0.94); /* Smooth bouncing */
		}
		.timer-label {
			color: @theme_fg_color; /* System text color */
			text-shadow: 0 2px 5px rgba(0,0,0,0.6);
			font-family: monospace;
			font-weight: 900;
			font-size: 22px;
			margin-top: 10px;
		}
	`)
	gtk.StyleContextAddProviderForDisplay(
		gdk.DisplayGetDefault(),
		cssProvider,
		gtk.STYLE_PROVIDER_PRIORITY_APPLICATION,
	)

	return &Overlay{
		win:   win,
		timer: timer,
		bars:  bars,
	}
}

func (o *Overlay) Show() {
	glib.IdleAdd(func() {
		o.win.Show()
	})
}

func (o *Overlay) Hide() {
	glib.IdleAdd(func() {
		o.win.Hide()
	})
}

func (o *Overlay) UpdateVolume(level float64) {
	glib.IdleAdd(func() {
		if level < 0 {
			level = 0
		}
		if level > 1 {
			level = 1
		}

		// Calculate heights for the 5 bars to make it look like a symmetric waveform
		// Base height is minimum, max expands based on 'level'
		// Pattern: Small (outer), Medium, Large (center)
		h1 := 10 + int(level*18) // Outer bars (1 and 5)
		h2 := 16 + int(level*36) // Inner bars (2 and 4)
		h3 := 22 + int(level*55) // Center bar (3)

		o.bars[0].SetSizeRequest(8, h1)
		o.bars[1].SetSizeRequest(8, h2)
		o.bars[2].SetSizeRequest(8, h3)
		o.bars[3].SetSizeRequest(8, h2)
		o.bars[4].SetSizeRequest(8, h1)
	})
}

func (o *Overlay) UpdateTimer(seconds int) {
	glib.IdleAdd(func() {
		mins := seconds / 60
		secs := seconds % 60
		o.timer.SetText(fmt.Sprintf("%02d:%02d", mins, secs))
	})
}
