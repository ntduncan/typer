package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	lipgloss "github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"

	"ntduncan.com/typer/styles"
	"ntduncan.com/typer/system"
	typetest "ntduncan.com/typer/type-test"
	"ntduncan.com/typer/utils"
)

type Model struct {
	cursor   int
	viewport viewport.Model
	test     typetest.TypeTest
}

var BestWPM string = ""
var isConfirmQuit bool = false

func InitModel(width int, height int, size int, mode utils.TestMode) Model {
	tt := typetest.New(size, mode)

	m := Model{
		cursor:   0,
		test:     tt,
		viewport: viewport.New(width, height),
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			if isConfirmQuit {

				config := system.Config{
					Size:     m.test.Size,
					Mode:     m.test.Mode,
					TopScore: BestWPM,
				}
				if err := system.SaveConfig(config); err != nil {
					panic(fmt.Sprintf("There was an error saving your configuration: %s", err))
				}

				return m, tea.Quit
			}
			isConfirmQuit = true

		case "tab":
			//restart
			m = InitModel(m.viewport.Width, m.viewport.Height, m.test.Size, m.test.Mode)
		case "+", "plus":
			var newSize int
			options := m.test.GetTestModeSizeOptions()

			for i, size := range options {
				if i == 3 {
					newSize = options[0]
					break
				}

				if m.test.Size == size {
					newSize = options[i+1]
					break
				}
			}

			m = InitModel(m.viewport.Width, m.viewport.Height, newSize, m.test.Mode)
		case "=":
			newMode := utils.WordsTest

			if m.test.Mode == utils.WordsTest {
				newMode = utils.TimeTest
			}

			options := utils.WordTestSizes
			if newMode == utils.TimeTest {
				options = utils.TimeTestSizes
			}

			m = InitModel(m.viewport.Width, m.viewport.Height, options[0], newMode)

		case "backspace":
			if m.cursor > 0 && m.cursor < len(m.test.Params) && (m.test.EndTime == time.Time{}) {
				m.test.Params[m.cursor].Input = ""
				m.test.Params[m.cursor].IsValid = false
				m.cursor--
			}

		default:

			if isConfirmQuit {
				if msg.String() == "y" || msg.String() == "Y" || msg.String() == "enter" {

					config := system.Config{
						Size:     m.test.Size,
						Mode:     m.test.Mode,
						TopScore: BestWPM,
					}
					if err := system.SaveConfig(config); err != nil {
						panic(fmt.Sprintf("There was an error saving your configuration: %s", err))
					}

					return m, tea.Quit
				}

				if msg.String() == "n" || msg.String() == "N" {
					isConfirmQuit = false
				}
			}

			if m.test.Mode == utils.TimeTest && m.test.StartTime.IsZero() {
				cmd = m.test.TestTimer.Start()
			}

			//Handle normal keypress
			if (m.test.EndTime.IsZero()) &&
				m.cursor != len(m.test.Params) {
				m.test.HandleKeyPress(msg.String(), m.cursor)
				if m.cursor != len(m.test.Params)-1 {
					m.cursor++
				}
			}
		}
	case tea.WindowSizeMsg:
		m.viewport.Height = msg.Height
		m.viewport.Width = msg.Width

	case timer.StartStopMsg:
		m.test.TestTimer, cmd = m.test.TestTimer.Update(msg)
	case timer.TickMsg:
		m.test.TestTimer, cmd = m.test.TestTimer.Update(msg)
	case timer.TimeoutMsg:
		m.test.TestTimer, cmd = m.test.TestTimer.Update(msg)
		m.test.EndTest()
	}

	return m, cmd
}

