package ui

import (
	"fmt"
	"path/filepath"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/S0vina/walkmoon/internal/audio"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/filepicker"
)

// ===== CORES =====
var (
	corRoxa     = lipgloss.Color("#874BFD")
	corVerde    = lipgloss.Color("#04B575")
	corVermelha = lipgloss.Color("#FF0000")
	corCinza    = lipgloss.Color("#626262")
	// corBranca   = lipgloss.Color("#FAFAFA")
)

// ===== ESTILOS =====
var (
	playerStyle = lipgloss.NewStyle().
		Padding(1, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(corRoxa)	

	songStyle    = lipgloss.NewStyle().Foreground(corVerde).Bold(true)
	commandStyle = lipgloss.NewStyle().Foreground(corCinza)
)

// ===== MODEL =====
type Model struct {
	player      *audio.AudioPlayer
	Playlist    []audio.Musica
	CurrentSong audio.Musica
	progressBar progress.Model

	CurrentTime time.Duration
	TotalTime   time.Duration

	Shuffle     bool
	ShowPicker  bool
	filepicker  filepicker.Model
	playerWidth int
	width       int
	height      int
}

func New(p *audio.AudioPlayer, playlist []audio.Musica) Model {
	fp := filepicker.New()
	fp.SetHeight(10)
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.AllowedTypes = []string{".mp3"}

	return Model{
		player:     p,
		filepicker: fp,
		Playlist:   playlist,
		progressBar: progress.New(
			progress.WithGradient("#874BFD", "#04B575"),
			progress.WithoutPercentage(),
		),
	}
}

// ===== HELPERS =====

func EventListening(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

func formatDuration(d time.Duration) string {
	s := int(d.Seconds())
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}

// ===== INIT =====

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		EventListening(m.player.EventChan),
		m.filepicker.Init(),
	)
}

// ===== UPDATE =====

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width  = msg.Width
		m.height = msg.Height

		margem := 10
		m.playerWidth       = m.width - margem
		m.progressBar.Width = m.playerWidth - playerStyle.GetHorizontalFrameSize() - 10
		return m, nil

	case progress.FrameMsg:
		pm, c := m.progressBar.Update(msg)
		m.progressBar = pm.(progress.Model)
		return m, c

	case audio.ProgressChanged:
		m.CurrentTime = msg.Current
		m.TotalTime   = msg.Total
		var pct float64
		if m.TotalTime > 0 {
			pct = float64(m.CurrentTime) / float64(m.TotalTime)
		}
		return m, tea.Batch(EventListening(m.player.EventChan), m.progressBar.SetPercent(pct))

	case audio.SongChanged:
		m.CurrentSong = msg.Song
		return m, EventListening(m.player.EventChan)

	case audio.ShuffleState:
		m.Shuffle = !m.Shuffle
		return m, EventListening(m.player.EventChan)

	case tea.KeyMsg:
		switch msg.String() {
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
		case "n", "p", "s":
			m.player.InputChan <- msg.String()
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

// ===== VIEW =====

func (m Model) View() string {	
	nome := filepath.Base(m.CurrentSong.Path)
	if nome == "." || nome == "" {
		nome = "Carregando a fita..."
	}

	mutado := ""
	if m.player.Volume.Silent {
		mutado = lipgloss.NewStyle().Foreground(corVermelha).Render(" [MUTADO]")
	}

	shuffleTexto, shuffleCor := "ON", corVerde
	if !m.Shuffle {
		shuffleTexto, shuffleCor = "OFF", corVermelha
	}

	estadoTexto, estadoCor := "⏸ TOCANDO", corVerde
	if m.player.Ctrl.Paused {
		estadoTexto, estadoCor = "▶ PAUSADO", corVermelha
	}

	// --- Footer (renderizado primeiro para calcular altura) ---
	innerWidth := m.playerWidth - playerStyle.GetHorizontalFrameSize()
	colEsq     := innerWidth / 2
	colMeio    := innerWidth / 4
	colDir     := innerWidth - colEsq - colMeio

	footerEstado  := lipgloss.NewStyle().Foreground(estadoCor).Bold(true).Render(estadoTexto)
	footerNome    := songStyle.Render(nome)
	footerShuffle := fmt.Sprintf("SHUFFLE: %s",
		lipgloss.NewStyle().Foreground(shuffleCor).Bold(true).Render(shuffleTexto))
	footerVolume  := fmt.Sprintf("vol: %.1f%s", m.player.Volume.Volume, mutado)

	colunaEsq  := lipgloss.JoinVertical(lipgloss.Left, footerEstado, footerNome)
	colunaMeio := lipgloss.NewStyle().Width(colMeio).Align(lipgloss.Left).Render(footerShuffle)
	colunaDir  := lipgloss.NewStyle().Width(colDir).Align(lipgloss.Right).Render(footerVolume)

	linhaInfo := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(colEsq).Render(colunaEsq),
		colunaMeio,
		colunaDir,
	)

	linhaTempo := fmt.Sprintf("%s %s %s",
		formatDuration(m.CurrentTime),
		m.progressBar.View(),
		formatDuration(m.TotalTime))

	footer := playerStyle.Width(m.playerWidth).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			linhaInfo,
			linhaTempo,
		),
	)

	// --- Bloco principal (comandos) ---
	figlet := `
__        ___    _     _  ____  __  ___   ___  _   _ 
\ \      / / \  | |   | |/ /  \/  |/ _ \ / _ \| \ | |
 \ \ /\ / / _ \ | |   | ' /| |\/| | | | | | | |  \| |
  \ V  V / ___ \| |___| . \| |  | | |_| | |_| | |\  |
   \_/\_/_/   \_\_____|_|\_\_|  |_|\___/ \___/|_| \_|`

	titulo := lipgloss.NewStyle().Foreground(corRoxa).Bold(true).Render(figlet)

	tecla := func(k, icon, desc string) string {
		return fmt.Sprintf("%s  %s  %s",
			commandStyle.Render(k),
			lipgloss.NewStyle().Foreground(corRoxa).Render(icon),
			commandStyle.Render(desc),
		)
	}

	comandos := lipgloss.JoinVertical(lipgloss.Left,
		tecla("space  ", "⏸", "play/pause"),
		tecla("p / n  ", "󰒮  󰒭", "anterior / próxima"),
		tecla("+ / -  ", "󰕾  󰖁", "volume"),
		tecla("m      ", "󰝟", "mute"),
		tecla("s      ", "󰒝", "shuffle"),
		tecla("f      ", "󰢿", "selecionar arquivo"),
		tecla("q      ", "󰉈", "sair"),
	)

	alturaFooter   := lipgloss.Height(footer)
	alturaComandos := m.height - alturaFooter

	var tela string
	if m.ShowPicker {
		conteudo := lipgloss.JoinVertical(lipgloss.Left, m.filepicker.View())
		blocoFile := playerStyle.
			Width(m.playerWidth).
			Height(alturaComandos - playerStyle.GetVerticalFrameSize()).
			Render(conteudo)
		tela = lipgloss.Place(m.width, alturaComandos, lipgloss.Center, lipgloss.Center, blocoFile)
	} else {
		conteudo := lipgloss.JoinVertical(lipgloss.Left,
			titulo,
			"",
			comandos,
		)
		blocoComandos := playerStyle.
			Width(m.playerWidth).
			Height(alturaComandos - playerStyle.GetVerticalFrameSize()).
			Render(conteudo)
		tela = lipgloss.Place(m.width, alturaComandos, lipgloss.Center, lipgloss.Center, blocoComandos)
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		tela,
		footer,
	)
}
