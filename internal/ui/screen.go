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
	filepicker		filepicker.Model
}

func New(p *audio.AudioPlayer) Model {
	fp := filepicker.New()
	fp.Height = 10
	fp.CurrentDirectory, _ = os.UserHomeDir()

	return Model{
		player: p,
		filepicker: fp,
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
		}

	}
	m.filepicker, cmd = m.filepicker.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	titulo := estiloTitulo.Render("WALKMOON")

	musicName := filepath.Base(m.CurrentSong.Path)
	if musicName == "." || musicName == "" {
		musicName = "Carregando a fita..."
	}
	musicaRenderizada := estiloMusica.Render(musicName)

	estadoStr := "⏸ TOCANDO "
	corEstado := "#04B575" // verde

	if m.player.Ctrl.Paused {
		estadoStr = "▶ PAUSADO "
		corEstado = "#FF0000" // vermelho
	}
	estadoRenderizado := lipgloss.NewStyle().Foreground(lipgloss.Color(corEstado)).Bold(true).Render(estadoStr)

	linhaPrincipal := fmt.Sprintf("%s %s", estadoRenderizado, musicaRenderizada)

	// volume e status
	mutado := ""
	if m.player.Volume.Silent {
		mutado = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(" [MUTADO]")
	}
	linhaStatus := fmt.Sprintf("volume: %.1f%s", m.player.Volume.Volume, mutado)

	// rodapé de atalhos
	comandos := estiloComando.Render("[space] play  •  [p]/[n] pular  •  [+]/[-] vol  •  [m] mute  • [s]shuffle • [q] sair")

	// monta tudo pulando linhas
	uiLivre := fmt.Sprintf("%s\n\n%s\n%s\n\n%s", titulo, linhaPrincipal, linhaStatus, comandos)

	// envelopa tudo dentro da borda arredondada e joga na tela
	return estiloBase.Render(uiLivre)
}
