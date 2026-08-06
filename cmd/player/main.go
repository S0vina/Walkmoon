package main

import (
	"log"
	"path/filepath"

	"github.com/S0vina/walkmoon/internal/audio"
	"github.com/S0vina/walkmoon/internal/config"
	"github.com/S0vina/walkmoon/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var ap *audio.AudioPlayer
	var state *config.PlayerState
	var keybinds *config.PlayerKeybinds
	var err error
	var musicDir string
	var playlist []audio.Musica

	state, err = config.LoadConfig[config.PlayerState]("player_state")
	if err != nil {
		log.Printf("directory for musics has been loaded")
	}

	ap, err = audio.New(state, keybinds)
	if err != nil {
		log.Printf("Some problem ocurred in ap constructor", err)
	}

	if state != nil {
		musicDir = filepath.Dir(state.LastTrackPath)
		playlist, err = ap.ScanFolder(musicDir)
	} else {
		playlist, err = ap.ScanFolder(musicDir)
	}

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		ap.Run(playlist)
	}()

	telaInicial := ui.New(ap, playlist)
	programa := tea.NewProgram(telaInicial, tea.WithAltScreen(), tea.WithFPS(60))

	if _, err := programa.Run(); err != nil {
		log.Fatal("erro ao iniciar a interface gráfica:", err)
	}
}
