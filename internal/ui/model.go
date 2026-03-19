package ui

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/esteban-herrera/rad.io/internal/player"
	"github.com/esteban-herrera/rad.io/internal/store"
)

type state int

const (
	stateList state = iota
	stateAddName
	stateAddURL
	stateEditTags
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	dividerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	playingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	inputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	vizStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	vizPauseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	metaStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	filterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	radioStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("178"))
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

var vizBars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

type listItem struct {
	isHeader bool
	tag      string        // set when isHeader=true
	station  store.Station // set when isHeader=false
	fullIdx  int           // index into m.stations
}

type Model struct {
	stations       []store.Station
	cursor         int
	state          state
	input          textinput.Model
	pendingName    string
	player         *player.Player
	nowPlaying     string
	errMsg         string
	tagFilter      string
	editingIdx     int
	tickCount      int
	ticking        bool
	nowPlayingMeta string
	vizMode        int             // 0 = bars (default), 1 = radio
	expandedTags   map[string]bool // which tag sections are open
	showList       bool            // whether the station list is visible
	showHelp       bool            // whether key hints are shown
	width          int             // terminal width
}

func New(stations []store.Station, p *player.Player) Model {
	ti := textinput.New()
	ti.CharLimit = 256
	return Model{
		stations:     stations,
		player:       p,
		input:        ti,
		expandedTags: map[string]bool{},
		showList:     true,
		showHelp:     false,
		width:        40,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) buildListItems() []listItem {
	var items []listItem

	if m.tagFilter != "" {
		items = append(items, listItem{isHeader: true, tag: m.tagFilter})
		if m.expandedTags[m.tagFilter] {
			for i, s := range m.stations {
				for _, t := range s.Tags {
					if t == m.tagFilter {
						items = append(items, listItem{station: s, fullIdx: i})
						break
					}
				}
			}
		}
		return items
	}

	// Group by sorted tags
	tags := m.allTags()
	tagToItems := map[string][]listItem{}
	for i, s := range m.stations {
		for _, t := range s.Tags {
			tagToItems[t] = append(tagToItems[t], listItem{station: s, fullIdx: i})
		}
	}
	for _, tag := range tags {
		items = append(items, listItem{isHeader: true, tag: tag})
		if m.expandedTags[tag] {
			items = append(items, tagToItems[tag]...)
		}
	}

	// Untagged section
	var untagged []listItem
	for i, s := range m.stations {
		if len(s.Tags) == 0 {
			untagged = append(untagged, listItem{station: s, fullIdx: i})
		}
	}
	if len(untagged) > 0 {
		items = append(items, listItem{isHeader: true, tag: "Untagged"})
		if m.expandedTags["Untagged"] {
			items = append(items, untagged...)
		}
	}

	return items
}

func findCursorForFullIdx(items []listItem, fullIdx int) int {
	for i, item := range items {
		if !item.isHeader && item.fullIdx == fullIdx {
			return i
		}
	}
	return -1
}

func (m Model) allTags() []string {
	seen := map[string]bool{}
	for _, s := range m.stations {
		for _, tag := range s.Tags {
			seen[tag] = true
		}
	}
	tags := make([]string, 0, len(seen))
	for t := range seen {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	return tags
}

// marquee returns a scrolling window into text. When text fits within width,
// it is returned as-is. Otherwise it scrolls at ~320ms per character.
func marquee(text string, width, tick int) string {
	if width <= 0 {
		return ""
	}
	r := []rune(text)
	if len(r) <= width {
		return text
	}
	r = append(r, ' ', ' ', ' ')
	total := len(r)
	offset := (tick / 4) % total
	var b strings.Builder
	for i := 0; i < width; i++ {
		b.WriteRune(r[(offset+i)%total])
	}
	return b.String()
}

// truncate clips s to max runes, appending … if cut.
func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			m.width = msg.Width
		}
		return m, nil

	case tickMsg:
		if m.player.IsPlaying() {
			m.tickCount++
			if m.tickCount%12 == 0 {
				m.nowPlayingMeta = m.player.NowPlayingMeta()
			}
			return m, tickCmd()
		}
		m.ticking = false
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case stateAddName, stateAddURL:
			return m.updateInput(msg)
		case stateEditTags:
			return m.updateEditTags(msg)
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errMsg = ""
	items := m.buildListItems()

	switch msg.String() {
	case "q", "ctrl+c":
		m.player.Stop()
		return m, tea.Quit

	case "h", "?":
		m.showHelp = !m.showHelp

	case "l":
		m.showList = !m.showList

	case "r":
		if len(m.stations) == 0 {
			break
		}
		s := m.stations[rand.Intn(len(m.stations))]
		if err := m.player.Play(s.URL); err != nil {
			m.errMsg = fmt.Sprintf("error: %v", err)
		} else {
			m.nowPlaying = s.Name
			m.nowPlayingMeta = ""
			m.showList = false
			if !m.ticking {
				m.ticking = true
				return m, tickCmd()
			}
		}

	case "s":
		m.player.Stop()
		m.nowPlaying = ""
		m.nowPlayingMeta = ""
		m.ticking = false
		m.showList = true

	case " ":
		m.player.TogglePause()

	case "+", "=":
		m.player.VolumeUp()

	case "-":
		m.player.VolumeDown()

	case "m":
		m.player.ToggleMute()

	case "v":
		m.vizMode = 1 - m.vizMode

	// list-only keys — ignored when list is hidden
	case "up", "k":
		if !m.showList {
			break
		}
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if !m.showList {
			break
		}
		if m.cursor < len(items)-1 {
			m.cursor++
		}

	case "enter":
		if !m.showList {
			m.showList = true
			break
		}
		if m.cursor >= len(items) {
			break
		}
		if items[m.cursor].isHeader {
			tag := items[m.cursor].tag
			m.expandedTags[tag] = !m.expandedTags[tag]
			newItems := m.buildListItems()
			if m.cursor >= len(newItems) {
				m.cursor = len(newItems) - 1
			}
			break
		}
		s := items[m.cursor].station
		if err := m.player.Play(s.URL); err != nil {
			m.errMsg = fmt.Sprintf("error: %v", err)
		} else {
			m.nowPlaying = s.Name
			m.nowPlayingMeta = ""
			m.showList = false
			if !m.ticking {
				m.ticking = true
				return m, tickCmd()
			}
		}

	case "a":
		if !m.showList {
			break
		}
		m.state = stateAddName
		m.input.Placeholder = "Station name"
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink

	case "d":
		if !m.showList || m.cursor >= len(items) || items[m.cursor].isHeader {
			break
		}
		fullIdx := items[m.cursor].fullIdx
		if m.nowPlaying == m.stations[fullIdx].Name {
			m.player.Stop()
			m.nowPlaying = ""
			m.nowPlayingMeta = ""
		}
		m.stations = append(m.stations[:fullIdx], m.stations[fullIdx+1:]...)
		newItems := m.buildListItems()
		if m.cursor >= len(newItems) && len(newItems) > 0 {
			m.cursor = len(newItems) - 1
		}
		if err := store.Save(m.stations); err != nil {
			m.errMsg = fmt.Sprintf("save error: %v", err)
		}

	case "t":
		if !m.showList || m.cursor >= len(items) || items[m.cursor].isHeader {
			break
		}
		fullIdx := items[m.cursor].fullIdx
		m.editingIdx = fullIdx
		m.state = stateEditTags
		m.input.Placeholder = "news, uk, ..."
		m.input.SetValue(strings.Join(m.stations[fullIdx].Tags, ", "))
		m.input.Focus()
		return m, textinput.Blink

	case "f":
		if !m.showList {
			break
		}
		tags := m.allTags()
		if m.tagFilter == "" {
			if len(tags) > 0 {
				m.tagFilter = tags[0]
			}
		} else {
			idx := -1
			for i, t := range tags {
				if t == m.tagFilter {
					idx = i
					break
				}
			}
			if idx < 0 || idx == len(tags)-1 {
				m.tagFilter = ""
			} else {
				m.tagFilter = tags[idx+1]
			}
		}
		if m.tagFilter != "" {
			m.expandedTags[m.tagFilter] = true
		}
		m.cursor = 0
	}
	return m, nil
}

