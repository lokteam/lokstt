package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Config struct {
	Language string `json:"language"`
	Model    string `json:"model"`
}

func getConfigPath() string {
	if p := os.Getenv("LOKSTT_CONFIG"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lokstt", "config.json")
}

func LoadConfig() Config {
	cfg := Config{Language: "auto", Model: "small"}
	data, err := os.ReadFile(getConfigPath())
	if err == nil {
		json.Unmarshal(data, &cfg)
	}
	return cfg
}

func saveConfig(cfg Config) {
	path := getConfigPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.WriteFile(path, data, 0644)
}

type App struct {
	Application    *gtk.Application
	Overlay        *Overlay
	OnConfigChange func(cfg Config)
	OnStop         func()
	OnCancel       func()
}

func NewApp() *App {
	app := gtk.NewApplication("com.github.lokstt", gio.ApplicationFlagsNone)
	uiApp := &App{Application: app}

	app.ConnectActivate(func() {
		// App runs in background, no main window initially.
		// Initialize the overlay so it's ready to be shown.
		uiApp.Overlay = NewOverlay(uiApp)
		uiApp.Application.Connect("startup", func() {})

		// Keep the application alive even without windows
		app.Hold()
	})

	return uiApp
}

func (a *App) Run(args []string) int {
	return a.Application.Run(args)
}

func (a *App) ShowSettings() {
	glib.IdleAdd(func() {
		win := gtk.NewApplicationWindow(a.Application)
		win.SetTitle("LokSTT Settings")
		win.SetDefaultSize(400, 500)

		mainBox := gtk.NewBox(gtk.OrientationVertical, 0)

		stack := gtk.NewStack()
		stack.SetVExpand(true)
		stack.SetHExpand(true)
		stack.SetTransitionType(gtk.StackTransitionTypeSlideLeftRight)

		cfg := LoadConfig()

		// GENERAL PAGE (MODELS)
		generalBox := gtk.NewBox(gtk.OrientationVertical, 10)
		generalBox.SetMarginTop(20)
		generalBox.SetMarginBottom(20)
		generalBox.SetMarginStart(20)
		generalBox.SetMarginEnd(20)

		modelLabel := gtk.NewLabel("")
		modelLabel.SetMarkup("<b>Default Inference Model</b>")
		modelLabel.SetHAlign(gtk.AlignStart)
		generalBox.Append(modelLabel)

		modelScroll := gtk.NewScrolledWindow()
		modelScroll.SetVExpand(true)
		modelScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

		modelList := gtk.NewListBox()
		modelList.SetSelectionMode(gtk.SelectionSingle)
		modelList.AddCSSClass("boxed-list")

		models := []struct{ id, name string }{
			{"tiny", "Tiny (Fastest, Lowest RAM)"},
			{"base", "Base (Fast)"},
			{"small", "Small (Balanced)"},
			{"medium", "Medium (Highly Accurate)"},
			{"large-v3-turbo", "Large-v3-turbo (Slowest, Best Quality)"},
		}

		var modelCheckButtons []*gtk.Image
		var selectedModelRow *gtk.ListBoxRow

		for _, m := range models {
			row := gtk.NewListBoxRow()
			rowBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
			rowBox.SetMarginTop(10)
			rowBox.SetMarginBottom(10)
			rowBox.SetMarginStart(10)
			rowBox.SetMarginEnd(10)

			nameLabel := gtk.NewLabel(m.name)
			nameLabel.SetHAlign(gtk.AlignStart)
			nameLabel.SetHExpand(true)

			checkIcon := gtk.NewImageFromIconName("object-select-symbolic")
			modelCheckButtons = append(modelCheckButtons, checkIcon)

			if m.id == cfg.Model {
				checkIcon.SetOpacity(1)
				selectedModelRow = row
			} else {
				checkIcon.SetOpacity(0)
			}

			rowBox.Append(nameLabel)
			rowBox.Append(checkIcon)
			row.SetChild(rowBox)
			modelList.Append(row)

			if selectedModelRow == row {
				modelList.SelectRow(row)
			}
		}

		modelList.ConnectRowActivated(func(row *gtk.ListBoxRow) {
			idx := row.Index()
			if idx >= 0 && idx < len(models) {
				cfg.Model = models[idx].id
				for i, icon := range modelCheckButtons {
					if i == idx {
						icon.SetOpacity(1)
					} else {
						icon.SetOpacity(0)
					}
				}
				go saveConfig(cfg)
				if a.OnConfigChange != nil {
					a.OnConfigChange(cfg)
				}
			}
		})

		modelScroll.SetChild(modelList)
		generalBox.Append(modelScroll)
		stack.AddTitled(generalBox, "general", "General")

		// LANGUAGES PAGE
		langBox := gtk.NewBox(gtk.OrientationVertical, 10)
		langBox.SetMarginTop(20)
		langBox.SetMarginBottom(20)
		langBox.SetMarginStart(20)
		langBox.SetMarginEnd(20)

		searchEntry := gtk.NewSearchEntry()
		searchEntry.SetPlaceholderText("Search languages...")
		langBox.Append(searchEntry)

		langScroll := gtk.NewScrolledWindow()
		langScroll.SetVExpand(true)
		langScroll.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)

		langList := gtk.NewListBox()
		langList.SetSelectionMode(gtk.SelectionSingle)
		langList.AddCSSClass("boxed-list")

		type langInfo struct {
			id   string
			name string
			row  *gtk.ListBoxRow
		}

		languages := []langInfo{
			{id: "auto", name: "Auto-detect"},
			{id: "en", name: "English"}, {id: "ru", name: "Russian"}, {id: "af", name: "Afrikaans"},
			{id: "sq", name: "Albanian"}, {id: "am", name: "Amharic"}, {id: "ar", name: "Arabic"},
			{id: "hy", name: "Armenian"}, {id: "as", name: "Assamese"}, {id: "az", name: "Azerbaijani"},
			{id: "ba", name: "Bashkir"}, {id: "eu", name: "Basque"}, {id: "be", name: "Belarusian"},
			{id: "bn", name: "Bengali"}, {id: "bs", name: "Bosnian"}, {id: "br", name: "Breton"},
			{id: "bg", name: "Bulgarian"}, {id: "my", name: "Burmese"}, {id: "ca", name: "Catalan"},
			{id: "zh", name: "Chinese"}, {id: "hr", name: "Croatian"}, {id: "cs", name: "Czech"},
			{id: "da", name: "Danish"}, {id: "nl", name: "Dutch"}, {id: "et", name: "Estonian"},
			{id: "fo", name: "Faroese"}, {id: "fi", name: "Finnish"}, {id: "fr", name: "French"},
			{id: "gl", name: "Galician"}, {id: "ka", name: "Georgian"}, {id: "de", name: "German"},
			{id: "el", name: "Greek"}, {id: "gu", name: "Gujarati"}, {id: "ht", name: "Haitian Creole"},
			{id: "ha", name: "Hausa"}, {id: "haw", name: "Hawaiian"}, {id: "he", name: "Hebrew"},
			{id: "hi", name: "Hindi"}, {id: "hu", name: "Hungarian"}, {id: "is", name: "Icelandic"},
			{id: "id", name: "Indonesian"}, {id: "it", name: "Italian"}, {id: "ja", name: "Japanese"},
			{id: "jw", name: "Javanese"}, {id: "kn", name: "Kannada"}, {id: "kk", name: "Kazakh"},
			{id: "km", name: "Khmer"}, {id: "ko", name: "Korean"}, {id: "la", name: "Latin"},
			{id: "lv", name: "Latvian"}, {id: "ln", name: "Lingala"}, {id: "lt", name: "Lithuanian"},
			{id: "lb", name: "Luxembourgish"}, {id: "mk", name: "Macedonian"}, {id: "mg", name: "Malagasy"},
			{id: "ms", name: "Malay"}, {id: "ml", name: "Malayalam"}, {id: "mt", name: "Maltese"},
			{id: "mi", name: "Maori"}, {id: "mr", name: "Marathi"}, {id: "mn", name: "Mongolian"},
			{id: "ne", name: "Nepali"}, {id: "no", name: "Norwegian"}, {id: "nn", name: "Nynorsk"},
			{id: "oc", name: "Occitan"}, {id: "ps", name: "Pashto"}, {id: "fa", name: "Persian"},
			{id: "pl", name: "Polish"}, {id: "pt", name: "Portuguese"}, {id: "pa", name: "Punjabi"},
			{id: "ro", name: "Romanian"}, {id: "sa", name: "Sanskrit"}, {id: "sr", name: "Serbian"},
			{id: "sn", name: "Shona"}, {id: "sd", name: "Sindhi"}, {id: "si", name: "Sinhala"},
			{id: "sk", name: "Slovak"}, {id: "sl", name: "Slovenian"}, {id: "so", name: "Somali"},
			{id: "es", name: "Spanish"}, {id: "su", name: "Sundanese"}, {id: "sw", name: "Swahili"},
			{id: "sv", name: "Swedish"}, {id: "tl", name: "Tagalog"}, {id: "tg", name: "Tajik"},
			{id: "ta", name: "Tamil"}, {id: "tt", name: "Tatar"}, {id: "te", name: "Telugu"},
			{id: "th", name: "Thai"}, {id: "bo", name: "Tibetan"}, {id: "tr", name: "Turkish"},
			{id: "tk", name: "Turkmen"}, {id: "uk", name: "Ukrainian"}, {id: "ur", name: "Urdu"},
			{id: "uz", name: "Uzbek"}, {id: "vi", name: "Vietnamese"}, {id: "cy", name: "Welsh"},
			{id: "yi", name: "Yiddish"}, {id: "yo", name: "Yoruba"},
		}

		var langCheckButtons []*gtk.Image
		var selectedLangRow *gtk.ListBoxRow

		for i := range languages {
			l := &languages[i]
			row := gtk.NewListBoxRow()
			l.row = row

			rowBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
			rowBox.SetMarginTop(10)
			rowBox.SetMarginBottom(10)
			rowBox.SetMarginStart(10)
			rowBox.SetMarginEnd(10)

			nameLabel := gtk.NewLabel(l.name)
			nameLabel.SetHAlign(gtk.AlignStart)
			nameLabel.SetHExpand(true)

			checkIcon := gtk.NewImageFromIconName("object-select-symbolic")
			langCheckButtons = append(langCheckButtons, checkIcon)

			if l.id == cfg.Language {
				checkIcon.SetOpacity(1)
				selectedLangRow = row
			} else {
				checkIcon.SetOpacity(0)
			}

			rowBox.Append(nameLabel)
			rowBox.Append(checkIcon)
			row.SetChild(rowBox)
			langList.Append(row)

			if selectedLangRow == row {
				langList.SelectRow(row)
			}
		}

		langList.SetFilterFunc(func(row *gtk.ListBoxRow) bool {
			searchText := strings.ToLower(searchEntry.Text())
			if searchText == "" {
				return true
			}
			for _, l := range languages {
				if l.row == row {
					return strings.Contains(strings.ToLower(l.name), searchText)
				}
			}
			return false
		})

		searchEntry.ConnectSearchChanged(func() {
			langList.InvalidateFilter()
		})

		langList.ConnectRowActivated(func(row *gtk.ListBoxRow) {
			var activatedIndex = -1
			for i, lang := range languages {
				if lang.row == row {
					activatedIndex = i
					break
				}
			}

			if activatedIndex != -1 {
				cfg.Language = languages[activatedIndex].id
				for i := range langCheckButtons {
					if i == activatedIndex {
						langCheckButtons[i].SetOpacity(1)
					} else {
						langCheckButtons[i].SetOpacity(0)
					}
				}
				go saveConfig(cfg)
				if a.OnConfigChange != nil {
					a.OnConfigChange(cfg)
				}
			}
		})

		langScroll.SetChild(langList)
		langBox.Append(langScroll)
		stack.AddTitled(langBox, "languages", "Languages")

		switcher := gtk.NewStackSwitcher()
		switcher.SetStack(stack)
		switcher.SetHAlign(gtk.AlignCenter)
		switcher.SetMarginBottom(10)
		switcher.SetMarginTop(10)

		mainBox.Append(stack)
		mainBox.Append(switcher)

		win.SetChild(mainBox)
		win.Show()
	})
}
