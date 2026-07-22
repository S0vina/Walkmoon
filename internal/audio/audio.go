package audio

// imports
import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhowden/tag"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
)

// struct: stores path and id(id stands for music number on the playlist)
type Musica struct {
	Id        int
	Path      string
	Ext       string
	Title     string
	Artist    string
	Album     string
	Genre     string
	ImageData []byte
	ImageMIME string
}

// Centralizes all components that perdure in AudioPlayer Execution
type AudioPlayer struct {
	SampleRate     beep.SampleRate
	Ctrl           *beep.Ctrl
	Volume         *effects.Volume
	CurrentIndex   int
	LoopPlaylist   bool
	LoopSong       bool
	PlayShuffle    bool
	InputChan      chan string
	EventChan      chan tea.Msg
	SelectSongChan chan string
	CurrentSong    Musica
	PlayerState    PlayerState
}

type PlayerState struct {
	PSlastTrackPath string  `json:"Path"`
	PSvolume        float64 `json: "Volume"`
	PScurrentIndex  int     `json: "CurrentIndex"`
	PSloopPlaylist  bool    `json: "LoopPlaylist"`
	PSloopSong      bool    `json: "LoopSong"`
	PSplayShuffle   bool    `json: "PlayShuffle"`
}

type (
	SongChanged  struct{ Song Musica }
	ShuffleState struct{}
)

type ProgressChanged struct {
	Current time.Duration
	Total   time.Duration
}

// method: creates a new audioPlayer
func New(state *PlayerState) (ap *AudioPlayer, err error) {
	sr := beep.SampleRate(44100)

	ctrl := &beep.Ctrl{Paused: true} // o player comeca pausado

	inputChan := make(chan string, 1)
	eventChan := make(chan tea.Msg, 5)
	selectChan := make(chan string, 1)
	var v float64
	v = -1
	ci := -1 // if stays minus 1, then no state has been created
	loopP := true
	loopS := false
	ps := false

	if state != nil {
		v = state.PSvolume
		ci = state.PScurrentIndex
		loopP = state.PSloopPlaylist
		loopS = state.PSloopSong
		ps = state.PSplayShuffle
	}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   v,
		Silent:   false,
	}
	if ci == -1 {
		ci = 0
	}

	CurrentIndex := ci

	loopPlaylist := loopP
	loopSong := loopS
	playShuffle := ps
	err = speaker.Init(sr, sr.N(time.Second/10))
	if err != nil {
		log.Printf("Speaker could not be initiaded")
	}

	ap = &AudioPlayer{
		SampleRate:     sr,
		Ctrl:           ctrl,
		Volume:         volume,
		CurrentIndex:   CurrentIndex,
		LoopPlaylist:   loopPlaylist,
		LoopSong:       loopSong,
		PlayShuffle:    playShuffle,
		InputChan:      inputChan,
		EventChan:      eventChan,
		SelectSongChan: selectChan,
	}

	// log.Printf("Player montado com sucesso")
	err = nil
	return
}

// method: scans the folder passed in the arg and return the files in it
func (ap *AudioPlayer) ScanFolder(root string) (playlist []Musica, err error) {
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	contador := 1

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Garante que não é uma pasta antes de checar a extensão
		if !info.IsDir() {
			ext := strings.TrimSpace(strings.ToLower(filepath.Ext(path))) // Pega a extensão (ex: ".mp3", ".flac")

			if ext == ".mp3" || ext == ".flac" || ext == ".wav" {
				file, err := os.Open(path)
				if err != nil {
					return nil // Ignora arquivos que não consegue abrir e continua o Walk
				}

				// Tenta ler os metadados do arquivo
				m, err := tag.ReadFrom(file)
				if err != nil {
					// Se o arquivo não tiver tags (muito comum em .wav),
					// ainda adicionamos à playlist usando o nome do arquivo como título básico
					musica := Musica{
						Id:    contador,
						Path:  path,
						Ext:   ext,
						Title: info.Name(), // Usa o nome do arquivo (ex: "musica.wav") como fallback
					}
					playlist = append(playlist, musica)
					contador++
					return nil
				}

				// Monta a struct Musica com os metadados extraídos
				musica := Musica{
					Id:     contador,
					Path:   path,
					Ext:    ext,
					Title:  m.Title(),
					Artist: m.Artist(),
					Album:  m.Album(),
					Genre:  m.Genre(),
				}

				// Se o título vier vazio nas tags, usa o nome do arquivo para não ficar em branco na UI
				if musica.Title == "" {
					musica.Title = info.Name()
				}

				// Se houver capa de álbum
				if pic := m.Picture(); pic != nil {
					musica.ImageData = pic.Data
					musica.ImageMIME = pic.MIMEType
				}

				// Adiciona a música completa na playlist
				playlist = append(playlist, musica)
				contador++
			}
		}
		return nil
	})

	return playlist, err
}

