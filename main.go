package main

import (
	"fmt"
	"log"
	"lokstt/paster"
	"lokstt/ui"
	"math"
	"net"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/gordonklaus/portaudio"
)

var SocketPath = "/tmp/lokstt.sock"

func init() {
	if s := os.Getenv("LOKSTT_SOCKET"); s != "" {
		SocketPath = s
	}
}

const (
	SampleRate = 16000
)

type Daemon struct {
	model            whisper.Model
	context          whisper.Context
	currentModelName string
	stream           *portaudio.Stream
	audioData        []float32
	recording        bool
	actualSampleRate float64
	mu               sync.Mutex
	ui               *ui.App
}

func (d *Daemon) LoadModel(modelName string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.currentModelName == modelName && d.model != nil {
		return nil
	}

	if d.model != nil {
		fmt.Println("Freeing previous model memory...")
		d.model.Close()
		d.model = nil
		d.context = nil
	}

	modelFiles := map[string]string{
		"tiny":           "ggml-tiny.bin",
		"base":           "ggml-base.bin",
		"small":          "ggml-small.bin",
		"medium":         "ggml-medium-q5_0.bin",
		"large-v3-turbo": "ggml-large-v3-turbo-q5_0.bin",
	}

	fileName, ok := modelFiles[modelName]
	if !ok {
		fileName = "ggml-small.bin"
	}

	path := fmt.Sprintf("/usr/share/whisper-models/%s", fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = "/usr/share/whisper.cpp-model-small/ggml-small.bin"
	}

	fmt.Printf("Loading Whisper model '%s' from %s\n", modelName, path)
	model, err := whisper.New(path)
	if err != nil {
		return fmt.Errorf("failed to load model: %w", err)
	}

	ctx, err := model.NewContext()
	if err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}

	d.model = model
	d.context = ctx
	d.currentModelName = modelName
	return nil
}

func NewDaemon(u *ui.App) (*Daemon, error) {
	cfg := ui.LoadConfig()

	d := &Daemon{
		audioData: make([]float32, 0),
		recording: false,
		ui:        u,
	}

	if err := d.LoadModel(cfg.Model); err != nil {
		return nil, err
	}

	u.OnConfigChange = func(newCfg ui.Config) {
		go func() {
			if err := d.LoadModel(newCfg.Model); err != nil {
				fmt.Println("Error reloading model:", err)
			}
		}()
	}

	u.OnStop = func() {
		d.StopAndTranscribe()
	}
	u.OnCancel = func() {
		d.CancelRecording()
	}

	return d, nil
}

func calculateVolume(in []float32) float64 {
	var sum float32
	for _, v := range in {
		sum += v * v
	}
	rms := math.Sqrt(float64(sum) / float64(len(in)))
	return math.Min(1.0, rms*3.0)
}

func (d *Daemon) StartRecording() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.recording {
		return "Already recording\n"
	}

	d.audioData = make([]float32, 0)
	err := portaudio.Initialize()
	if err != nil {
		return fmt.Sprintf("Error initializing portaudio: %v\n", err)
	}

	deviceInfo, err := portaudio.DefaultInputDevice()
	if err != nil {
		return fmt.Sprintf("Error getting default device: %v\n", err)
	}

	d.actualSampleRate = deviceInfo.DefaultSampleRate
	d.stream, err = portaudio.OpenDefaultStream(1, 0, d.actualSampleRate, 0, func(in []float32) {
		d.audioData = append(d.audioData, in...)
		level := calculateVolume(in)
		d.ui.Overlay.UpdateVolume(level)
	})
	if err := d.stream.Start(); err != nil {
		return fmt.Sprintf("Error starting stream: %v\n", err)
	}

	d.recording = true
	d.ui.Overlay.Show()
	d.ui.Overlay.UpdateTimer(0)

	go func() {
		start := time.Now()
		for {
			d.mu.Lock()
			if !d.recording {
				d.mu.Unlock()
				break
			}
			d.mu.Unlock()
			elapsed := int(time.Since(start).Seconds())
			d.ui.Overlay.UpdateTimer(elapsed)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	fmt.Println("Started recording...")
	return "Recording started\n"
}

func (d *Daemon) CancelRecording() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.recording {
		return "Not recording\n"
	}

	d.stream.Stop()
	d.stream.Close()
	portaudio.Terminate()
	d.recording = false
	d.ui.Overlay.Hide()
	fmt.Println("Recording canceled.")
	return "Recording canceled\n"
}

