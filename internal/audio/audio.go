package audio

// imports
import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/flac"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/wav"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/dhowden/tag"
	tea "github.com/charmbracelet/bubbletea"
)

// struct: stores path and id(id stands for music number on the playlist)
type Musica struct {
	Id        int
	Path      string
	Ext 	  string	
	Title     string
	Artist    string
	Album     string
	Genre     string
	ImageData []byte
	ImageMIME string
}

// Centralizes all components that perdure in AudioPlayer Execution
type AudioPlayer struct {
	SampleRate   beep.SampleRate
	Ctrl         *beep.Ctrl
	Volume       *effects.Volume
	CurrentIndex int
	LoopPlaylist bool
	LoopSong     bool
	PlayShuffle  bool
	InputChan    chan string
	EventChan    chan tea.Msg
	SelectSongChan	chan string
	CurrentSong  Musica
}

type SongChanged struct{ Song Musica }
type ShuffleState struct {}

// method: creates a new audioPlayer
func New() (ap *AudioPlayer, err error) {
	sr := beep.SampleRate(44100)

	ctrl := &beep.Ctrl{Paused: false}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   -1,
		Silent:   false,
	}

	inputChan := make(chan string, 1)
	eventChan := make(chan tea.Msg, 5)
	selectChan := make(chan string, 1)

	CurrentIndex := 0

	loopPlaylist := true
	loopSong := false
	playShuffle := false
	speaker.Init(sr, sr.N(time.Second/10))

	ap = &AudioPlayer{
		SampleRate:   sr,
		Ctrl:         ctrl,
		Volume:       volume,
		CurrentIndex: CurrentIndex,
		LoopPlaylist: loopPlaylist,
		LoopSong:     loopSong,
		PlayShuffle:  playShuffle,
		InputChan:    inputChan,
    	EventChan:		eventChan,
		SelectSongChan:	selectChan,
	}

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
			ext := strings.ToLower(filepath.Ext(path)) // Pega a extensão (ex: ".mp3", ".flac")

			// Filtra apenas as extensões suportadas pelo seu player
			if ext == ".mp3" || ext == ".flac" || ext == ".wav" {
				
				// Abre o arquivo de áudio encontrado
				file, err := os.Open(path)
				if err != nil {
					return nil // Pula se der erro ao abrir
				}
				defer file.Close()

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

		log.Print(ap.CurrentSong.Artist)

		ap.EventChan <- SongChanged{Song: ap.CurrentSong} 
		imm, end := ap.Play(ap.CurrentIndex, queue)
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
		log.Println("Formato não suportado:", extSong)
		return
	}

	// Verifica se houve erro em qualquer uma das decodificações
	if err != nil {
		log.Printf("Erro ao decodificar %s: %v\n", extSong, err)
		return
	}
	defer streamer.Close()

	// A partir daqui você tem o 'streamer' e o 'format' prontos para mandar para o speaker!
	_ = format // Evita erro de variável não utilizada no exemplo

	if format.SampleRate != ap.SampleRate {
		beep.Resample(3, format.SampleRate, ap.SampleRate, streamer)
		log.Println("Speaker resampled in %d hz", format.SampleRate.N)
	}

	speaker.Lock()
	ap.SampleRate = format.SampleRate
	ap.Ctrl.Streamer = streamer
	speaker.Unlock()

	// trigger for the end of the function if the speaker.Play ends
	done := make(chan bool, 1) // arg "1" allows a channel with buffer

	speaker.Play(beep.Seq(ap.Volume, beep.Callback(func() {
		done <- true
	})))

	next_song := true

MainLoop:
	for {
		select {
		case <-done:
			return

		case song := <-ap.SelectSongChan:
			for i, s := range queue {
				if s.Path == song{
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
				next_song = false
				speaker.Clear()
				break MainLoop

			// jump for the next song
			case "n":
				next_song = true
				speaker.Clear()
				break MainLoop

			case "s":
				ap.PlayShuffle = !ap.PlayShuffle
				ap.EventChan <- ShuffleState{}

			case "q":
				end = true
				break MainLoop
			}
		}
	}

	if !next_song {
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
	if newVolume <= 3 && newVolume >= -5 {
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

// method: Changes the loopPlaylist bool value
// func (ap *AudioPlayer) ToggleLoopPlaylist() {
// 	ap.LoopPlaylist = !ap.LoopPlaylist
// }
