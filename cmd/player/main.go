package main

import (
	"fmt"
	"log"
	"os"

	"github.com/S0vina/walkmoon/internal/audio"
	"github.com/S0vina/walkmoon/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Error: wrong args
	if len(os.Args) < 2 {
		fmt.Println("Uso: walkmoon <caminho_da_pasta>")
		return
	}

	// path of the folder with the musics
	folderPath := os.Args[1]

	ap, errNewAP := audio.New()
	if errNewAP != nil {
		log.Fatal(errNewAP)
	}

	// songs = [archives], err = any error
	playlist, errScan := ap.ScanFolder(folderPath)

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
	go func() {
		ap.Run(playlist)
	}()

	telaInicial := ui.New(ap)
	programa := tea.NewProgram(telaInicial, tea.WithAltScreen())

	if _, err := programa.Run(); err != nil {
		log.Fatal("erro ao iniciar a interface gráfica:", err)
	}
}
