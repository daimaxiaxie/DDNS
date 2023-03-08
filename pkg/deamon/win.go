//go:build windows

package deamon

import (
	"DDNS/pkg/common"
	"encoding/json"
	"fmt"
	"time"
)

/* C Win API
//send message to taskbar
BOOL Shell_NotifyIconW(
  DWORD            dwMessage,  //what action
  PNOTIFYICONDATAW lpData      //action's info
);
*/

const (
	WindowName = "DDNS"
	ICONPath   = "D:/GolandProjects/DDNS/resources/dns.ico"
)

func Run(configData []byte) {
	t := Task{}
	if err := t.Init(configData); err != nil {
		panic(err)
	}

	var gui GUI
	if err := gui.Init(); err != nil {
		panic(err)
	}
	gui.Run(t.Run)
}

func Stop() {

}

type Task struct {
	config        *common.Config
	cloudProvider common.CloudProvider
}

func (t *Task) Init(data []byte) error {
	config := common.Config{}
	err := json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	cloudProvider, err := common.Manager.GetCloudProvider(config.Cloud, config.Version)
	if err != nil {
		return err
	}
	err = cloudProvider.Init(config.Extra)
	if err != nil {
		return err
	}

	t.config = &config
	t.cloudProvider = cloudProvider
	return nil
}

func (t *Task) Run(exit chan bool) {
	ticker := time.NewTicker(time.Duration(t.config.Duration) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-exit:
			return
		case <-ticker.C:
			if err := t.cloudProvider.Update(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
