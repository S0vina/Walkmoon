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

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/effects"
)

type Musica struct{
	Id int
	Path string
}
    
func main() {
	fmt.Println("WELCOME TO THE WALKMOON!")
	// Error: wrong args
	if len(os.Args) < 2 {
		fmt.Println("Uso: walkmoon <caminho_da_pasta>")
		return
	}

	// path of the folder with the musics
	folderPath := os.Args[1]

	// songs = [archives], err = any error
	playlist, err := scanFolder(folderPath)	

	// if err is not null, break the program
	if err != nil {
		log.Fatal(err)
	}

	// if folder with no songs, break the program
	if len(playlist) == 0 {
		fmt.Println("Nenhuma musica foi encontrada.")
		return
	}
	// debug quant of songs in the folder
	fmt.Println("You got %d musics", len(playlist))

	// init the speaker with the form of the file (SampleRate form)
	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Second/10))

	// for loop that play n songs of the folder
	count := 0
	for count < len(playlist){
		song := playlist[count]

		fmt.Printf("Tocando: [%d] %s\n", song.Id, filepath.Base(song.Path))
		aux := playAndWait(song.Path)
		count += aux
	}

	fmt.Println("Todas as musicas foram tocadas")
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
// function that creates a streamer with the actual file of the for loop and after that, playes it in
// a speaker object
func playAndWait(filePath string, sr int) int {
	// f = a file
	f, err := os.Open(filePath)
	if err != nil {
		log.Println("Erro ao abrir arquivo:", err)
		aux := 1
		return aux
	}
	defer f.Close()

	// decode the mp3 file
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Println("Erro ao decodificar mp3:", err)
		aux := 1
		return aux
	}
	defer streamer.Close()

	if format.SampleRate != sr {
		beep.Resample(3, format.SampleRate, sr, s)
		print("Speaker resampled in %d hz", format.SampleRate.N)
	} 

	// pause controller
	ctrl := &beep.Ctrl{Streamer: beep.Loop(1, streamer), Paused: false} 

	// volume controller
	volume := &effects.Volume{
		Streamer: ctrl,
		Base: 2,
		Volume: 0,
		Silent: false,
	}

	// allows the end of the function
	done := make(chan bool, 1) // arg "1" allows a channel with buffer

	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		done <- true
	})))

	// init a variable with a text scanner for input in the terminal
	scanner := bufio.NewScanner(os.Stdin)

	n_song := true
	break_for := false

	inputChan := make(chan string, 1)

	go func() {
		for scanner.Scan() {
            inputChan <- strings.TrimSpace(scanner.Text())
        }
	}()

	for {

		// Opcoes atuais de acao com o streamer volume
		fmt.Println("\nPress [p] to pause/resume")
		fmt.Println("Press [i] to increase volume")
		fmt.Println("Press [k] to decrease volume")
		fmt.Println("Press [m] to mute volume")
		fmt.Println("Press [l] to next song")
		fmt.Println("Press [j] to previous song")
		fmt.Print("-> ")
		select {
			case <-done:
				fmt.Println("Playing next song...")
				return 1

			case resp := <-inputChan: 
				// Switch case for player manipulation 
        		switch resp {
				// pause player
        		case "p":
        		    speaker.Lock()
        		    ctrl.Paused = !ctrl.Paused
        		    speaker.Unlock()
					if ctrl.Paused {
						fmt.Println("Pausado")
					} else{
						fmt.Println("Despausado")
					}
				
				// increase volume
        		case "i":
					if volume.Volume < 3{
						volume.Volume += 0.5
						fmt.Println("Volume atual: %f", volume.Volume)
						continue
					}
					fmt.Println("Volume maximo!!!")
					
				// decrease volume
        		case "k":
					if volume.Volume > -5 {
						volume.Volume += -0.5
						fmt.Println("Volume atual: %f", volume.Volume)
						continue
					}
					fmt.Println("Volume minimo!!!")
				
				// mute player
				case "m":
					volume.Silent = !volume.Silent
					if volume.Silent {
						fmt.Println("Mutado")
						continue
					}
					fmt.Println("Desmutado")
				
				// jump for the previous song
				case "j":
					n_song = false
					break_for = true
					speaker.Clear()
					break
				
				// jump for the next song
				case "l":
					n_song = true
					break_for = true
					speaker.Clear()
					break
				
				// case if the user press enter with nothing writed
        		case "":    
        		    continue
					
        		default:
        		    fmt.Printf("Comando '%s' não reconhecido\n", resp)
        		}
				
		}
		
		fmt.Println("Ainda estou aqui")

		if(break_for) {break}
	}

	aux := 1

	if (!n_song){
		aux = -1
	}  

	return aux
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