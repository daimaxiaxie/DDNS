//go:build linux || freebsd
// +build linux freebsd

package deamon

import (
	"DDNS/pkg/common"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const ChildProcess = "ddnsDeamon"
const LogName = "log"
const LockFileName = "/tmp/ddns.pid"

func CreateChildProcess() {
	t := time.Now().Local()
	name := fmt.Sprintf("%s-%d-%d-%02d%02d%02d", LogName, t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	log, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		fmt.Println(err)
	}
	args := make([]string, 0, 2)
	args = append(args, ChildProcess)
	args = append(args, os.Args[1:]...)
	cmd := Command(args...)
	cmd.Stdout = log
	cmd.Stderr = log
	if err := cmd.Start(); err != nil {
		panic(err)
	}
}

func Run(configData []byte) {
	//getppid != 1 (maybe)

	t := Task{}
	err := t.Init(configData)
	if err != nil {
		panic(err)
	}

	Register(ChildProcess, t.Run)
	if !Init() {
		processProtect := ProcessProtect{}
		if err := processProtect.Init(LockFileName); err != nil {
			panic(err)
		}
		defer processProtect.Close()

		if processProtect.Try() {
			CreateChildProcess()
			processProtect.UnLock()
		} else {
			fmt.Println("DDNS is already running")
		}

		//os.Exit(0)
		return
	}
}

func Stop() {
	err := RPCRequest(ProcessExit)
	if err != nil {
		panic(err)
	}
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

func (t *Task) Run() {
	processProtect := ProcessProtect{}
	if err := processProtect.Init(LockFileName); err != nil {
		panic(err)
	}
	defer processProtect.Close()

	var tryCount = 0
	for tryCount = 0; tryCount < 3; tryCount++ {
		if processProtect.Try() {
			break
		} else {
			time.Sleep(time.Duration(tryCount+1) * time.Second)
		}
	}
	if tryCount > 2 {
		panic("cannot get process lock")
	}
	defer processProtect.UnLock()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var client = make(chan int, 4)
	defer close(client)
	var exit int32 = 0
	rpc, err := CreateIPCRPCServer(client, &exit)
	if err != nil {
		panic(err)
	}
	defer rpc.Close()

	ticker := time.NewTicker(time.Duration(t.config.Duration) * time.Minute)

	for atomic.LoadInt32(&exit) == 0 {
		select {
		case <-signals:
			atomic.StoreInt32(&exit, 1)
		case c, ok := <-client:
			if !ok {
				return
			}
			switch c {
			case ProcessStart:

			}
			atomic.StoreInt32(&exit, 1)
		case <-ticker.C:
			if err := t.cloudProvider.Update(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
