package ui

import (
	"fmt"
	"path/filepath"
	"os"

	"github.com/S0vina/walkmoon/internal/audio"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/filepicker"
)

var (
	estiloBase = lipgloss.NewStyle().
			Padding(1, 3).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD"))

	estiloTitulo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1)

	estiloMusica = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	estiloComando = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type Model struct {
	player 			*audio.AudioPlayer
	CurrentSong 	audio.Musica
	Shuffle			bool
	ShowPicker		bool
	filepicker		filepicker.Model
}

func New(p *audio.AudioPlayer) Model {
	fp := filepicker.New()
	fp.Height = 10
	fp.CurrentDirectory, _ = os.UserHomeDir()

	return Model{
		player: p,
		filepicker: fp,
		Shuffle: false,
		ShowPicker: false,
	}
}

func EventListening(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return <- ch
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		EventListening(m.player.EventChan),
		m.filepicker.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case audio.SongChanged:
		m.CurrentSong = msg.Song
		return m, EventListening(m.player.EventChan)
	case audio.ShuffleState:
		m.Shuffle = !m.Shuffle
		return m, EventListening(m.player.EventChan)

	case tea.KeyMsg:
		tecla := msg.String()

		switch tecla {

		case "q", "ctrl+c":
			m.player.InputChan <- "q"
			return m, tea.Quit
		case " ":
			m.player.TogglePause()
		case "+":
			m.player.AddVolume(0.5)
		case "-":
			m.player.AddVolume(-0.5)
		case "m":
			m.player.ToggleMute()
		case "n":
			m.player.InputChan <- tecla
		case "p":
			m.player.InputChan <- tecla
		case "s":
			m.player.InputChan <- tecla
		case "f":
			m.ShowPicker = !m.ShowPicker
		}

	}
	m.filepicker, cmd = m.filepicker.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	// ===== TÍTULO =====
	titulo := estiloTitulo.Render("WALKMOON")

	// ===== NOME DA MÚSICA =====
	nome := filepath.Base(m.CurrentSong.Path)

	// fallback se não tiver música carregada
	if nome == "." || nome == "" {
		nome = "Carregando a fita..."
	}

	musica := estiloMusica.Render(nome)

	// ===== ESTADO (PLAY / PAUSE) =====
	estadoTexto := "⏸ TOCANDO "
	cor := "#04B575" // verde

	if m.player.Ctrl.Paused {
		estadoTexto = "▶ PAUSADO "
		cor = "#FF0000" // vermelho
	}

	estado := lipgloss.NewStyle().
		Foreground(lipgloss.Color(cor)).
		Bold(true).
		Render(estadoTexto)

	// linha principal = estado + nome da música
	linhaPrincipal := fmt.Sprintf("%s %s", estado, musica)

	// ===== VOLUME =====
	mutado := ""

	// se estiver mudo, adiciona tag vermelha
	if m.player.Volume.Silent {
		mutado = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Render(" [MUTADO]")
	}

	linhaVolume := fmt.Sprintf("volume: %.1f%s",
		m.player.Volume.Volume,
		mutado,
	)

	shuffleState := "ON"
	color := "#04B575"
	if(!m.Shuffle){
		shuffleState = "OFF"
		color = "#FF0000"
	}

	shuffle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Render(shuffleState)


	linhaShuffle := fmt.Sprintf("SHUFFLE: %s", shuffle)

	// ===== COMANDOS =====
	comandos := estiloComando.Render(
		"[space] play  •  [p]/[n] pular  •  [+]/[-] vol  •  [m] mute  • [s] shuffle • [q] sair",
	)

	// ===== JUNTA TUDO =====
	conteudo := ""
	if(m.ShowPicker){
		conteudo += fmt.Sprintf("%s", m.filepicker.View())
	} else {
		conteudo += fmt.Sprintf(
			"%s\n\n%s\n%s\n%s\n\n%s",
			titulo,
			linhaPrincipal,
			linhaVolume,
			linhaShuffle,
			comandos,
		)
	}

	// aplica borda final
	return estiloBase.Render(conteudo)
}
