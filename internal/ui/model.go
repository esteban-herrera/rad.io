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
	vizMode        int             // 0=bars 1=radio 2=dancer 3=wave 4=pulse 5=cradle
	themeIdx       int             // index into themes slice; 0 = Default
	expandedTags   map[string]bool // which tag sections are open
	showList       bool            // whether the station list is visible
	showHelp       bool            // whether key hints are shown
	width          int             // terminal width
	volChangedAt   int             // tickCount when vol last changed; -1 = never
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
		volChangedAt: -1,
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
		m.volChangedAt = m.tickCount

	case "-":
		m.player.VolumeDown()
		m.volChangedAt = m.tickCount

	case "m":
		m.player.ToggleMute()
		m.volChangedAt = m.tickCount

	case "v":
		m.vizMode = (m.vizMode + 1) % 6

	case "T":
		m.themeIdx = (m.themeIdx + 1) % len(themes)

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

func renderRadio(tickCount int, paused bool, style lipgloss.Style) string {
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
		sb.WriteString(style.Render("  "+content) + "\n")
	}
	return sb.String()
}

func renderDancer(tickCount int, paused bool, style lipgloss.Style) string {
	frames := [4][3]string{
		{` \o/`, `  |`, ` / \`},
		{`  o`, ` \|`, ` / \`},
		{`  o/`, `  |\`, ` / \`},
		{`  o`, `  |/`, ` / \`},
	}
	notes := [4]rune{'♪', '♫', '♩', '♬'}

	var frame [3]string
	var noteRow0, noteRow1 string
	if paused {
		frame = [3]string{`  o`, `  |`, ` / \`}
		noteRow1 = "  ⏸"
	} else {
		f := (tickCount / 6) % 4
		frame = frames[f]
		noteRow0 = " " + string(notes[f])
	}

	var sb strings.Builder
	for row := 0; row < 3; row++ {
		line := frame[row]
		if row == 0 {
			line += noteRow0
		} else if row == 1 {
			line += noteRow1
		}
		sb.WriteString(style.Render("  "+line) + "\n")
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

// renderWave draws a 2-row oscilloscope. Positive amplitude fills the top row,
// negative fills the bottom row.
func renderWave(tickCount, numBars int, style lipgloss.Style) string {
	t := float64(tickCount)
	topRow := make([]rune, numBars)
	botRow := make([]rune, numBars)
	for x := 0; x < numBars; x++ {
		fx := float64(x)
		amp := (math.Sin(t*0.18+fx*0.6) + math.Sin(t*0.09+fx*0.3+1.1)) / 2
		idx := int(math.Abs(amp) * 7.5)
		if idx > 7 {
			idx = 7
		}
		if amp > 0 {
			topRow[x] = vizBars[idx]
			botRow[x] = ' '
		} else {
			topRow[x] = ' '
			botRow[x] = vizBars[idx]
		}
	}
	return style.Render("  "+string(topRow)) + "\n" + style.Render("  "+string(botRow)) + "\n"
}

// renderPulse draws a single-row sonar ring expanding from the centre.
func renderPulse(tickCount, numBars int, paused bool, style lipgloss.Style) string {
	center := numBars / 2
	var sb strings.Builder
	if paused {
		for x := 0; x < numBars; x++ {
			if x == center {
				sb.WriteRune('⏸')
			} else {
				sb.WriteByte(' ')
			}
		}
	} else {
		radius := tickCount % (numBars / 2)
		for x := 0; x < numBars; x++ {
			dist := x - center
			if dist < 0 {
				dist = -dist
			}
			switch {
			case dist == radius && radius == 0:
				sb.WriteRune('●')
			case dist == radius:
				sb.WriteRune('◉')
			default:
				sb.WriteByte(' ')
			}
		}
	}
	return "  " + style.Render(sb.String()) + "\n"
}

// renderCradle draws a 2-row Newton's Cradle with non-uniform frame timing.
func renderCradle(tickCount int, paused bool, style lipgloss.Style) string {
	frames := [4][2]string{
		{` /\|||||`, `o  ooooo`},
		{`  \|||||`, ` ooooo `},
		{` \|||||\`, `ooooo  o`},
		{`  \|||||`, ` ooooo `},
	}
	var frame [2]string
	if paused {
		frame = frames[1]
		frame[0] += " ⏸"
	} else {
		phase := tickCount % 16
		var f int
		switch {
		case phase < 6:
			f = 0
		case phase < 8:
			f = 1
		case phase < 14:
			f = 2
		default:
			f = 3
		}
		frame = frames[f]
	}
	var sb strings.Builder
	sb.WriteString(style.Render("  "+frame[0]) + "\n")
	sb.WriteString(style.Render("  "+frame[1]) + "\n")
	return sb.String()
}

func (m Model) View() string {
	th := themes[m.themeIdx]
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
	div := th.divider.Render("  " + strings.Repeat("─", divLen))

	// Title + optional theme badge + optional filter badge
	b.WriteString(th.title.Render("  rad.io"))
	if th.Name != "Default" {
		b.WriteString("  " + th.help.Render("["+th.Name+"]"))
	}
	if m.tagFilter != "" {
		b.WriteString("  " + th.filter.Render("[#"+m.tagFilter+"]"))
	}
	b.WriteString("\n")
	b.WriteString(div + "\n")

	switch m.state {
	case stateList, stateEditTags:
		showingList := m.showList || !m.player.IsPlaying()

		if showingList {
			items := m.buildListItems()
			if len(m.stations) == 0 {
				b.WriteString(th.help.Render("  No stations. Press 'a' to add one.") + "\n")
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
							b.WriteString(th.selected.Render(line) + "\n")
						} else {
							b.WriteString(th.header.Render(line) + "\n")
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
						b.WriteString(th.selected.Render(prefix+name) + "\n")
					} else if isPlaying {
						b.WriteString(th.playing.Render(prefix+name) + "\n")
					} else {
						b.WriteString(th.normal.Render(prefix+name) + "\n")
					}
				}
			}
			b.WriteString(div + "\n")
		}

		// Visualization + playing info
		if m.player.IsPlaying() {
			switch m.vizMode {
			case 0:
				if m.player.IsPaused() {
					b.WriteString("  " + th.vizPause.Render("⏸ "+strings.Repeat(string(vizBars[0]), numBars)) + "\n")
				} else {
					b.WriteString("  " + th.viz.Render(renderViz(m.tickCount, numBars)) + "\n")
				}
			case 1:
				b.WriteString(renderRadio(m.tickCount, m.player.IsPaused(), th.radio))
			case 2:
				b.WriteString(renderDancer(m.tickCount, m.player.IsPaused(), th.dancer))
			case 3: // waveform — 2 rows; paused shows flat line
				if m.player.IsPaused() {
					b.WriteString("  " + th.vizPause.Render("⏸ "+strings.Repeat(string(vizBars[0]), numBars)) + "\n")
					b.WriteString("  " + th.vizPause.Render(strings.Repeat(" ", numBars)) + "\n")
				} else {
					b.WriteString(renderWave(m.tickCount, numBars, th.viz))
				}
			case 4: // pulse — 1 row expanding sonar ring
				b.WriteString(renderPulse(m.tickCount, numBars, m.player.IsPaused(), th.viz))
			case 5: // newton's cradle — 2 rows
				b.WriteString(renderCradle(m.tickCount, m.player.IsPaused(), th.dancer))
			}

			// Scrolling station name
			b.WriteString(th.playing.Render("  ♪ "+marquee(m.nowPlaying, nameMax, m.tickCount)) + "\n")

			// Vol line — only shown for ~2.4 s after last change
			if m.volChangedAt >= 0 && m.tickCount-m.volChangedAt < 30 {
				vol := m.player.Volume()
				volLine := fmt.Sprintf("vol:%s %d%%", renderVolBar(vol), vol)
				if m.player.IsMuted() {
					volLine += "  [muted]"
				}
				b.WriteString(th.playing.Render("  "+volLine) + "\n")
			}

			// Scrolling meta
			if m.nowPlayingMeta != "" {
				b.WriteString(th.meta.Render("  ♬ "+marquee(m.nowPlayingMeta, nameMax, m.tickCount)) + "\n")
			}
		}

		if m.errMsg != "" {
			b.WriteString(th.err.Render("  "+m.errMsg) + "\n")
		}

		if m.state == stateEditTags {
			b.WriteString(th.input.Render("  Tags (comma-separated):") + "\n")
			b.WriteString("  " + m.input.View() + "\n")
			b.WriteString(th.help.Render("  enter:save  esc:cancel") + "\n")
		} else if m.showHelp {
			if !showingList {
				b.WriteString(th.help.Render("  l:list  r:rnd  v:viz  T:theme  h:help  q:quit") + "\n")
				b.WriteString(th.help.Render("  space:pause  m:mute  +/-:vol  s:stop") + "\n")
			} else {
				b.WriteString(th.help.Render("  ↵:open/play  a:add  d:del  t:tags  f:filter") + "\n")
				b.WriteString(th.help.Render("  r:rnd  l:hide  v:viz  T:theme  space:pause  +/-:vol  q:quit") + "\n")
			}
		} else {
			b.WriteString(th.help.Render("  h:help") + "\n")
		}

	case stateAddName, stateAddURL:
		prompt := "Station name:"
		if m.state == stateAddURL {
			prompt = "Stream URL:"
		}
		b.WriteString(th.input.Render("  "+prompt) + "\n")
		b.WriteString("  " + m.input.View() + "\n")
		b.WriteString(th.help.Render("  enter:confirm  esc:cancel") + "\n")
	}

	return b.String()
}
