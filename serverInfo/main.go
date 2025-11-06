package serverInfo

import (
	"encoding/json"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sunvc/NoLets/serverInfo/cpu"
	"github.com/sunvc/NoLets/serverInfo/load"
	"github.com/sunvc/NoLets/serverInfo/mem"
	"github.com/sunvc/NoLets/serverInfo/net"
	"github.com/sunvc/NoLets/serverInfo/process"
)

var (
	monitPIDCPU   atomic.Value
	monitPIDRAM   atomic.Value
	monitPIDConns atomic.Value

	monitOSCPU      atomic.Value
	monitOSRAM      atomic.Value
	monitOSTotalRAM atomic.Value
	monitOSLoadAvg  atomic.Value
	monitOSConns    atomic.Value
)

type stats struct {
	PID statsPID `json:"pid"`
	OS  statsOS  `json:"os"`
}

type statsPID struct {
	CPU   float64 `json:"cpu"`
	RAM   uint64  `json:"ram"`
	Conns int     `json:"conns"`
}

type statsOS struct {
	CPU      float64 `json:"cpu"`
	RAM      uint64  `json:"ram"`
	TotalRAM uint64  `json:"total_ram"`
	LoadAvg  float64 `json:"load_avg"`
	Conns    int     `json:"conns"`
}

var (
	mutex sync.RWMutex
	once  sync.Once
	data  = &stats{}
)

func init() {
	once.Do(func() {
		p, _ := process.NewProcess(int32(os.Getpid())) //nolint:errcheck // TODO: Handle error
		numcpu := runtime.NumCPU()
		updateStatistics(p, numcpu)

		go func() {
			for {
				time.Sleep(10 * time.Second)

				updateStatistics(p, numcpu)
			}
		}()
	})
}

func updateStatistics(p *process.Process, numcpu int) {
	pidCPU, err := p.Percent(0)
	if err == nil {
		monitPIDCPU.Store(pidCPU / float64(numcpu))
	}

	if osCPU, err := cpu.Percent(0, false); err == nil && len(osCPU) > 0 {
		monitOSCPU.Store(osCPU[0])
	}

	if pidRAM, err := p.MemoryInfo(); err == nil && pidRAM != nil {
		monitPIDRAM.Store(pidRAM.RSS)
	}

	if osRAM, err := mem.VirtualMemory(); err == nil && osRAM != nil {
		monitOSRAM.Store(osRAM.Used)
		monitOSTotalRAM.Store(osRAM.Total)
	}

	if loadAvg, err := load.Avg(); err == nil && loadAvg != nil {
		monitOSLoadAvg.Store(loadAvg.Load1)
	}

	pidConns, err := net.ConnectionsPid("tcp", p.Pid)
	if err == nil {
		monitPIDConns.Store(len(pidConns))
	}

	osConns, err := net.Connections("tcp")
	if err == nil {
		monitOSConns.Store(len(osConns))
	}
}

func GetServerInfo() ([]byte, error) {
	mutex.Lock()
	data.PID.CPU, _ = monitPIDCPU.Load().(float64)
	data.PID.RAM, _ = monitPIDRAM.Load().(uint64)
	data.PID.Conns, _ = monitPIDConns.Load().(int)

	data.OS.CPU, _ = monitOSCPU.Load().(float64)
	data.OS.RAM, _ = monitOSRAM.Load().(uint64)
	data.OS.TotalRAM, _ = monitOSTotalRAM.Load().(uint64)
	data.OS.LoadAvg, _ = monitOSLoadAvg.Load().(float64)
	data.OS.Conns, _ = monitOSConns.Load().(int)
	mutex.Unlock()

	return json.Marshal(data)
}
