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
	Playlist		[]audio.Musica
	CurrentSong 	audio.Musica
	Shuffle			bool
	ShowPicker		bool
	filepicker		filepicker.Model
	width  int
	height int
}

func New(p *audio.AudioPlayer, playlist []audio.Musica) Model {
	fp := filepicker.New()
	fp.Height = 10
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.AllowedTypes = []string{".mp3"}

	return Model{
		player: p,
		filepicker: fp,
		Shuffle: false,
		ShowPicker: false,
		Playlist: playlist,
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

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

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
	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {	
		m.player.SelectSongChan <- path
		m.ShowPicker = false
	}
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

	picker := estiloComando.Render("[f] Para selecionar musica")
	// ===== COMANDOS =====
	comandos := estiloComando.Render(
		"[space] play    •    [p]/[n] pular    •    [+]/[-] vol\n\n[m] mute    •    [s] shuffle    •    [q] sair",
	)

	// ===== JUNTA TUDO =====
	conteudo := ""
	if(m.ShowPicker){
		conteudo += fmt.Sprintf("%s", m.filepicker.View())
	} else {
		conteudo += fmt.Sprintf(
			"%s\n\n%s\n%s\n%s\n\n\n\n%s\n\n%s",
			titulo,
			linhaPrincipal,
			linhaVolume,
			linhaShuffle,
			picker,
			comandos,
		)
	}

	// aplica borda final
	player := estiloBase.Render(conteudo)

	container := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD"))

	larguraInterna := m.width - container.GetHorizontalFrameSize()
	alturaInterna := m.height - container.GetVerticalFrameSize()

	centralizado := lipgloss.Place(
		larguraInterna,
		alturaInterna,
		lipgloss.Center,
		lipgloss.Center,
		player,
	)

	return container.
		Width(larguraInterna).
		Height(alturaInterna).
		Render(centralizado)
	
}