// #####--------------------- TO DO ---------------------####
// func decTypeArchive(f *os.File) {

// }

// Logica de contagem de tempo
// select {
// case <-done:
// 	return
// case <-time.After(time.Second):
// 	speaker.Lock()
// 	fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
// 	speaker.Unlock()
// }

// method: run audioPlayer
func (ap *AudioPlayer) Run(playlist []Musica) {
	var queue []Musica
	var shuffleList []Musica

	queue = playlist

RunLoop:
	for ap.CurrentIndex < len(queue) {
		ap.CurrentSong = queue[ap.CurrentIndex]
		count := 0

		// log.Print(ap.CurrentSong.Artist)

		ap.EventChan <- SongChanged{Song: ap.CurrentSong}
		imm, end := ap.Play(ap.CurrentIndex, queue)
		speaker.Lock()
		ap.Ctrl.Paused = false
		speaker.Unlock()
		if end {
			return
		}

		count += ap.CurrentIndex + imm

		// if playlist ends and no LoopPlaylist set
		if count > len(queue)-1 && !ap.LoopPlaylist {
			return
		}

		// if playshuffle and shufflelist is not initiated
		if ap.PlayShuffle && len(shuffleList) == 0 {
			shuffleList = GenerateShuffle(playlist)

			// assuming that queue is now, plalylist
			musicaAtual := queue[ap.CurrentIndex]

			// Se, por azar a primeira música do Shuffle for a mesma que está tocando agora
			if len(shuffleList) > 1 && shuffleList[0].Title == musicaAtual.Title {
				// Rotaciona a lista e joga a primeira música para o final
				primeira := shuffleList[0]
				shuffleList = append(shuffleList[1:], primeira)
			}

			queue = shuffleList
			ap.CurrentIndex = 0
			continue RunLoop
		}

		// if sequencial mode is on and shuffleMode was on before
		if !ap.PlayShuffle && len(shuffleList) > 0 {

			currentSong := shuffleList[ap.CurrentIndex]
			aux := currentSong.Id + imm

			shuffleList = nil

			queue = playlist
			ap.CurrentIndex = ap.updateSong(aux, queue)
			continue RunLoop
		}

		ap.CurrentIndex = ap.updateSong(count, queue)

		// log.Println(ap.PlayShuffle)
	}
}

func (ap *AudioPlayer) updateSong(count int, queue []Musica) (nextSongIndex int) {
	totalMusicas := len(queue)
	if totalMusicas == 0 {
		return 0
	}

	// Trata o comportamento ao voltar (count < 0)
	if count < 0 {
		if ap.LoopPlaylist {
			// Se tem loop, dá a volta e vai para a última música
			return totalMusicas - 1
		}
		// Se nao tem loop
		return 0
	}

	// Trata o comportamento se avançar além do fim da playlist
	if count >= totalMusicas {
		if ap.LoopPlaylist {
			// Se tem loop, volta para a primeira música
			return 0
		}
		// Se NÃO tem loop, trava na última música para evitar dar panic no slice
		return totalMusicas - 1
	}

	return count
}

