package main

// imports
import (
	"fmt"
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"strconv"
	"math/rand"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/effects"
)

type Musica struct{
	Id int
	Path string
}

// Centralizes all components that perdure in AudioPlayer Execution
type audioPlayer struct{
	SampleRate beep.SampleRate
	Ctrl       *beep.Ctrl
	Volume     *effects.Volume
	InputChan 	chan string
	Scanner 	*bufio.Scanner
}

// Function that scans the folder passed in the arg and return the files in it
func scanFolder(root string) (playlist []Musica, err error) {
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

// func that creates a new audioPlayer
func NewAP () (ap *audioPlayer, err error){
	sr := beep.SampleRate(44100)

	ctrl := &beep.Ctrl{Paused: false}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base: 2,
		Volume: 0,
		Silent: false,
	}

	inputChan := make(chan string, 1)
	scanner := bufio.NewScanner(os.Stdin)

	speaker.Init(sr, sr.N(time.Second/10))

	ap = &audioPlayer{
		SampleRate: sr,
		Ctrl: 		ctrl,
		Volume:     volume,
		InputChan: 	inputChan,
		Scanner: 	scanner,
	}

	err = nil

	return
}

func playAndWait(filePath string, ap *audioPlayer) (aux int, end bool) {
	// f = a path for a song
	end = false
	aux = 1
	f, err := os.Open(filePath)
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
	break_for := false

	for {

		fmt.Print("\033[H\033[2J")

		fmt.Println("WELCOME TO THE WALKMOON!")
	  fmt.Println("------------------------")

		// Opcoes atuais de acao com o streamer volume
		fmt.Println("\nPress [p] to pause/resume")
		fmt.Println("Press [i] to increase volume")
		fmt.Println("Press [k] to decrease volume")
		fmt.Println("Press [m] to mute volume")
		fmt.Println("Press [l] to next song")
		fmt.Println("Press [j] to previous song")
		fmt.Println("Press [q] to quit")
		fmt.Print("-> ")
		select {
			case <-done:
				fmt.Println("Playing next song...")
				return 

			case resp := <-ap.InputChan: 
				// Switch case for player manipulation 
        		switch resp {
				// pause player
        		case "p":
        		    speaker.Lock()
        		    ap.Ctrl.Paused = !ap.Ctrl.Paused
        		    speaker.Unlock()
					if ap.Ctrl.Paused {
						fmt.Println("Pausado")
					} else{
						fmt.Println("Despausado")
					}
				
				// increase volume
        		case "i":
					if ap.Volume.Volume < 3{
						ap.Volume.Volume += 0.5
						fmt.Println("Volume atual: %f", ap.Volume.Volume)
						continue
					}
					fmt.Println("Volume maximo!!!")
					
				// decrease volume
        		case "k":
					if ap.Volume.Volume > -5 {
						ap.Volume.Volume += -0.5
						fmt.Println("Volume atual: %f", ap.Volume.Volume)
						continue
					}
					fmt.Println("Volume minimo!!!")
				
				// mute player
				case "m":
					ap.Volume.Silent = !ap.Volume.Silent
					if ap.Volume.Silent {
						fmt.Println("Mutado")
						continue
					}
					fmt.Println("Desmutado")
				
				// jump for the previous song
				case "j":
					next_song = false
					break_for = true
					speaker.Clear()
					break
				
				// jump for the next song
				case "l":
					next_song = true
					break_for = true
					speaker.Clear()
					break

				case "q":
					fmt.Println("Saindo........")
					break_for = true
					end = true
				
				// case if the user press enter with nothing writed
        		case "":    
        		    continue
					
        		default:
        		    fmt.Printf("Comando '%s' não reconhecido\n", resp)
        		}
				
		}

		if(break_for) {break}
	}

	if (!next_song){
		aux = -1
	}  

	return
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
    
func main() {
	// Error: wrong args
	if len(os.Args) < 2 {
		fmt.Println("Uso: walkmoon <caminho_da_pasta>")
		return
	}

	// path of the folder with the musics
	folderPath := os.Args[1]

	if os.Args[2] != "0" && os.Args[2] != "1"{
		fmt.Println("No format %s format exists", os.Args[2])
	}

	mode, errMode := strconv.Atoi(os.Args[2])
	if errMode != nil {
		log.Fatal(errMode)
	}

	ap, errNewAP := NewAP() 
	if errNewAP != nil {
		log.Fatal(errNewAP)
	}

	go func() {
		for ap.Scanner.Scan() {
			ap.InputChan <- strings.TrimSpace(ap.Scanner.Text())
		}
	}()

	// songs = [archives], err = any error
	playlist, errScan := scanFolder(folderPath)	

	// if ocurre some error in scanFolder func
	if errScan != nil {
		log.Fatal(errScan)
	}
	
	// if folder with no songs, break the program
	if len(playlist) == 0 {
		fmt.Println("No songs founded.")
		return
	}

	// for loop that play n songs of the folder
	count := 0

	if mode == 0 {
		for count < len(playlist){
			song := playlist[count]

			fmt.Printf("Tocando: [%d] %s\n", song.Id, filepath.Base(song.Path))
			aux, end := playAndWait(song.Path, ap)
			if end {
				return
			}
			count += aux
		}

	}else {
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

			fmt.Printf("Tocando: [%d] %s\n", song.Id, filepath.Base(song.Path))
			aux, end := playAndWait(song.Path, ap)
			if end {
				return
			}

			count += aux
		}
	}

	fmt.Println("Todas as musicas foram tocadas")
}
