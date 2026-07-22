package main

import (
<<<<<<< HEAD
<<<<<<< HEAD
=======
	"encoding/json"
	"fmt"
>>>>>>> ab294f2 (Feature: now, walkmoon have a shortcut created by install.sh.)
	"log"
<<<<<<< HEAD
=======
	"os"
<<<<<<< HEAD
	"encoding/json"
>>>>>>> b14e0f9 (Feat: now player can be initiated from that last state (music, volume, paused) that its left)
=======
>>>>>>> ab294f2 (Feature: now, walkmoon have a shortcut created by install.sh.)
=======
	"log"
>>>>>>> 6cfe84f (Refact: Now, playerstate is created and updated in .config directory.)
	"path/filepath"

	"github.com/S0vina/walkmoon/internal/audio"
	"github.com/S0vina/walkmoon/internal/config"
	"github.com/S0vina/walkmoon/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	var ap *audio.AudioPlayer
	var state *audio.PlayerState
	var err error
	var musicDir string
	var playlist []audio.Musica
=======

=======
>>>>>>> ab294f2 (Feature: now, walkmoon have a shortcut created by install.sh.)
	var state *audio.PlayerState
	var defaultDir string
>>>>>>> b14e0f9 (Feat: now player can be initiated from that last state (music, volume, paused) that its left)

<<<<<<< HEAD
	state, musicDir, err = config.LoadState[audio.PlayerState]()
=======
	configPath := "../../assets/memory/playerState.json"

	jsonPlayerState, err := os.ReadFile(configPath)

	// caso o json ainda nao tenha sido criado
>>>>>>> ab294f2 (Feature: now, walkmoon have a shortcut created by install.sh.)
	if err != nil {
<<<<<<< HEAD
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

=======
		//home, errHome := os.UserHomeDir()
		//if errHome != nil {
		//	defaultDir = "assets/music"
		//	log.Printf("err home: %s", errHome)
		//}else {
		//	defaultDir = home
		//	log.Println(home)
		//}
=======
	var ap *audio.AudioPlayer
	var state *audio.PlayerState
	var err error
	var musicDir string
	var playlist []audio.Musica

	state, musicDir, err = config.LoadState[audio.PlayerState]()
	if err != nil {
		log.Printf("directory for musics has been loaded")
	}
>>>>>>> 6cfe84f (Refact: Now, playerstate is created and updated in .config directory.)

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

<<<<<<< HEAD
	// songs = [archives], err = any error
	playlist, errScan := ap.ScanFolder(defaultDir)

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
>>>>>>> b14e0f9 (Feat: now player can be initiated from that last state (music, volume, paused) that its left)
=======
>>>>>>> 6cfe84f (Refact: Now, playerstate is created and updated in .config directory.)
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
<<<<<<< HEAD
}
=======
}
>>>>>>> 6cfe84f (Refact: Now, playerstate is created and updated in .config directory.)
