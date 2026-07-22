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
	var state *audio.PlayerState
	var err error
	var musicDir string
	var playlist []audio.Musica

	state, musicDir, err = config.LoadState[audio.PlayerState]()
	if err != nil {
		log.Printf("directory for musics has been loaded")
	}

	ap, err = audio.New(state)
	if err != nil {
		log.Printf("Some problem ocurred in ap constructor", err)
	}

	if state != nil {
		musicDir = filepath.Dir(state.PSlastTrackPath)
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

	err = config.SaveState(ap.PlayerState)
	if err != nil {
		log.Printf("PlayerState could not be created", err)
	}
}
