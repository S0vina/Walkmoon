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
	Id: int,
	Path: string
}

func main() {
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

	// for loop that play n songs of the folder
	counter := 0
	for counter < len(playlist){
		song := playlist[counter]

		fmt.Printf("Tocando: [%d] %s\n", song.Id, filepath.Base(song.Path))
		playAndWait(song)

		counter++
	}
}


// Function that scans the folder passed in the arg and return the files in it
func scanFolder(root string) ([]musica, error) {
	var playlist []musica
	contador := 0

	// cals the filepath.walk function that walks into the directory given as arg
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Filtra apenas arquivos .mp3 (case-insensitive)
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			musica := {Id : contador, Path: path}
			
			playlist = append(playlist, musica)
			contador++
		}
		return nil
	})

	return playlist, err
}
// function that creates a streamer with the actual file of the for loop and after that, playes it in
// a speaker object
func playAndWait(filePath string) {
	// f = a file
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

	// init the speaker with the form of the file (SampleRate form)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// init the ctrl with a beep controler (allows pauses and, rollback and rollfront (sei lá o nome disso))
	ctrl := &beep.Ctrl{Streamer: beep.Loop(1, streamer), Paused: false} 
	// int the volume variablie with ai struct that allows us change volume, mute, and desmute the speaker
	// It brings the before control functions
	volume := &effects.Volume{
		Streamer: ctrl,
		Base: 2,
		Volume: 0,
		Silent: false,
	}
	
	// init a variable with a text scanner for input in the terminal
	scanner := bufio.NewScanner(os.Stdin)

	// allows the end of the function
	done := make(chan bool)
	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		done <- true
	})))

	// for loop that allows use the scan for interact with the speaker until the songs end
	for {
		// Logica de contagem de tempo 
		// select {
		// case <-done:
		// 	return
		// case <-time.After(time.Second):
		// 	speaker.Lock()
		// 	fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
		// 	speaker.Unlock()
		// }

		// Opcoes atuais de acao com o streamer volume
		fmt.Println("\nPress [p] to pause/resume")
        fmt.Println("Press [i] to increase volume")
        fmt.Println("Press [k] to decrease volume")
		fmt.Println("Press [m] to mute volume")
		fmt.Println("Press [l] to next song")
		fmt.Println("Press [j] to previous song")
        fmt.Print("-> ")

        // Aguarda a entrada do usuário
        if !scanner.Scan() {
            break
        }

        // scanner.Text() cathcs the string. TrimSpace remove spaces and and \n
        resp := strings.TrimSpace(scanner.Text())

		// Switch case com as opcoes de manipulacao do speaker possiveis
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
			continue
		
		// jump for the next song
		case "l":
			continue
		
		// case if the user press enter with nothing writed
        case "":    
            continue
			
        default:
            fmt.Printf("Comando '%s' não reconhecido\n", resp)
        }
	}
}

// func decTypeArchive(f *os.File) {

// }