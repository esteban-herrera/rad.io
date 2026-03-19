package player

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

const ipcSocket = "/tmp/rad.io.sock"

type Player struct {
	cmd    *exec.Cmd
	volume int
	muted  bool
	paused bool
}

func New() *Player {
	return &Player{volume: 100}
}

func (p *Player) Play(url string) error {
	p.Stop()
	p.paused = false
	args := []string{
		"--no-video", "--quiet", "--really-quiet",
		"--input-ipc-server=" + ipcSocket,
		fmt.Sprintf("--volume=%d", p.volume),
	}
	if p.muted {
		args = append(args, "--mute=yes")
	}
	args = append(args, url)
	p.cmd = exec.Command("mpv", args...)
	return p.cmd.Start()
}

func (p *Player) Stop() {
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		_ = p.cmd.Wait()
		p.cmd = nil
	}
	p.paused = false
}

func (p *Player) IsPlaying() bool {
	return p.cmd != nil && p.cmd.Process != nil
}

func (p *Player) sendIPC(args ...interface{}) ([]byte, error) {
	if !p.IsPlaying() {
		return nil, fmt.Errorf("not playing")
	}
	cmd := map[string]interface{}{"command": args}
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	var conn net.Conn
	for i := 0; i < 15; i++ {
		conn, err = net.Dial("unix", ipcSocket)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(300 * time.Millisecond))
	if _, err := conn.Write(data); err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	// mpv may send multiple lines; find the command response (skip event messages)
	for _, line := range strings.Split(string(buf[:n]), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, `"event"`) {
			continue
		}
		return []byte(line), nil
	}
	return buf[:n], nil
}

func (p *Player) TogglePause() {
	if !p.IsPlaying() {
		return
	}
	p.sendIPC("cycle", "pause") //nolint
	p.paused = !p.paused
}

func (p *Player) SetVolume(v int) {
	if v < 0 {
		v = 0
	}
	if v > 130 {
		v = 130
	}
	p.volume = v
	p.sendIPC("set_property", "volume", v) //nolint
}

func (p *Player) VolumeUp() {
	p.SetVolume(p.volume + 5)
}

func (p *Player) VolumeDown() {
	p.SetVolume(p.volume - 5)
}

func (p *Player) ToggleMute() {
	if !p.IsPlaying() {
		return
	}
	p.sendIPC("cycle", "mute") //nolint
	p.muted = !p.muted
}

func (p *Player) IsPaused() bool {
	return p.paused
}

func (p *Player) Volume() int {
	return p.volume
}

func (p *Player) IsMuted() bool {
	return p.muted
}

func (p *Player) NowPlayingMeta() string {
	if !p.IsPlaying() {
		return ""
	}
	resp, err := p.sendIPC("get_property", "metadata")
	if err != nil {
		return ""
	}

	var result struct {
		Data  map[string]interface{} `json:"data"`
		Error string                 `json:"error"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return ""
	}
	if result.Error != "" && result.Error != "success" {
		return ""
	}
	if result.Data == nil {
		return ""
	}

	if icyTitle, ok := result.Data["icy-title"].(string); ok && icyTitle != "" {
		return icyTitle
	}
	artist, _ := result.Data["artist"].(string)
	title, _ := result.Data["title"].(string)
	if artist != "" && title != "" {
		return artist + " - " + title
	}
	return title
}