func (m Model) View() string {
	colors := styles.Colors

	title := lipgloss.NewStyle().
		Foreground(colors.Orange).
		Bold(true).
		BorderRight(true).
		BorderStyle(lipgloss.DoubleBorder()).
		Padding(0, 1, 0, 0).
		Margin(0, 1).
		Render("FunKeyType")

	wpm := m.test.GetWPM()
	wpmStyled := lipgloss.
		NewStyle().
		Align(lipgloss.Left).
		Padding(0, 1, 0, 0).
		Render("WPM: " + wpm)

	if wpm > BestWPM || BestWPM == "0.00" {
		BestWPM = wpm
	}

	bestWPMStyled := lipgloss.
		NewStyle().
		Foreground(colors.Orange).
		Render(BestWPM)

	testLen := lipgloss.NewStyle().
		Bold(true).
		BorderLeft(true).
		BorderStyle(lipgloss.DoubleBorder()).
		Padding(0, 1).
		Margin(0, 1).
		Render(m.test.GetTestSize())

	timer := lipgloss.
		NewStyle().
		Align(lipgloss.Left).
		Padding(0, 1, 0, 0).
		Render("| Time", m.test.TestTimer.View())

	subBar := "Top Score: " + bestWPMStyled + " | " + wpmStyled

	if m.test.Mode == utils.TimeTest {
		subBar += timer
	}

	subBar = lipgloss.
		NewStyle().
		Padding(0, 3).
		Render(subBar)

	mode := "Words"
	if m.test.Mode == utils.TimeTest {
		mode = "Timed"
	}

	modeStyled := lipgloss.NewStyle().
		Foreground(colors.Orange).
		Render(mode)

	topBar := lipgloss.
		NewStyle().
		Bold(true).
		Width(m.viewport.Width-2).
		Border(lipgloss.DoubleBorder()).
		Padding(0, 1).
		Render(title + "Mode: " + modeStyled + testLen)

	body := ""

	if isConfirmQuit {
		body = "Confirm Quit? [Y]es [N]o"
	} else {
		correct := lipgloss.NewStyle().Foreground(lipgloss.Color("#85DEAD"))
		incorrect := lipgloss.NewStyle().Foreground(colors.White).Background(lipgloss.Color(colors.Red))
		blockCursor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF")).Background(colors.Black)
		lineCursor := lipgloss.NewStyle().Underline(true)
		blank := lipgloss.NewStyle()

		for i, p := range m.test.Params {
			if i == m.cursor {
				if (m.test.EndTime != time.Time{}) {
					body += blockCursor.Render(p.Char)
				} else {
					body += lineCursor.Render(p.Char)
				}
				continue
			} else if p.IsValid {
				body += correct.Render(p.Char)
				continue
			} else if !p.IsValid && p.Input != "" {
				body += incorrect.Render(p.Char)
				continue
			} else {
				body += blank.Render(p.Char)
				continue
			}
		}
	}

	body = lipgloss.
		NewStyle().
		Height(m.viewport.Height-10).
		Width(m.viewport.Width-2).
		Align(lipgloss.Left).
		Padding(0, 10).
		Render(body)

	f := wordwrap.String(body, m.viewport.Width-10)
	m.viewport.SetContent(fmt.Sprintf("%v\n%v\n\n%v\n\n\n%v", topBar, subBar, f, m.footer()))

	return m.viewport.View()
}

func (m Model) footer() string {
	f := strings.Builder{}

	cmdMenu := lipgloss.
		NewStyle().
		Width(m.viewport.Width - 2).
		Border(lipgloss.DoubleBorder()).
		Foreground(lipgloss.Color("#A7A7A7")).
		Render("\"esc\": Exit | \"tab\": New Test | \"+\": Test Length | \"=\": Toggle Mode")

	f.WriteString(cmdMenu)

	return f.String()
}

func main() {
	config, configErr := system.LoadConfig()
	if configErr != nil {
		panic(configErr)
	}

	BestWPM = config.TopScore

	p := tea.NewProgram(InitModel(10, 10, config.Size, config.Mode), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Exited with error: %s", err)
		os.Exit(1)
	}
}
