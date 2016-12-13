package player

import (
	"sync"

	"os/exec"

	"fmt"
	"strings"

	"log"
	"net/url"
	"os"
	"time"

	"github.com/ddliu/go-httpclient"
)

type VlcHttpCommand string
type VlcHttpArgument struct {
	Name, Value string
}

type VlcHttpPlayer struct {
	httpAddress string
	httpPort    string

	arguments []string

	isPlaying      bool
	isPlayingMutex sync.Mutex

	process *os.Process
}

var (
	oncePlayer sync.Once
	player     *VlcHttpPlayer
)

const (
	VlcCommandPlay   = VlcHttpCommand("in_play")
	VlcCommandPause  = VlcHttpCommand("pl_pause")
	VlcCommandResume = VlcHttpCommand("pl_play")
)

func (v *VlcHttpPlayer) init() error {
	c := exec.Command("vlc", v.arguments...)
	if err := c.Start(); err != nil {
		return err
	}

	v.process = c.Process

	httpInterfaceReadyChan := make(chan error)
	go func() {
		max := 20
		for i := 0; i < max; i++ {
			_, err := httpclient.Get(v.urlBase(), nil)
			if err != nil && i >= max-1 {
				httpInterfaceReadyChan <- err
			} else if err == nil {
				break
			}

			time.Sleep(1 * time.Second)
		}

		httpInterfaceReadyChan <- nil
	}()

	return <-httpInterfaceReadyChan
}

func (v *VlcHttpPlayer) urlBase() string {
	return fmt.Sprintf("http://%s/requests/status.xml", v.httpAddress+v.httpPort)
}

func (v *VlcHttpPlayer) url(command VlcHttpCommand, args ...VlcHttpArgument) string {
	u, _ := url.Parse(v.urlBase())
	q := u.Query()
	q.Set("command", string(command))

	for _, arg := range args {
		q.Set(arg.Name, arg.Value)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func (v *VlcHttpPlayer) execCommand(command VlcHttpCommand, args ...VlcHttpArgument) error {
	commandUrl := v.url(command, args...)
	_, err := httpclient.Get(commandUrl, nil)
	return err
}

func (v *VlcHttpPlayer) setIsPlaying(val bool) {
	v.isPlayingMutex.Lock()
	v.isPlaying = val
	v.isPlayingMutex.Unlock()
}

func (v *VlcHttpPlayer) IsPlaying() bool {
	v.isPlayingMutex.Lock()
	defer v.isPlayingMutex.Unlock()
	return v.isPlaying
}

func (v *VlcHttpPlayer) Resume() error {
	if !v.IsPlaying() {
		v.setIsPlaying(true)
		return v.execCommand(VlcCommandResume)
	}

	return nil
}

func (v *VlcHttpPlayer) Play(source string) error {
	defer v.setIsPlaying(true)
	return v.execCommand(VlcCommandPlay, VlcHttpArgument{
		Name:  "input",
		Value: source,
	})
}

func (v *VlcHttpPlayer) Pause() error {
	if v.IsPlaying() {
		v.setIsPlaying(false)
		return v.execCommand(VlcCommandPause)
	}

	return nil
}

func Init() error {
	oncePlayer.Do(func() {
		httpAddress := "127.0.0.1"
		httpPort := ":8091"

		player = &VlcHttpPlayer{
			httpAddress: httpAddress,
			httpPort:    httpPort,
			arguments: []string{
				"--no-video",
				"--quiet",
				"--qt-start-minimized",
				"-I http",
				fmt.Sprintf("--http-port=%s", strings.Replace(httpPort, ":", "", 1)),
				fmt.Sprintf("--http-host=%s", httpAddress),
			},
		}

		if err := player.init(); err != nil {
			log.Fatal(err)
		}
	})

	return nil
}

func Get() *VlcHttpPlayer {
	return player
}
