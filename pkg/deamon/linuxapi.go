//go:build linux || freebsd
// +build linux freebsd

package deamon

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"syscall"
)

const RPCPort = 3362

const (
	UnOpen = iota
	OpenFile
	Lock
	Destroy
)

const (
	ProcessStart = iota
	ProcessUpdate
	ProcessExit
)

const (
	CommandSuccess = 1
	CommandFailed  = 0
)

var initializers = make(map[string]func())

func Register(name string, initializer func()) {
	if _, exist := initializers[name]; exist {
		panic(fmt.Sprintf("exec %s already registered", name))
	}
	initializers[name] = initializer
}

func Init() bool {
	if initializer, exist := initializers[os.Args[0]]; exist {
		initializer()
		return true
	}
	return false
}

func Command(args ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: Self(),
		Args: args,
	}
}

func Self() string {
	name := os.Args[0]
	if name == filepath.Base(name) {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}

	if absPath, err := filepath.Abs(name); err != nil {
		return absPath
	}

	return name
}

type ProcessProtect struct {
	fileName string
	lockFile *os.File
	status   int
}

func (p *ProcessProtect) Init(name string) error {
	p.fileName = name
	lockFile, err := os.OpenFile(p.fileName, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Errorf("open lock file fail %s", err)
	}
	p.lockFile = lockFile
	p.status = OpenFile
	return nil
}

func (p *ProcessProtect) Try() bool {
	err := syscall.Flock(int(p.lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return false
	}
	p.status = Lock
	return true
}

func (p *ProcessProtect) UnLock() {
	syscall.Flock(int(p.lockFile.Fd()), syscall.LOCK_UN)
}

func (p *ProcessProtect) Close() {
	if p.status > OpenFile {
		p.UnLock()
	}
	if p.status > UnOpen {
		p.lockFile.Close()
		//os.Remove(p.fileName)
	}
	p.status = Destroy
}

type IPC struct {
	command chan int
}

func (ipc *IPC) Init(cmd chan int) {
	ipc.command = cmd
}

func (ipc *IPC) Command(request int, response *int) error {
	ipc.command <- request
	*response = CommandSuccess
	return nil
}

func CreateIPCRPCServer(cmd chan int, exit *int32) (net.Listener, error) {
	ipc := new(IPC)
	ipc.Init(cmd)
	if err := rpc.RegisterName("IPC", ipc); err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", RPCPort))
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				if atomic.LoadInt32(exit) > 0 {
					return
				} else {
					fmt.Println(err)
				}
			}
			go rpc.ServeConn(conn)
		}
	}()

	return listener, nil
}

func RPCRequest(cmd int) error {
	client, err := rpc.Dial("tcp", fmt.Sprintf("localhost:%d", RPCPort))
	if err != nil {
		return err
	}

	var reply int = -1
	err = client.Call("IPC.Command", cmd, &reply)
	if err != nil {
		return err
	}
	if reply == CommandSuccess {
		return nil
	}
	return fmt.Errorf("rpc command exec fail")
}