func (m Model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.player.Stop()
		return m, tea.Quit

	case "esc":
		m.state = stateList
		m.input.Blur()
		return m, nil

	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			return m, nil
		}
		if m.state == stateAddName {
			m.pendingName = val
			m.state = stateAddURL
			m.input.Placeholder = "Stream URL"
			m.input.SetValue("")
			return m, textinput.Blink
		}
		// stateAddURL
		m.stations = append(m.stations, store.Station{
			Name: m.pendingName,
			URL:  val,
		})
		if err := store.Save(m.stations); err != nil {
			m.errMsg = fmt.Sprintf("save error: %v", err)
		}
		m.state = stateList
		m.input.Blur()
		m.expandedTags["Untagged"] = true
		newIdx := len(m.stations) - 1
		items := m.buildListItems()
		c := findCursorForFullIdx(items, newIdx)
		if c < 0 {
			c = 0
		}
		m.cursor = c
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateEditTags(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.player.Stop()
		return m, tea.Quit

	case "esc":
		m.state = stateList
		m.input.Blur()
		return m, nil

	case "enter":
		parts := strings.Split(m.input.Value(), ",")
		tags := make([]string, 0, len(parts))
		for _, p := range parts {
			t := strings.TrimSpace(p)
			if t != "" {
				tags = append(tags, t)
			}
		}
		m.stations[m.editingIdx].Tags = tags
		if err := store.Save(m.stations); err != nil {
			m.errMsg = fmt.Sprintf("save error: %v", err)
		}
		m.state = stateList
		m.input.Blur()
		if len(tags) > 0 {
			m.expandedTags[tags[0]] = true
		} else {
			m.expandedTags["Untagged"] = true
		}
		items := m.buildListItems()
		c := findCursorForFullIdx(items, m.editingIdx)
		if c < 0 {
			c = 0
		}
		m.cursor = c
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func renderViz(tickCount, numBars int) string {
	t := float64(tickCount)
	var sb strings.Builder
	for x := 0; x < numBars; x++ {
		fx := float64(x)
		h := (math.Sin(t*0.2+fx*0.8) + math.Sin(t*0.13+fx*0.4+1.5) + math.Sin(t*0.07+fx*1.2)) / 3.0
		idx := int((h+1.0)/2.0*7.0 + 0.5)
		if idx < 0 {
			idx = 0
		}
		if idx > 7 {
			idx = 7
		}
		sb.WriteRune(vizBars[idx])
	}
	return sb.String()
}

func renderRadio(tickCount int, paused bool) string {
	radioLines := [4]string{
		"┌─────────┐",
		"│ (( ◉ )) │",
		"│  ─────  │",
		"└─────────┘",
	}
	notes := [4]rune{'♫', '♪', '♩', '♬'}
	phases := [4]int{0, 4, 8, 2}

	var sb strings.Builder
	for row := 0; row < 4; row++ {
		var content string
		if paused {
			if row == 0 {
				content = radioLines[row] + " ⏸"
			} else {
				content = radioLines[row]
			}
		} else {
			noteOffset := ((tickCount + phases[row]) / 2) % 12
			content = radioLines[row] + strings.Repeat(" ", 1+noteOffset) + string(notes[row])
		}
		sb.WriteString(radioStyle.Render("  "+content) + "\n")
	}
	return sb.String()
}

func renderVolBar(vol int) string {
	const barLen = 6
	filled := int(float64(vol)/130.0*float64(barLen) + 0.5)
	if filled > barLen {
		filled = barLen
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)
}

func (m Model) View() string {
	var b strings.Builder

	// Derive sizing from terminal width
	nameMax := m.width - 4 // 2-char indent + 2-char prefix
	if nameMax < 4 {
		nameMax = 4
	}
	numBars := m.width - 2
	if numBars > 24 {
		numBars = 24
	}
	if numBars < 4 {
		numBars = 4
	}
	divLen := m.width - 2
	if divLen < 4 {
		divLen = 4
	}
	div := dividerStyle.Render("  " + strings.Repeat("─", divLen))

	// Title + optional filter badge
	b.WriteString(titleStyle.Render("  rad.io"))
	if m.tagFilter != "" {
		b.WriteString("  " + filterStyle.Render("[#"+m.tagFilter+"]"))
	}
	b.WriteString("\n")
	b.WriteString(div + "\n")

	switch m.state {
	case stateList, stateEditTags:
		showingList := m.showList || !m.player.IsPlaying()

		if showingList {
			items := m.buildListItems()
			if len(m.stations) == 0 {
				b.WriteString(helpStyle.Render("  No stations. Press 'a' to add one.") + "\n")
			} else {
				for i, item := range items {
					if item.isHeader {
						arrow := "▶ "
						if m.expandedTags[item.tag] {
							arrow = "▼ "
						}
						tag := truncate(item.tag, nameMax)
						line := "  " + arrow + tag
						if i == m.cursor {
							b.WriteString(selectedStyle.Render(line) + "\n")
						} else {
							b.WriteString(headerStyle.Render(line) + "\n")
						}
						continue
					}
					s := item.station
					isPlaying := m.nowPlaying == s.Name && m.player.IsPlaying()
					prefix := "    "
					if isPlaying {
						prefix = "  ▶ "
					}
					var name string
					if i == m.cursor {
						name = marquee(s.Name, nameMax, m.tickCount)
					} else {
						name = truncate(s.Name, nameMax)
					}
					if i == m.cursor {
						b.WriteString(selectedStyle.Render(prefix+name) + "\n")
					} else if isPlaying {
						b.WriteString(playingStyle.Render(prefix+name) + "\n")
					} else {
						b.WriteString(normalStyle.Render(prefix+name) + "\n")
					}
				}
			}
			b.WriteString(div + "\n")
		}

		// Visualization + playing info
		if m.player.IsPlaying() {
			if m.vizMode == 0 {
				if m.player.IsPaused() {
					b.WriteString("  " + vizPauseStyle.Render("⏸ "+strings.Repeat(string(vizBars[0]), numBars)) + "\n")
				} else {
					b.WriteString("  " + vizStyle.Render(renderViz(m.tickCount, numBars)) + "\n")
				}
			} else {
				b.WriteString(renderRadio(m.tickCount, m.player.IsPaused()))
			}

			// Scrolling station name
			b.WriteString(playingStyle.Render("  ♪ "+marquee(m.nowPlaying, nameMax, m.tickCount)) + "\n")

			// Vol line
			vol := m.player.Volume()
			volLine := fmt.Sprintf("vol:%s %d%%", renderVolBar(vol), vol)
			if m.player.IsMuted() {
				volLine += "  [muted]"
			}
			b.WriteString(playingStyle.Render("  "+volLine) + "\n")

			// Scrolling meta
			if m.nowPlayingMeta != "" {
				b.WriteString(metaStyle.Render("  ♬ "+marquee(m.nowPlayingMeta, nameMax, m.tickCount)) + "\n")
			}
		}

		if m.errMsg != "" {
			b.WriteString(errorStyle.Render("  "+m.errMsg) + "\n")
		}

		if m.state == stateEditTags {
			b.WriteString(inputStyle.Render("  Tags (comma-separated):") + "\n")
			b.WriteString("  " + m.input.View() + "\n")
			b.WriteString(helpStyle.Render("  enter:save  esc:cancel") + "\n")
		} else if m.showHelp {
			if !showingList {
				b.WriteString(helpStyle.Render("  l:list  r:rnd  v:viz  h:help  q:quit") + "\n")
				b.WriteString(helpStyle.Render("  space:pause  m:mute  +/-:vol  s:stop") + "\n")
			} else {
				b.WriteString(helpStyle.Render("  ↵:open/play  a:add  d:del  t:tags  f:filter") + "\n")
				b.WriteString(helpStyle.Render("  r:rnd  l:hide  v:viz  space:pause  +/-:vol  q:quit") + "\n")
			}
		} else {
			b.WriteString(helpStyle.Render("  h:help") + "\n")
		}

	case stateAddName, stateAddURL:
		prompt := "Station name:"
		if m.state == stateAddURL {
			prompt = "Stream URL:"
		}
		b.WriteString(inputStyle.Render("  "+prompt) + "\n")
		b.WriteString("  " + m.input.View() + "\n")
		b.WriteString(helpStyle.Render("  enter:confirm  esc:cancel") + "\n")
	}

	return b.String()
}
