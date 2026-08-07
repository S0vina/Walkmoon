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

	keybinds, err = config.LoadConfig[config.PlayerKeybinds]("player_keybinds")
	if err != nil {
		// Se o arquivo não existe ou deu erro, carregamos o padrão
		log.Println("Arquivo de atalhos não encontrado ou corrompido. Criando padrão...")
		keybinds = config.DefaultKeybinds()

		// Já aproveita e salva no disco para o usuário ter o arquivo lá e poder editar depois
		config.SaveConfig("player_keybinds", keybinds)
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

	log.Printf("programa fechou pelo x")
}
