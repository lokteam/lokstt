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

		// Keep the application alive even without windows
		app.Hold()
	})

	return uiApp
}

func (a *App) Run(args []string) int {
	return a.Application.Run(args)
}

func (a *App) ShowSettings() {
	win := gtk.NewApplicationWindow(a.Application)
	win.SetTitle("LokSTT Settings")
	win.SetDefaultSize(400, 500)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 0)

		stack := gtk.NewStack()
		stack.SetVExpand(true)
		stack.SetHExpand(true)
		stack.SetTransitionType(gtk.StackTransitionTypeSlideLeftRight)

		cfg := LoadConfig()

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
			if m.id != cfg.Model {
				checkIcon.SetOpacity(0)
			}
			modelCheckButtons = append(modelCheckButtons, checkIcon)

			rowBox.Append(nameLabel)
			rowBox.Append(checkIcon)
			row.SetChild(rowBox)

			modelList.Append(row)
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

		languages := []struct{ id, name string }{
			{"auto", "Auto-detect"},
			{"en", "English"},
			{"ru", "Russian"},
			{"af", "Afrikaans"},
			{"sq", "Albanian"},
			{"am", "Amharic"},
			{"ar", "Arabic"},
			{"hy", "Armenian"},
			{"as", "Assamese"},
			{"az", "Azerbaijani"},
			{"ba", "Bashkir"},
			{"eu", "Basque"},
			{"be", "Belarusian"},
			{"bn", "Bengali"},
			{"bs", "Bosnian"},
			{"br", "Breton"},
			{"bg", "Bulgarian"},
			{"my", "Burmese"},
			{"ca", "Catalan"},
			{"zh", "Chinese"},
			{"hr", "Croatian"},
			{"cs", "Czech"},
			{"da", "Danish"},
			{"nl", "Dutch"},
			{"et", "Estonian"},
			{"fo", "Faroese"},
			{"fi", "Finnish"},
			{"fr", "French"},
			{"gl", "Galician"},
			{"ka", "Georgian"},
			{"de", "German"},
			{"el", "Greek"},
			{"gu", "Gujarati"},
			{"ht", "Haitian Creole"},
			{"ha", "Hausa"},
			{"haw", "Hawaiian"},
			{"he", "Hebrew"},
			{"hi", "Hindi"},
			{"hu", "Hungarian"},
			{"is", "Icelandic"},
			{"id", "Indonesian"},
			{"it", "Italian"},
			{"ja", "Japanese"},
			{"jw", "Javanese"},
			{"kn", "Kannada"},
			{"kk", "Kazakh"},
			{"km", "Khmer"},
			{"ko", "Korean"},
			{"la", "Latin"},
			{"lv", "Latvian"},
			{"ln", "Lingala"},
			{"lt", "Lithuanian"},
			{"lb", "Luxembourgish"},
			{"mk", "Macedonian"},
			{"mg", "Malagasy"},
			{"ms", "Malay"},
			{"ml", "Malayalam"},
			{"mt", "Maltese"},
			{"mi", "Maori"},
			{"mr", "Marathi"},
			{"mn", "Mongolian"},
			{"ne", "Nepali"},
			{"no", "Norwegian"},
			{"nn", "Nynorsk"},
			{"oc", "Occitan"},
			{"ps", "Pashto"},
			{"fa", "Persian"},
			{"pl", "Polish"},
			{"pt", "Portuguese"},
			{"pa", "Punjabi"},
			{"ro", "Romanian"},
			{"sa", "Sanskrit"},
			{"sr", "Serbian"},
			{"sn", "Shona"},
			{"sd", "Sindhi"},
			{"si", "Sinhala"},
			{"sk", "Slovak"},
			{"sl", "Slovenian"},
			{"so", "Somali"},
			{"es", "Spanish"},
			{"su", "Sundanese"},
			{"sw", "Swahili"},
			{"sv", "Swedish"},
			{"tl", "Tagalog"},
			{"tg", "Tajik"},
			{"ta", "Tamil"},
			{"tt", "Tatar"},
			{"te", "Telugu"},
			{"th", "Thai"},
			{"bo", "Tibetan"},
			{"tr", "Turkish"},
			{"tk", "Turkmen"},
			{"uk", "Ukrainian"},
			{"ur", "Urdu"},
			{"uz", "Uzbek"},
			{"vi", "Vietnamese"},
			{"cy", "Welsh"},
			{"yi", "Yiddish"},
			{"yo", "Yoruba"},
		}

		var langCheckButtons []*gtk.Image

		for _, l := range languages {
			row := gtk.NewListBoxRow()
			rowBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
			rowBox.SetMarginTop(10)
			rowBox.SetMarginBottom(10)
			rowBox.SetMarginStart(10)
			rowBox.SetMarginEnd(10)

			nameLabel := gtk.NewLabel(l.name)
			nameLabel.SetHAlign(gtk.AlignStart)
			nameLabel.SetHExpand(true)

			checkIcon := gtk.NewImageFromIconName("object-select-symbolic")
			if l.id != cfg.Language {
				checkIcon.SetOpacity(0)
			}
			langCheckButtons = append(langCheckButtons, checkIcon)

			rowBox.Append(nameLabel)
			rowBox.Append(checkIcon)
			row.SetChild(rowBox)

			langList.Append(row)
		}

		langList.SetFilterFunc(func(row *gtk.ListBoxRow) bool {
			idx := row.Index()
			if idx < 0 || idx >= len(languages) {
				return false
			}
			searchText := strings.ToLower(searchEntry.Text())
			if searchText == "" {
				return true
			}
			langName := strings.ToLower(languages[idx].name)
			return strings.Contains(langName, searchText)
		})

		searchEntry.ConnectSearchChanged(func() {
			langList.InvalidateFilter()
		})

		langList.ConnectRowActivated(func(row *gtk.ListBoxRow) {
			idx := row.Index()
			if idx >= 0 && idx < len(languages) {
				cfg.Language = languages[idx].id
				for i, icon := range langCheckButtons {
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
		win.Present()
	})
}
