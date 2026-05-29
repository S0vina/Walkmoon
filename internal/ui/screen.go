package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/S0vina/walkmoon/internal/audio" 
)

// ponteiro do audio player.
type Model struct {
	player *audio.AudioPlayer
}

func New(p *audio.AudioPlayer) Model {
	return Model{
		player: p,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		tecla := msg.String()

		switch tecla {

		case "q", "ctrl+c":
			m.player.InputChan <- "q"
			return m, tea.Quit
		case "p":
			m.player.TogglePause()
		case "+":
			m.player.AddVolume(0.5)
		case "-":
			m.player.AddVolume(-0.5)
		case "m":
			m.player.ToggleMute()
		case "l", "j":
			m.player.InputChan <- tecla
		}	
	}
	return m, nil
}

func (m Model) View() string {
	s := "\n  WALKMOON - Interface Terminal\n\n"
	
	if m.player.Ctrl.Paused {
		s += "  estado: [▶ PAUSADO]\n"
	} else {
		s += "  estado: [⏸ TOCANDO]\n"
	}

	s += fmt.Sprintf("  volume: %.1f\n", m.player.Volume.Volume)
	
	s += "\n  comandos:\n"
	s += "  [p] play/pause  |  [+]/[-] volume\n"
	s += "  [m] mute        |  [l]/[j] pular/voltar\n"
	s += "  [q] sair\n"
	
	return s
}
