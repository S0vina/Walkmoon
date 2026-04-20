package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"
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

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done // Bloqueia até a música terminar
}
