package server_monitor

import (
	"github.com/jinzhu/copier"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_db"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_model"
	"github.com/siaikin/home-dashboard/internal/app/server_monitor/monitor_realtime"
	"github.com/siaikin/home-dashboard/internal/pkg/arguments"
	"log"
	"time"
)

func Initial() {
	if err := initialDeviceTable(); err != nil {
		panic(err)
	}
}

func Start(port int) {
	monitor_realtime.StartSystemRealtimeStatLoop(time.Second)

	startServer(port, arguments.MockMode)
}

func Stop() {
	monitor_realtime.StopSystemRealtimeStatLoop()
	stopServer(5 * time.Second)
}

func initialDeviceTable() error {
	db, err := monitor_db.OpenOrCreateDB()
	if err != nil {
		log.Printf("open db failed, %s", err)
		return err
	}

	systemStat := monitor_realtime.GetCachedSystemRealtimeStat()

	for _, v := range systemStat.Network {
		networkInfo := monitor_model.StoredSystemNetworkAdapterInfo{
			SystemNetworkAdapterInfo: monitor_realtime.SystemNetworkAdapterInfo{
				Index:       uint32(v.InterfaceStat.Index),
				Type:        v.InterfaceStat.Type,
				Description: v.InterfaceStat.Description,
			},
		}

		db.Where(map[string]interface{}{"Index": networkInfo.Index}).Attrs(networkInfo).FirstOrCreate(&networkInfo)
	}

	for _, v := range systemStat.Cpu {
		cpuInfo := monitor_model.StoredSystemCpuInfo{}
		if err := copier.Copy(&cpuInfo, &v.InfoStat); err != nil {
			log.Printf("cpoy failed, %s", err)
			return err
		}

		db.Where(map[string]interface{}{"CPU": cpuInfo.CPU}).Attrs(cpuInfo).FirstOrCreate(&cpuInfo)
	}

	for _, v := range systemStat.Disk {
		diskInfo := monitor_model.StoredSystemDiskInfo{}
		if err := copier.Copy(&diskInfo, &v.PartitionStat); err != nil {
			log.Printf("cpoy failed, %s", err)
			return err
		}

		db.Where(map[string]interface{}{"Mountpoint": diskInfo.Mountpoint}).Attrs(diskInfo).FirstOrCreate(&diskInfo)
	}

	return nil
}
