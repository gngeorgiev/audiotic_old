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

	"io/ioutil"

	"encoding/base64"

	"bytes"

	"encoding/json"

	"strconv"

	xj "github.com/basgys/goxml2json"
	"github.com/ddliu/go-httpclient"
	"github.com/go-errors/errors"
)

type VlcHttpCommand string
type VlcHttpArgument struct {
	Name, Value string
}

type VlcHttpPlayer struct {
	httpAddress, httpPort, httpPassword string

	authorizationHeader string

	source, name, state, thumbnail string
	duration, time                 int
	statsMutex                     sync.Mutex

	arguments []string

	isPlaying      bool
	isPlayingMutex sync.Mutex

	startedPlayingChan chan struct{}
	stoppedPlayingChah chan struct{}
	pausedPlayingChan  chan struct{}
	releaseChan        chan struct{}

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
	VlcCommandStop   = VlcHttpCommand("pl_stop")
	VlcCommandStatus = VlcHttpCommand("status")
)

func (v *VlcHttpPlayer) init() error {
	c := exec.Command("vlc", v.arguments...)
	if err := c.Start(); err != nil {
		return err
	}

	v.startedPlayingChan = make(chan struct{})
	v.stoppedPlayingChah = make(chan struct{})
	v.pausedPlayingChan = make(chan struct{})
	v.releaseChan = make(chan struct{})
	v.state = "stopped"

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

	go v.listenEvents()

	authBytes := []byte(fmt.Sprintf(":%s", v.httpPassword))
	authBase64 := base64.StdEncoding.EncodeToString(authBytes)
	v.authorizationHeader = fmt.Sprintf("Basic %s", authBase64)

	return <-httpInterfaceReadyChan
}

func (v *VlcHttpPlayer) urlBase() string {
	return fmt.Sprintf("http://%s/requests/status.xml", v.httpAddress+v.httpPort)
}

func (v *VlcHttpPlayer) url(command VlcHttpCommand, args ...VlcHttpArgument) string {
	if command == VlcCommandStatus {
		return v.urlBase()
	}

	u, _ := url.Parse(v.urlBase())
	q := u.Query()
	q.Set("command", string(command))

	for _, arg := range args {
		q.Set(arg.Name, arg.Value)
	}

	u.RawQuery = q.Encode()
	return u.String()
}

func (v *VlcHttpPlayer) executeRequest(url string) (*httpclient.Response, error) {
	return httpclient.
		WithHeader("Authorization", v.authorizationHeader).
		Get(url, nil)
}

func (v *VlcHttpPlayer) execCommand(command VlcHttpCommand, args ...VlcHttpArgument) error {
	commandUrl := v.url(command, args...)
	_, err := v.executeRequest(commandUrl)
	return err
}

func (v *VlcHttpPlayer) execCommandResponse(command VlcHttpCommand, args ...VlcHttpArgument) ([]byte, error) {
	commandUrl := v.url(command, args...)
	r, err := v.executeRequest(commandUrl)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()
	return ioutil.ReadAll(r.Body)
}

func (v *VlcHttpPlayer) IsPlaying() bool {
	v.isPlayingMutex.Lock()
	defer v.isPlayingMutex.Unlock()
	return v.isPlaying
}

func (v *VlcHttpPlayer) Resume() error {
	if !v.IsPlaying() {
		v.notifyStartPlaying()
		return v.execCommand(VlcCommandResume)
	}

	return nil
}

func (v *VlcHttpPlayer) listenEvents() {
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-t.C:
			if !v.IsPlaying() {
				continue
			}

			v.time += 100
			if v.time/1000 >= v.duration {
				go v.notifyStopPlaying()
			}
		case <-v.startedPlayingChan:
			v.isPlaying = true
			if v.state != "paused" {
				v.time = 0
			}
			v.state = "playing"
		case <-v.stoppedPlayingChah:
			v.state = "stopped"
			v.isPlaying = false
			v.time = 0
		case <-v.pausedPlayingChan:
			v.isPlaying = false
			v.state = "paused"
		case <-v.releaseChan:
			t.Stop()
			return
		}
	}
}

func (v *VlcHttpPlayer) notifyStartPlaying() {
	v.startedPlayingChan <- struct{}{}
}

func (v *VlcHttpPlayer) notifyStopPlaying() {
	v.stoppedPlayingChah <- struct{}{}
}

func (v *VlcHttpPlayer) notifyPausedPlaying() {
	v.pausedPlayingChan <- struct{}{}
}

func (v *VlcHttpPlayer) Play(source, name, thumbnail string, duration int) error {
	v.statsMutex.Lock()
	v.source = source
	v.name = name
	v.duration = duration
	v.statsMutex.Unlock()

	if err := v.execCommand(VlcCommandPlay, VlcHttpArgument{
		Name:  "input",
		Value: source,
	}); err != nil {
		return err
	}

	readyChan := make(chan error)
	go func() {
		t := time.NewTimer(30 * time.Second)

		for {
			select {
			case <-t.C:
				readyChan <- errors.New("Timeout playing track")
				v.Stop()
				return
			default:
				s, _ := v.statusInternal()
				if s.Root.State == "playing" && s.Root.Duration != "0" {
					v.notifyStartPlaying()
					t.Stop()
					readyChan <- nil
					return
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	return <-readyChan
}

func (v *VlcHttpPlayer) Pause() error {
	if v.IsPlaying() {
		v.notifyPausedPlaying()
		return v.execCommand(VlcCommandPause)
	}

	return nil
}

func (v *VlcHttpPlayer) Stop() error {
	if v.IsPlaying() {
		v.notifyStopPlaying()
		return v.execCommand(VlcCommandStop)
	}

	return nil
}

func (v *VlcHttpPlayer) Release() error {
	v.Stop()
	v.releaseChan <- struct{}{}
	return v.process.Kill()
}

type VlcStatusRoot struct {
	Duration  string `json:"length"`
	Time      string `json:"time"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	State     string `json:"state"`
	Thumbnail string `json:"thumbnail"`
}

type VlcStatus struct {
	Root *VlcStatusRoot `json:"root"`
}

func (v *VlcHttpPlayer) statusInternal() (*VlcStatus, error) {
	r, err := v.execCommandResponse(VlcCommandStatus)
	if err != nil {
		return nil, err
	}

	xml := bytes.NewReader(r)
	j, err := xj.Convert(xml)
	if err != nil {
		return nil, err
	}

	var status *VlcStatus
	if err := json.Unmarshal(j.Bytes(), &status); err != nil {
		return nil, err
	}

	return status, nil
}

func (v *VlcHttpPlayer) Status() (*VlcStatus, error) {
	status := &VlcStatus{Root: &VlcStatusRoot{}}
	status.Root.Name = v.name
	status.Root.Duration = strconv.Itoa(v.duration)
	status.Root.Source = v.source
	status.Root.Time = strconv.Itoa(v.time / 1000)
	status.Root.State = v.state
	status.Root.Thumbnail = v.thumbnail

	return status, nil
}

func Init() error {
	oncePlayer.Do(func() {
		httpAddress := "127.0.0.1"
		httpPort := ":8091"
		httpPassword := "rsd"

		player = &VlcHttpPlayer{
			httpAddress:  httpAddress,
			httpPort:     httpPort,
			httpPassword: httpPassword,
			arguments: []string{
				"--no-video",
				"--quiet",
				"--qt-start-minimized",
				"-I http",
				fmt.Sprintf("--http-port=%s", strings.Replace(httpPort, ":", "", 1)),
				fmt.Sprintf("--http-host=%s", httpAddress),
				fmt.Sprintf("--http-password=%s", httpPassword),
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