func (ap *AudioPlayer) Play(songIndex int, queue []Musica) (imm int, end bool) {
	// f = a path for a song
	end = false
	imm = 1

	pathSong, extSong := queue[songIndex].Path, queue[songIndex].Ext

	f, err := os.Open(pathSong)
	if err != nil {
		log.Println("Erro ao abrir arquivo:", err)
		return
	}
	defer f.Close()

	var streamer beep.StreamSeekCloser
	var format beep.Format

	// Escolhe o decodificador baseado na extensão
	switch extSong {
	case ".mp3":
		streamer, format, err = mp3.Decode(f)
	case ".flac":
		streamer, format, err = flac.Decode(f)
	case ".wav":
		streamer, format, err = wav.Decode(f)
	default:
		log.Printf("extensão invalida '%s'", extSong)
		return
	}

	defer streamer.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Verifica se houve erro em qualquer uma das decodificações
	if err != nil {
		log.Printf("Erro ao decodificar %s: %v\n", extSong, err)
		return
	}

	originalStreamer := streamer
	var finalStreamer beep.Streamer = streamer

	if format.SampleRate != ap.SampleRate {
		finalStreamer = beep.Resample(3, format.SampleRate, ap.SampleRate, originalStreamer)
		// log.Println("Speaker resampled in %d hz", format.SampleRate.N)
	}

	speaker.Lock()
	ap.Ctrl.Streamer = finalStreamer
	speaker.Unlock()

	// trigger for the end of the function if the speaker.Play ends
	done := make(chan bool, 1) // arg "1" allows a channel with buffer

	speaker.Play(beep.Seq(ap.Volume, beep.Callback(func() {
		done <- true
	})))

	nextSong := true

MainLoop:
	for {
		select {
		case <-done:
			return

		case <-ticker.C:
			current := format.SampleRate.D(originalStreamer.Position())
			total := format.SampleRate.D(originalStreamer.Len())

			select {
			case ap.EventChan <- ProgressChanged{Current: current, Total: total}:
			default:
			}

		case song := <-ap.SelectSongChan:
			for i, s := range queue {
				if s.Path == song {
					ap.CurrentIndex = i
					break
				}
			}
			speaker.Clear()
			imm = 0
			break MainLoop

		case resp := <-ap.InputChan:

			// Switch case for player manipulation
			switch resp {
			// jump for the previous song
			case "p":
				nextSong = false
				speaker.Clear()
				break MainLoop

			// jump for the next song
			case "n":
				nextSong = true
				speaker.Clear()
				break MainLoop

			case "s":
				ap.PlayShuffle = !ap.PlayShuffle
				ap.EventChan <- ShuffleState{}

			case "q":
				end = true
				song := queue[songIndex]
				err := ap.StoreStatePlayer(song)
				if err != nil {
					log.Printf("Error: PlayerState json could'nt be created")
				}
				break MainLoop

			case "j":
				speaker.Lock()
				newPos := originalStreamer.Position() - ap.SampleRate.N(5*time.Second)

				if newPos < 0 {
					newPos = 0
				}

				err := originalStreamer.Seek(newPos)
				if err != nil {
					log.Fatal(err)
					log.Printf("Não avançou")
				}
				speaker.Unlock()

			case "l":
				speaker.Lock()
				newPos := originalStreamer.Position() + ap.SampleRate.N(5*time.Second)

				if newPos >= originalStreamer.Len() {
					newPos = originalStreamer.Len() - ap.SampleRate.N(time.Second)
				}
				err := originalStreamer.Seek(newPos)
				if err != nil {
					log.Fatal(err)
				}
				speaker.Unlock()

			}

		}
	}

	if !nextSong {
		imm = -1
	}

	return
}

func (ap *AudioPlayer) TogglePause() {
	speaker.Lock()
	ap.Ctrl.Paused = !ap.Ctrl.Paused
	speaker.Unlock()
}

func (ap *AudioPlayer) AddVolume(value float64) {
	speaker.Lock()
	newVolume := ap.Volume.Volume + value
	if newVolume <= 0 && newVolume >= -6 {
		ap.Volume.Volume = newVolume
	}
	speaker.Unlock()
}

func (ap *AudioPlayer) ToggleMute() {
	speaker.Lock()
	ap.Volume.Silent = !ap.Volume.Silent
	speaker.Unlock()
}

func GenerateShuffle(playlist []Musica) (shuffleList []Musica) {
	// Cria uma cópia para não modificar a playlist original
	shuffleList = make([]Musica, len(playlist))
	copy(shuffleList, playlist)

	// Algoritmo de Fisher-Yates
	for i := len(shuffleList) - 1; i > 0; i-- {
		// Sorteia um índice entre 0 e i
		j := rand.Intn(i + 1)

		// Troca os elementos de lugar (Go faz isso em uma linha!)
		shuffleList[i], shuffleList[j] = shuffleList[j], shuffleList[i]
	}

	return shuffleList
}

func (ap *AudioPlayer) StoreStatePlayer(song Musica) error {
	ap.PlayerState = PlayerState{
		PSlastTrackPath: song.Path,
		PSvolume:        ap.Volume.Volume, // Pega o volume atual do Beep
		PScurrentIndex:  ap.CurrentIndex,
		PSloopPlaylist:  ap.LoopPlaylist,
		PSloopSong:      ap.LoopSong,
		PSplayShuffle:   ap.PlayShuffle,
	}

	return nil
}
