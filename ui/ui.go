package ui

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Config struct {
	Language string `json:"language"`
	Model    string `json:"model"`
}

func getConfigPath() string {
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
	glib.IdleAdd(func() {
		win := gtk.NewApplicationWindow(a.Application)
		win.SetTitle("LokSTT Settings")
		win.SetDefaultSize(300, 250)

		box := gtk.NewBox(gtk.OrientationVertical, 10)
		box.SetMarginTop(20)
		box.SetMarginBottom(20)
		box.SetMarginStart(20)
		box.SetMarginEnd(20)

		label := gtk.NewLabel("Transcription Language:")
		label.SetHAlign(gtk.AlignStart)
		box.Append(label)

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

		var langNames []string
		var langSelected uint = 0
		cfg := LoadConfig()

		for i, l := range languages {
			langNames = append(langNames, l.name)
			if l.id == cfg.Language {
				langSelected = uint(i)
			}
		}

		langDrop := gtk.NewDropDownFromStrings(langNames)
		langDrop.SetEnableSearch(true)
		exp := gtk.NewPropertyExpression(gtk.GTypeStringObject, nil, "string")
		langDrop.SetExpression(exp)
		langDrop.SetSelected(langSelected)

		modelLabel := gtk.NewLabel("Model (Quantized q5_0):")
		modelLabel.SetHAlign(gtk.AlignStart)
		modelLabel.SetMarginTop(10)

		models := []struct{ id, name string }{
			{"tiny", "Tiny (Fastest, Lowest RAM)"},
			{"base", "Base (Fast)"},
			{"small", "Small (Balanced)"},
			{"medium", "Medium (Highly Accurate)"},
			{"large-v3-turbo", "Large-v3-turbo (Slowest, Best Quality)"},
		}

		var modelNames []string
		var modelSelected uint = 2
		for i, m := range models {
			modelNames = append(modelNames, m.name)
			if m.id == cfg.Model {
				modelSelected = uint(i)
			}
		}

		modelDrop := gtk.NewDropDownFromStrings(modelNames)
		modelDrop.SetSelected(modelSelected)

		saveAndNotify := func() {
			lIdx := int(langDrop.Selected())
			mIdx := int(modelDrop.Selected())
			
			if lIdx >= 0 && lIdx < len(languages) {
				cfg.Language = languages[lIdx].id
			}
			if mIdx >= 0 && mIdx < len(models) {
				cfg.Model = models[mIdx].id
			}

			go saveConfig(cfg)
			if a.OnConfigChange != nil {
				a.OnConfigChange(cfg)
			}
		}

		langDrop.Connect("notify::selected", saveAndNotify)
		modelDrop.Connect("notify::selected", saveAndNotify)

		box.Append(langDrop)
		box.Append(modelLabel)
		box.Append(modelDrop)
		win.SetChild(box)
		win.Show()
	})
}