func (d *Daemon) StopAndTranscribe() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.recording {
		return "Not recording\n"
	}

	d.stream.Stop()
	d.stream.Close()
	portaudio.Terminate()
	d.recording = false
	d.ui.Overlay.Hide()
	fmt.Println("Stopped recording. Processing...")

	if len(d.audioData) == 0 {
		return "No audio data recorded\n"
	}

	var resampledData []float32
	if d.actualSampleRate != SampleRate {
		ratio := float64(SampleRate) / d.actualSampleRate
		newLen := int(float64(len(d.audioData)) * ratio)
		resampledData = make([]float32, newLen)
		for i := 0; i < newLen; i++ {
			oldIdx := int(float64(i) / ratio)
			if oldIdx < len(d.audioData) {
				resampledData[i] = d.audioData[oldIdx]
			}
		}
	} else {
		resampledData = d.audioData
	}

	fmt.Printf("Captured %d samples (resampled to %d). Transcribing...\n", len(d.audioData), len(resampledData))

	var textBuilder strings.Builder
	var cb whisper.SegmentCallback = func(segment whisper.Segment) {
		textBuilder.WriteString(segment.Text)
		textBuilder.WriteString(" ")
	}

	cfg := ui.LoadConfig()
	if cfg.Language != "auto" {
		d.context.SetLanguage(cfg.Language)
	} else {
		d.context.SetLanguage("auto")
	}

	threads := uint(runtime.NumCPU())
	if threads > 2 {
		threads = threads - 1
	}
	d.context.SetThreads(threads)

	err := d.context.Process(resampledData, nil, cb, nil)
	if err != nil {
		return fmt.Sprintf("Transcription error: %v\n", err)
	}

	text := strings.TrimSpace(textBuilder.String())

	re := regexp.MustCompile(`\[.*?\]|\(.*?\)|♪.*?♪`)
	text = strings.TrimSpace(re.ReplaceAllString(text, ""))

	fmt.Println("Transcription:", text)

	if text != "" {
		go paster.PasteText(text)
	}

	return fmt.Sprintf("Transcribed: %s\n", text)
}

func (d *Daemon) Close() {
	if d.recording && d.stream != nil {
		d.stream.Stop()
		d.stream.Close()
		portaudio.Terminate()
	}
	if d.model != nil {
		d.model.Close()
	}
}

func main() {
	u := ui.NewApp()

	if len(os.Args) > 1 && os.Args[1] == "--settings" {
		u.Application.ConnectActivate(func() {
			u.ShowSettings()
		})
		os.Exit(u.Run(os.Args[1:]))
	}

	go func() {
		daemon, err := NewDaemon(u)
		if err != nil {
			log.Fatal(err)
		}
		defer daemon.Close()

		if _, err := os.Stat(SocketPath); err == nil {
			os.Remove(SocketPath)
		}

		l, err := net.Listen("unix", SocketPath)
		if err != nil {
			log.Fatal("listen error:", err)
		}
		defer l.Close()

		fmt.Println("Listening on", SocketPath)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			l.Close()
			os.Exit(0)
		}()

		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal("accept error:", err)
			}

			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				n, err := c.Read(buf)
				if err != nil {
					return
				}

				cmd := strings.TrimSpace(string(buf[:n]))
				var response string

				switch cmd {
				case "START":
					response = daemon.StartRecording()
				case "STOP":
					response = daemon.StopAndTranscribe()
				case "TOGGLE":
					if daemon.recording {
						response = daemon.StopAndTranscribe()
					} else {
						response = daemon.StartRecording()
					}
				case "CANCEL":
					response = daemon.CancelRecording()
				case "SETTINGS":
					u.ShowSettings()
					response = "Settings opened"
				default:
					response = "Unknown command\n"
				}

				c.Write([]byte(response))
			}(conn)
		}
	}()

	os.Exit(u.Run(os.Args))
}
