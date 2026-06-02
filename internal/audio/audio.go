package audio

// imports
import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"math/rand"
	

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/effects"
	tea "github.com/charmbracelet/bubbletea"
)

// struct: stores path and id(id stands for music number on the playlist)
type Musica struct{
	Id int
	Path string
}

// Centralizes all components that perdure in AudioPlayer Execution
type AudioPlayer struct{
	SampleRate beep.SampleRate
	Ctrl       		*beep.Ctrl
	Volume     		*effects.Volume
	CurrentIndex 	int
	LoopPlaylist 	bool
	LoopSong		bool
	InputChan 		chan string
	EventChan		chan tea.Msg
	CurrentSong 	Musica
}

type SongChanged struct{ Song Musica }

// method: creates a new audioPlayer
func New() (ap *AudioPlayer, err error){
	sr := beep.SampleRate(44100)

	ctrl := &beep.Ctrl{Paused: false}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base: 2,
		Volume: 0,
		Silent: false,
	}

	inputChan := make(chan string, 1)
	eventChan := make(chan tea.Msg, 5)

	CurrentIndex := 0

	loopPlaylist := true
	loopSong := false
	speaker.Init(sr, sr.N(time.Second/10))

	ap = &AudioPlayer{
		SampleRate: 	sr,
		Ctrl: 			ctrl,
		Volume:     	volume,
		CurrentIndex: 	CurrentIndex,
		LoopPlaylist: 	loopPlaylist,
		LoopSong: 		loopSong,
		InputChan: 		inputChan,
		EventChan:		eventChan,
	}

	err = nil

	return
}

// method: scans the folder passed in the arg and return the files in it
func (ap *AudioPlayer) ScanFolder(root string) (playlist []Musica, err error) {
	contador := 0

	// cals the filepath.walk function that walks into the directory given as arg
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Filtra apenas arquivos .mp3 (case-insensitive)
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			musica := Musica{Id: contador, Path: path}
			
			playlist = append(playlist, musica)
			contador++

			// implementar a lógica de loop da folder com uma flag loop ativada por uma flag
			// no comando de terminal chamado -l
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
    

func (ap *AudioPlayer) Play(path_song string) (imm int, end bool ){
	// f = a path for a song
	end = false
	imm = 1
	f, err := os.Open(path_song)
	if err != nil {
		log.Println("Erro ao abrir arquivo:", err)
		return
	}
	defer f.Close()

	// decode the mp3 file
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Println("Erro ao decodificar mp3:", err)
		return
	}
	defer streamer.Close()

	if format.SampleRate != ap.SampleRate {
		beep.Resample(3, format.SampleRate, ap.SampleRate, streamer)
		// !!! log.Println("Speaker resampled in %d hz", format.SampleRate.N) LOOK THIS AFTER. IS NOT RESSAMPLING GOOD
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

			case resp := <-ap.InputChan: 
				// Switch case for player manipulation 
        		switch resp {		
					// jump for the previous song
					case "j":
						next_song = false
						speaker.Clear()
						break MainLoop
					
					// jump for the next song
					case "l":
						next_song = true
						speaker.Clear()
						break MainLoop

					case "q":
						end = true
						break MainLoop
        		}
		}
	}

	if (!next_song){
		imm = -1
	}  

	return
}

func (ap *AudioPlayer) TogglePause(){
	speaker.Lock()
	ap.Ctrl.Paused = !ap.Ctrl.Paused
	speaker.Unlock()
}

func(ap *AudioPlayer) AddVolume(value float64) {
	speaker.Lock()
	newVolume := ap.Volume.Volume + value
	if newVolume <= 3 && newVolume >= -5 {
		ap.Volume.Volume = newVolume
	}
	speaker.Unlock()
}

func(ap *AudioPlayer) ToggleMute() {
	speaker.Lock()
	ap.Volume.Silent = !ap.Volume.Silent
	speaker.Unlock()
}

// method: plays the playlist in order of songs
func (ap *AudioPlayer) PlaySequencial(playlist []Musica) {
	SequencialLoop:
	for ap.CurrentIndex < len(playlist){
		ap.CurrentSong = playlist[ap.CurrentIndex]
		count := 0

		ap.EventChan <- SongChanged{Song: ap.CurrentSong}

		imm, end := ap.Play(ap.CurrentSong.Path)
		if end {
			return
		}

		count += ap.CurrentIndex + imm

		// if the user loops in reverse the playlist
		if count < 0 {
			ap.CurrentIndex = len(playlist) - 1
			continue SequencialLoop
		}

		// if the playlist ends and loopplaylist is on
		if count > len(playlist) - 1 && ap.LoopPlaylist {
			ap.CurrentIndex = 0
			continue SequencialLoop
		} 
		// if playlist ends and no LoopPlaylist set
		if count > len(playlist) - 1 && !ap.LoopPlaylist {
			return
		}

		ap.CurrentIndex = count
	}
}

// method: plays the playlist in random mode
func (ap *AudioPlayer) PlayShuffle(playlist []Musica) {
	shuffleList := GenerateShuffle(playlist)
	//log.Println(shuffleList)
	ShuffleLoop:
	for ap.CurrentIndex < len(shuffleList){
		ap.CurrentSong = shuffleList[ap.CurrentIndex]
		count := 0

		ap.EventChan <- SongChanged{Song: ap.CurrentSong}
		imm, end := ap.Play(ap.CurrentSong.Path)
		if end {
			return
		}

		count += ap.CurrentIndex + imm

		// if the user loops in reverse the playlist
		if count < 0 {
			ap.CurrentIndex = len(shuffleList) - 1
			continue ShuffleLoop
		}

		// if the playlist ends and loopplaylist is on
		if count > len(shuffleList) - 1 && ap.LoopPlaylist {
			ap.CurrentIndex = 0
			continue ShuffleLoop
		} 
		// if playlist ends and no LoopPlaylist set
		if count > len(shuffleList) - 1 && !ap.LoopPlaylist {
			return
		}

		ap.CurrentIndex = count
	}
}

func GenerateShuffle(playlist []Musica) (shuffleList []Musica){
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
