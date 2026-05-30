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
)

// struct: stores path and id(id stands for music number on the playlist)
type Musica struct{
	Id int
	Path string
}

// Centralizes all components that perdure in AudioPlayer Execution
type AudioPlayer struct{
	SampleRate beep.SampleRate
	Ctrl       *beep.Ctrl
	Volume     *effects.Volume
	InputChan 	chan string
	CurrentSong string
}

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

	speaker.Init(sr, sr.N(time.Second/10))

	ap = &AudioPlayer{
		SampleRate: sr,
		Ctrl: 		ctrl,
		Volume:     volume,
		InputChan: 	inputChan,
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
		print("Speaker resampled in %d hz", format.SampleRate.N)
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
	count := 0

	for count < len(playlist){
		song := playlist[count]

		ap.CurrentSong = song.Path
		imm, end := ap.Play(song.Path)
		if end {
			return
		}
		count += imm
	}
}

// method: plays the playlist in random mode
func (ap *AudioPlayer) PlayShuffle(playlist []Musica) {
	count := 0
	var played_songs[] int 

	for {
		// generate a random number for choose the next song
		var num int = rand.Intn(len(playlist))

		// trying to implement a queue of songs played
		// create a flag last_song. If it's true, than take the last number of the queue
		// and play that song. Make last_song false. And assim sucessivamente. 
		if len(played_songs) < 49 {
			played_songs = append(played_songs, num)
		}

		song := playlist[num]

		imm, end := ap.Play(song.Path)
		if end {
			return
		}

		count += imm
	}
}

