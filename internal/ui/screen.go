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
	corBranca   = lipgloss.Color("#FAFAFA")
)

// ===== ESTILOS =====
var (
	containerStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(corRoxa)

	playerStyle = lipgloss.NewStyle().
		Padding(1, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(corRoxa)

	titleStyle = lipgloss.NewStyle().
		Foreground(corBranca).
		Background(corRoxa).
		Bold(true).
		Padding(0, 1)

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

		margem := 20 // ← ajuste para controlar a largura do player
		m.playerWidth       = m.width - containerStyle.GetHorizontalFrameSize() - margem
		m.progressBar.Width = m.playerWidth - playerStyle.GetHorizontalFrameSize()
		return m, nil

	// FrameMsg é enviado internamente pela biblioteca para avançar a animação
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
		// SetPercent anima suavemente a barra até o novo valor
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
	if m.ShowPicker {
		return playerStyle.Width(m.playerWidth).Render(m.filepicker.View())
	}

	// --- Título ---
	titulo := titleStyle.Render("WALKMOON")

	// --- Música e estado ---
	nome := filepath.Base(m.CurrentSong.Path)
	if nome == "." || nome == "" {
		nome = "Carregando a fita..."
	}

	estadoTexto, cor := "⏸ TOCANDO", corVerde
	if m.player.Ctrl.Paused {
		estadoTexto, cor = "▶ PAUSADO", corVermelha
	}
	estado        := lipgloss.NewStyle().Foreground(cor).Bold(true).Render(estadoTexto)
	linhaPrincipal := fmt.Sprintf("%s  %s", estado, songStyle.Render(nome))

	// --- Volume ---
	mutado := ""
	if m.player.Volume.Silent {
		mutado = lipgloss.NewStyle().Foreground(corVermelha).Render(" [MUTADO]")
	}
	linhaVolume := fmt.Sprintf("volume: %.1f%s", m.player.Volume.Volume, mutado)

	// --- Shuffle ---
	shuffleTexto, shuffleCor := "ON", corVerde
	if !m.Shuffle {
		shuffleTexto, shuffleCor = "OFF", corVermelha
	}
	linhaShuffle := fmt.Sprintf("SHUFFLE: %s",
		lipgloss.NewStyle().Foreground(shuffleCor).Bold(true).Render(shuffleTexto))

	// --- Progresso ---
	linhaTempo := fmt.Sprintf("%s / %s",
		formatDuration(m.CurrentTime),
		formatDuration(m.TotalTime-m.CurrentTime))

	// --- Comandos ---
	picker   := commandStyle.Render("[f] selecionar música")
	comandos := commandStyle.Render("[space] play  •  [p]/[n] pular  •  [+]/[-] vol  •  [m] mute  •  [s] shuffle  •  [q] sair")

	// --- Layout ---
	conteudo := lipgloss.JoinVertical(lipgloss.Left,
		titulo,
		"",
		linhaPrincipal,
		linhaVolume,
		linhaShuffle,
		"",
		m.progressBar.View(),
		linhaTempo,
		"",
		picker,
		"",
		comandos,
	)

	player := playerStyle.Width(m.playerWidth).Render(conteudo)

	// Centraliza o player dentro do container que ocupa o terminal inteiro
	largura := m.width  - containerStyle.GetHorizontalFrameSize()
	altura  := m.height - containerStyle.GetVerticalFrameSize()

	centralizado := lipgloss.Place(largura, altura, lipgloss.Center, lipgloss.Center, player)

	return containerStyle.Width(largura).Height(altura).Render(centralizado)
}
