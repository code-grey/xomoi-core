package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The Hacker Aesthetic: Neon Cyan and subtle grays.
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FFCC")).
		Border(lipgloss.RoundedBorder()).
		Padding(0, 2)

	subTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

type model struct {
	serverAddr string
	width      int
	height     int
	ticks      int
}

func InitialModel(server string) model {
	return model{serverAddr: server}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.ticks++
		// Skeleton: Here we will query the Go Backend API or connect to MQTT 
		// to pull real-time telemetry into the UI.
		return m, tickCmd()
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return "Initializing Xomoi Matrix..."
	}

	header := titleStyle.Render("XOMOI SOVEREIGN EDGE NODE")
	status := subTitleStyle.Render(fmt.Sprintf("Connected to: %s | Session Uptime: %ds", m.serverAddr, m.ticks))
	
	body := "\n\n[ LIVE TELEMETRY FEED ]\n"
	body += ">> Waiting for sensor data...\n"
	body += ">> [sys] Mochi-MQTT Broker active on :1883\n"
	body += ">> [sys] Ingestion Worker Pool standing by.\n"
	
	footer := "\n\nPress 'q' or 'ctrl+c' to exit."

	// Render the UI perfectly centered in the terminal
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, 
		lipgloss.JoinVertical(lipgloss.Center, header, status, body, footer))
}
