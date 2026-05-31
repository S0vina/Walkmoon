package ui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/S0vina/walkmoon/internal/audio"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	player *audio.AudioPlayer
}

func New(p *audio.AudioPlayer) Model {
	return Model{
		player: p,
	}
}

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return doTick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tickMsg:
		return m, doTick()

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
			m.player.CurrentSong = "Carregando próxima..."
			m.player.InputChan <- tecla
		case "p":
			m.player.CurrentSong = "Carregando anterior..."
			m.player.InputChan <- tecla
		case "s":
			m.player.InputChan <- tecla
		}

	}
	return m, nil
}

func (m Model) View() string {
	titulo := estiloTitulo.Render("WALKMOON")

	musicName := filepath.Base(m.player.CurrentSong)
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
