package main
package audio

import (
	"fmt"
	"log"
	"os"

	"github.com/S0vina/walkmoon/internal/audio"
	"github.com/S0vina/walkmoon/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

type PlayerState struct {
	PSlastTrackPath string `json:"Path"`
	PSvolume int `json: "Volume"`
	PScurrentIndex int `json: "CurrentIndex"`
	PSloopPlaylist bool `json: "LoopPlaylist"`
	PSloopSong bool `json: "LoopSong"`
	PSplayShuffle bool `json: "PlayShuffle"`
}

func main() {
	// Error: wrong args
	// if len(os.Args) < 2 {
	// 	fmt.Println("Uso: walkmoon <caminho_da_pasta>")
	// 	return
	// }
	var state PlayerState
	var defaulDir string

	configPath := "assets/memory/playerState.json"

	jsonPlayerState, err := os.ReadFile(configPath)

	// caso o json ainda nao tenha sido criado
	if err != nil {
		home, errHome := os.UserHomeDir()
		if errHome != nil {
			defaulDir = "."
		}else {
			defaulDir = home
		}

		state = nil
	} else {
		err = json.Unmarshal(jsonPlayerState, &state)
		if err != nil {
			log.Fatalf("Erro ao decodificar o JSON: %v", err)
		}

		defaultDir = filepath.Dir(state.LastTrackPath)
	}

	_ = defaulDir

	ap, errNewAP := audio.New(state)
	if errNewAP != nil {
		log.Fatal(errNewAP)
	}

	// songs = [archives], err = any error
	playlist, errScan := ap.ScanFolder(defaulDir)

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

	telaInicial := ui.New(ap, playlist)
	programa := tea.NewProgram(telaInicial, tea.WithAltScreen(), tea.WithFPS(60))

	if _, err := programa.Run(); err != nil {
		log.Fatal("erro ao iniciar a interface gráfica:", err)
	}
}
