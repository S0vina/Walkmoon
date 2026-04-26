package main

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: walkmoon <caminho_da_pasta")
		return
	}

	forlderPaht := os.Args[1]

	songs, err := scanFolder(forlderPaht)

	if err != nil {
		log.Fatal(err)
	}

	if len(songs) == 0 {
		fmt.Println("Nenhuma musica foi encontrada.")
	}

	for _, song := range songs {
		fmt.Println("Tocando: %s\n", filepath.Base(song))
		playAndWait(song)
	}
}

func scanFolder(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Filtra apenas arquivos .mp3 (case-insensitive)
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func playAndWait(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Println("Erro ao abrir arquivo:", err)
		return
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Println("Erro ao decodificar mp3:", err)
		return
	}
	defer streamer.Close()

	// Inicializa o speaker com o formato da primeira música
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: false} 
	volume := &effects.Volume{
		Streamer: ctrl,
		Base: 2,
		Volume: 0,
		Silent: false,
	}
	
	scanner := bufio.NewScanner(os.Stdin)

	done := make(chan bool)
	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		done <- true
	})))


	
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
        fmt.Println("Press [d] to decrease volume")
		fmt.Println("Press [m] to mute volume")
        fmt.Print("-> ")

        // Aguarda a entrada do usuário
        if !scanner.Scan() {
            break
        }

        // scanner.Text() pega a string e strings.TrimSpace remove espaços e o \n
        resp := strings.TrimSpace(scanner.Text())

		// Switch case com as opcoes de manipulacao do speaker possiveis
        switch resp {
        case "p":
            speaker.Lock()
            ctrl.Paused = !ctrl.Paused
            speaker.Unlock()
			if ctrl.Paused {
				fmt.Println("Pausado")
			} else{
				fmt.Println("Despausado")
			}
            

        case "i":
			if volume.Volume < 3{
				volume.Volume += 0.5
				fmt.Println("Volume atual: %f", volume.Volume)
				continue
			}
			fmt.Println("Volume maximo!!!")
			

        case "d":
			if volume.Volume > -5 {
				volume.Volume += -0.5
				fmt.Println("Volume atual: %f", volume.Volume)
				continue
			}
			fmt.Println("Volume minimo!!!")
	
		
		case "m":
			volume.Silent = !volume.Silent
			if volume.Silent {
				fmt.Println("Mutado")
				continue
			}
			fmt.Println("Desmutado")

        case "":
            // Caso o usuário aperte Enter sem digitar nada, ignoramos
            continue
			
        default:
            fmt.Printf("Comando '%s' não reconhecido\n", resp)
        }
	}
}

// func decTypeArchive(f *os.File) {

// }
