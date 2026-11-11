package probe

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"

	"github.com/Haameed/f5_f5os_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

type ComponentsResponse struct {
	OpenconfigPlatformComponents `json:"openconfig-platform:components"`
}

type OpenconfigPlatformComponents struct {
	Component []Component `json:"component"`
}

type Component struct {
	Name       string             `json:"name"`
	Config     ComponentConfig    `json:"config,omitempty"`
	State      ComponentState     `json:"state,omitempty"`
	Storage    *Storage           `json:"storage,omitempty"`
	CPU        *CPU               `json:"cpu,omitempty"`
	FPGA       *IntegratedCircuit `json:"integrated-circuit,omitempty"`
	FanTray    *FanTray           `json:"f5-fan-psu-stats:fantray,omitempty"`
	PSUStats   *PSUStats          `json:"f5-fan-psu-stats:psu-stats,omitempty"`
	Properties *Properties        `json:"properties,omitempty"`
}

type ComponentConfig struct {
	Name string `json:"name"`
	Mode string `json:"f5-platform-lcd:mode,omitempty"`
}

type ComponentState struct {
	SerialNo    string            `json:"serial-no,omitempty"`
	PartNo      string            `json:"part-no,omitempty"`
	Empty       bool              `json:"empty,omitempty"`
	Description string            `json:"description,omitempty"`
	TPMStatus   string            `json:"f5-platform:tpm-integrity-status,omitempty"`
	Memory      *MemoryState      `json:"f5-platform:memory,omitempty"`
	Temperature *TemperatureState `json:"f5-platform:temperature,omitempty"`
}

// Memory & Temperature (unchanged)
type MemoryState struct {
	Available     string `json:"available"`
	Free          string `json:"free"`
	UsedPercent   int    `json:"used-percent"`
	PlatformTotal string `json:"platform-total"`
	PlatformUsed  string `json:"platform-used"`
}

type CPU struct {
	State CPUState `json:"state"`
}

type CPUState struct {
	Processors *Processors     `json:"f5-platform:processors,omitempty"`
	CPUUtil    *CPUUtilization `json:"f5-platform:cpu-utilization,omitempty"`
	CPUThreads *CPUThreads     `json:"f5-platform:cpu-threads,omitempty"`
}

type Processors struct {
	Processor []Processor `json:"processor"`
}

type Processor struct {
	CPUIndex int        `json:"cpu-index"`
	State    CPUDetails `json:"state"`
}

type CPUDetails struct {
	CacheSize string `json:"cachesize"`
	CoreCnt   string `json:"core-cnt"`
	ThreadCnt string `json:"thread-cnt"`
	Freq      string `json:"freq"`
}

type CPUUtilization struct {
	Current    int `json:"current"`
	FiveSecAvg int `json:"five-second-avg"`
	OneMinAvg  int `json:"one-minute-avg"`
	FiveMinAvg int `json:"five-minute-avg"`
}

type CPUThreads struct {
	CPUThread []CPUThread `json:"cpu-thread"`
}

type CPUThread struct {
	ThreadIndex int    `json:"thread-index"`
	Thread      string `json:"thread"`
	Current     int    `json:"current"`
	FiveSecAvg  int    `json:"five-second-avg"`
	OneMinAvg   int    `json:"one-minute-avg"`
	FiveMinAvg  int    `json:"five-minute-avg"`
}

type IntegratedCircuit struct {
	State ICState `json:"state"`
}

type ICState struct {
	FPGAs FPGAs `json:"f5-platform:fpgas"`
}

type FPGAs struct {
	FPGA []FPGA `json:"fpga"`
}

type FPGA struct {
	Index string `json:"fpga-index"`
	State struct {
		Version string `json:"version"`
	} `json:"state"`
}
type TemperatureState struct {
	Current json.Number `json:"current"`
	Average json.Number `json:"average"`
	Minimum json.Number `json:"minimum"`
	Maximum json.Number `json:"maximum"`
}

type Properties struct {
	Property []Property `json:"property"`
}

type Property struct {
	Name  string `json:"name"`
	State struct {
		Value string `json:"value"`
	} `json:"state"`
}
type FanTray struct {
	FanStats map[string]json.Number `json:"fan-stats"`
}

func (f *FanTray) GetFans() map[string]float64 {
	fans := make(map[string]float64)
	if f == nil || f.FanStats == nil {
		return fans
	}
	for k, v := range f.FanStats {
		if num, err := v.Float64(); err == nil {
			fans[k] = num
		}
	}
	return fans
}

type Storage struct {
	State StorageState `json:"state"`
}

type StorageState struct {
	Disks struct {
		Disk []Disk `json:"disk"`
	} `json:"f5-platform:disks"`
}

type Disk struct {
	DiskName string    `json:"disk-name"`
	State    DiskState `json:"state"`
}

type DiskState struct {
	Model    string  `json:"model"`
	Vendor   string  `json:"vendor"`
	SerialNo string  `json:"serial-no"`
	Size     string  `json:"size"`
	Type     string  `json:"type"`
	DiskIO   *DiskIO `json:"disk-io,omitempty"`
}

type DiskIO struct {
	TotalIOPS    json.Number `json:"total-iops"`
	ReadIOPS     json.Number `json:"read-iops"`
	ReadBytes    json.Number `json:"read-bytes"`
	ReadLatency  json.Number `json:"read-latency-ms"`
	WriteIOPS    json.Number `json:"write-iops"`
	WriteBytes   json.Number `json:"write-bytes"`
	WriteLatency json.Number `json:"write-latency-ms"`
	ReadMerged   json.Number `json:"read-merged"`
	WriteMerged  json.Number `json:"write-merged"`
}
type PSUStats struct {
	CurrentIn  string `json:"psu-current-in"`
	CurrentOut string `json:"psu-current-out"`
	VoltageIn  string `json:"psu-voltage-in"`
	VoltageOut string `json:"psu-voltage-out"`

	Temperature1 string `json:"psu-temperature-1"`
	Temperature2 string `json:"psu-temperature-2"`
	Temperature3 string `json:"psu-temperature-3"`
	Fan1Speed    int    `json:"psu-fan-1-speed"`
}

func GetHardwareProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		f5osInfo = prometheus.NewDesc(
			"bigip_f5os_info",
			"BigIP F5OS platform info",
			[]string{"target", "description", "tpm_status"}, nil,
		)
		memoryAvailable = prometheus.NewDesc(
			"bigip_f5os_memory_available_bytes",
			"BigIP F5OS memory available in bytes",
			[]string{"target"}, nil,
		)
		memoryfree = prometheus.NewDesc(
			"bigip_f5os_memory_free_bytes",
			"BigIP F5OS memory free in bytes",
			[]string{"target"}, nil,
		)
		memoryUsedRatio = prometheus.NewDesc(
			"bigip_f5os_memory_used_ratio",
			"BigIP F5OS memory used ratio (0-1)",
			[]string{"target"}, nil,
		)
		memoryPlatformUsed = prometheus.NewDesc(
			"bigip_f5os_memory_platform_used_bytes",
			"BigIP F5OS memory platform used in bytes",
			[]string{"target"}, nil,
		)
		memoryPlatformTotal = prometheus.NewDesc(
			"bigip_f5os_memory_platform_total_bytes",
			"BigIP F5OS memory platform total in bytes",
			[]string{"target"}, nil,
		)
		temperatureCurrent = prometheus.NewDesc(
			"bigip_f5os_temperature_current_celsius",
			"BigIP F5OS temperature current celsius",
			[]string{"target"}, nil,
		)
		temperatureAVG = prometheus.NewDesc(
			"bigip_f5os_temperature_average_celsius",
			"BigIP F5OS temperature average celsius",
			[]string{"target"}, nil,
		)
		temperatureMIN = prometheus.NewDesc(
			"bigip_f5os_temperature_minimum_celsius",
			"BigIP F5OS temperature minimum celsius",
			[]string{"target"}, nil,
		)
		temperatureMAX = prometheus.NewDesc(
			"bigip_f5os_temperature_maximum_celsius",
			"BigIP F5OS temperature maximum celsius",
			[]string{"target"}, nil,
		)
		storageDiskSizeBytes = prometheus.NewDesc(
			"bigip_f5os_storage_disk_size_bytes",
			"BigIP F5OS storage disk size in bytes",
			[]string{"target", "disk_name", "model", "serial", "type"}, nil,
		)
		// Note: these disk I/O metrics (IOPS, latency, merged-request counts)
		// are reported by F5OS as cumulative counters, not instantaneous
		// rates, despite the field names (see alerts/README.md). They keep
		// CounterValue/_total and are renamed to drop the misleading
		// rate-style units (_iops, _ms) in favor of _operations_total /
		// _latency_seconds_total.
		storageDiskTotalIOPS = prometheus.NewDesc(
			"bigip_f5os_storage_disk_operations_total",
			"BigIP F5OS cumulative total disk operations",
			[]string{"target", "disk_name"}, nil,
		)
		storageDiskReadIOPS = prometheus.NewDesc(
			"bigip_f5os_storage_disk_read_operations_total",
			"BigIP F5OS cumulative disk read operations",
			[]string{"target", "disk_name"}, nil,
		)
		storageDiskWriteIOPS = prometheus.NewDesc(
			"bigip_f5os_storage_disk_write_operations_total",
			"BigIP F5OS cumulative disk write operations",
			[]string{"target", "disk_name"}, nil,
		)
		storageReadMerged = prometheus.NewDesc(
			"bigip_f5os_storage_disk_read_merged_total",
			"BigIP F5OS cumulative merged disk read requests",
			[]string{"target", "disk_name"}, nil,
		)
		storageWriteMerged = prometheus.NewDesc(
			"bigip_f5os_storage_disk_write_merged_total",
			"BigIP F5OS cumulative merged disk write requests",
			[]string{"target", "disk_name"}, nil,
		)
		storageReadBytes = prometheus.NewDesc(
			"bigip_f5os_storage_disk_read_bytes_total",
			"BigIP F5OS cumulative disk bytes read",
			[]string{"target", "disk_name"}, nil,
		)
		storageWriteBytes = prometheus.NewDesc(
			"bigip_f5os_storage_disk_write_bytes_total",
			"BigIP F5OS cumulative disk bytes written",
			[]string{"target", "disk_name"}, nil,
		)
		storageReadLatency = prometheus.NewDesc(
			"bigip_f5os_storage_disk_read_latency_seconds_total",
			"BigIP F5OS cumulative disk read latency, in seconds",
			[]string{"target", "disk_name"}, nil,
		)
		storageWriteLatency = prometheus.NewDesc(
			"bigip_f5os_storage_disk_write_latency_seconds_total",
			"BigIP F5OS cumulative disk write latency, in seconds",
			[]string{"target", "disk_name"}, nil,
		)
		cpuCnt = prometheus.NewDesc(
			"bigip_f5os_cpu_count",
			"BigIP F5OS CPU count",
			[]string{"target"}, nil,
		)
		cpuThreadCnt = prometheus.NewDesc(
			"bigip_f5os_cpu_thread_count",
			"BigIP F5OS CPU thread count",
			[]string{"target"}, nil,
		)
		cpuFreq = prometheus.NewDesc(
			"bigip_f5os_cpu_frequency_hz",
			"BigIP F5OS CPU frequency in Hz",
			[]string{"target"}, nil,
		)
		cpuCacheSize = prometheus.NewDesc(
			"bigip_f5os_cpu_cache_size_bytes",
			"BigIP F5OS CPU cache size in bytes",
			[]string{"target"}, nil,
		)
		cpuUtilization = prometheus.NewDesc(
			"bigip_f5os_cpu_utilization_ratio",
			"BigIP F5OS CPU utilization ratio (0-1)",
			[]string{"target"}, nil,
		)
		cpuUtilization5SecAvg = prometheus.NewDesc(
			"bigip_f5os_cpu_utilization_5sec_avg_ratio",
			"BigIP F5OS CPU utilization 5 second average ratio (0-1)",
			[]string{"target"}, nil,
		)
		cpuUtilization1MinAvg = prometheus.NewDesc(
			"bigip_f5os_cpu_utilization_1min_avg_ratio",
			"BigIP F5OS CPU utilization 1 minute average ratio (0-1)",
			[]string{"target"}, nil,
		)
		cpuUtilization5MinAvg = prometheus.NewDesc(
			"bigip_f5os_cpu_utilization_5min_avg_ratio",
			"BigIP F5OS CPU utilization 5 minute average ratio (0-1)",
			[]string{"target"}, nil,
		)
		cpuThreadUtilizationCurrent = prometheus.NewDesc(
			"bigip_f5os_cpu_thread_utilization_ratio",
			"BigIP F5OS CPU thread utilization ratio (0-1)",
			[]string{"target", "thread_index", "thread_name"}, nil,
		)
		cpuThreadUtilization5SecAvg = prometheus.NewDesc(
			"bigip_f5os_cpu_thread_utilization_5sec_avg_ratio",
			"BigIP F5OS CPU thread utilization 5 second average ratio (0-1)",
			[]string{"target", "thread_index", "thread_name"}, nil,
		)
		cpuThreadUtilization1MinAvg = prometheus.NewDesc(
			"bigip_f5os_cpu_thread_utilization_1min_avg_ratio",
			"BigIP F5OS CPU thread utilization 1 minute average ratio (0-1)",
			[]string{"target", "thread_index", "thread_name"}, nil,
		)
		cpuThreadUtilization5MinAvg = prometheus.NewDesc(
			"bigip_f5os_cpu_thread_utilization_5min_avg_ratio",
			"BigIP F5OS CPU thread utilization 5 minute average ratio (0-1)",
			[]string{"target", "thread_index", "thread_name"}, nil,
		)
		fanSpeed = prometheus.NewDesc(
			"bigip_f5os_fan_speed_rpm",
			"BigIP F5OS fan speed in RPM",
			[]string{"target", "fan_name"}, nil,
		)
		psuCurrentIn = prometheus.NewDesc(
			"bigip_f5os_psu_current_in_amperes",
			"BigIP F5OS PSU input current, in amperes",
			[]string{"target", "psu_name"}, nil,
		)
		psuCurrentOut = prometheus.NewDesc(
			"bigip_f5os_psu_current_out_amperes",
			"BigIP F5OS PSU output current, in amperes",
			[]string{"target", "psu_name"}, nil,
		)
		psuVoltageIn = prometheus.NewDesc(
			"bigip_f5os_psu_voltage_in_volts",
			"BigIP F5OS PSU input voltage, in volts",
			[]string{"target", "psu_name"}, nil,
		)
		psuVoltageOut = prometheus.NewDesc(
			"bigip_f5os_psu_voltage_out_volts",
			"BigIP F5OS PSU output voltage, in volts",
			[]string{"target", "psu_name"}, nil,
		)
		psuTemperature1 = prometheus.NewDesc(
			"bigip_f5os_psu_temperature_1_celsius",
			"BigIP F5OS PSU temperature 1 in Celsius",
			[]string{"target", "psu_name"}, nil,
		)
		psuTemperature2 = prometheus.NewDesc(
			"bigip_f5os_psu_temperature_2_celsius",
			"BigIP F5OS PSU temperature 2 in Celsius",
			[]string{"target", "psu_name"}, nil,
		)
		psuTemperature3 = prometheus.NewDesc(
			"bigip_f5os_psu_temperature_3_celsius",
			"BigIP F5OS PSU temperature 3 in Celsius",
			[]string{"target", "psu_name"}, nil,
		)
		psuFan1Speed = prometheus.NewDesc(
			"bigip_f5os_psu_fan_1_speed_rpm",
			"BigIP F5OS PSU fan 1 speed in RPM",
			[]string{"target", "psu_name"}, nil,
		)
	)

	var m []prometheus.Metric

	var resp ComponentsResponse
	if err := c.Get("/api/data/openconfig-platform:components", &resp); err != nil {
		log.Printf("Hardware stats unavailable for %s. error was %v", target, err)
	}

	for _, comp := range resp.Component {
		switch comp.Name {
		case "platform":
			m = append(m, prometheus.MustNewConstMetric(f5osInfo, prometheus.GaugeValue, 0, target, comp.State.Description, comp.State.TPMStatus))
			m = append(m, prometheus.MustNewConstMetric(memoryAvailable, prometheus.GaugeValue, stringToFloat(comp.State.Memory.Available), target))
			m = append(m, prometheus.MustNewConstMetric(memoryfree, prometheus.GaugeValue, stringToFloat(comp.State.Memory.Free), target))
			m = append(m, prometheus.MustNewConstMetric(memoryUsedRatio, prometheus.GaugeValue, float64(comp.State.Memory.UsedPercent)/100, target))
			m = append(m, prometheus.MustNewConstMetric(memoryPlatformUsed, prometheus.GaugeValue, stringToFloat(comp.State.Memory.PlatformUsed), target))
			m = append(m, prometheus.MustNewConstMetric(memoryPlatformTotal, prometheus.GaugeValue, stringToFloat(comp.State.Memory.PlatformTotal), target))
			m = append(m, prometheus.MustNewConstMetric(temperatureCurrent, prometheus.GaugeValue, stringToFloat(comp.State.Temperature.Current.String()), target))
			m = append(m, prometheus.MustNewConstMetric(temperatureAVG, prometheus.GaugeValue, stringToFloat(comp.State.Temperature.Average.String()), target))
			m = append(m, prometheus.MustNewConstMetric(temperatureMIN, prometheus.GaugeValue, stringToFloat(comp.State.Temperature.Minimum.String()), target))
			m = append(m, prometheus.MustNewConstMetric(temperatureMAX, prometheus.GaugeValue, stringToFloat(comp.State.Temperature.Maximum.String()), target))
			for _, disk := range comp.Storage.State.Disks.Disk {
				// disk.State.Size is reported as a string like "700.00GB";
				// stringToFloat strips the unit suffix, so multiply by 1e9
				// to convert decimal gigabytes to the Prometheus base unit
				// (bytes).
				m = append(m, prometheus.MustNewConstMetric(storageDiskSizeBytes, prometheus.GaugeValue, stringToFloat(disk.State.Size)*1e9, target, disk.DiskName, disk.State.Model, disk.State.SerialNo, disk.State.Type))
				m = append(m, prometheus.MustNewConstMetric(storageDiskTotalIOPS, prometheus.CounterValue, stringToFloat(disk.State.DiskIO.TotalIOPS.String()), target, disk.DiskName))
				m = append(m, prometheus.MustNewConstMetric(storageDiskReadIOPS, prometheus.CounterValue, stringToFloat(disk.State.DiskIO.ReadIOPS.String()), target, disk.DiskName))
				m = append(m, prometheus.MustNewConstMetric(storageDiskWriteIOPS, prometheus.CounterValue, stringToFloat(disk.State.DiskIO.WriteIOPS.String()), target, disk.DiskName))
				m = append(m, prometheus.MustNewConstMetric(storageReadBytes, prometheus.CounterValue, stringToFloat(disk.State.DiskIO.ReadBytes.String()), target, disk.DiskName))
				m = append(m, prometheus.MustNewConstMetric(storageWriteBytes, prometheus.CounterValue, stringToFloat(disk.State.DiskIO.WriteBytes.String()), target, disk.DiskName))
				// read/write latency are reported in milliseconds; divide by
				// 1000 to expose the Prometheus base unit (seconds).
				m = append(m, prometheus.MustNewConstMetric(storageReadLatency, prometheus.CounterValue, stringToFloat(disk.State.DiskIO.ReadLatency.String())/1000, target, disk.DiskName))
				m = append(m, prometheus.MustNewConstMetric(storageWriteLatency, prometheus.CounterValue, stringToFloat(disk.State.DiskIO.WriteLatency.String())/1000, target, disk.DiskName))
				// read-merged/write-merged are cumulative counters (see
				// alerts/README.md); fixed from GaugeValue to CounterValue.
				m = append(m, prometheus.MustNewConstMetric(storageReadMerged, prometheus.CounterValue, stringToFloat(string(disk.State.DiskIO.ReadMerged)), target, disk.DiskName))
				m = append(m, prometheus.MustNewConstMetric(storageWriteMerged, prometheus.CounterValue, stringToFloat(string(disk.State.DiskIO.WriteMerged)), target, disk.DiskName))
			}
			m = append(m, prometheus.MustNewConstMetric(cpuCnt, prometheus.GaugeValue, stringToFloat(comp.CPU.State.Processors.Processor[0].State.CoreCnt), target))
			m = append(m, prometheus.MustNewConstMetric(cpuThreadCnt, prometheus.GaugeValue, stringToFloat(comp.CPU.State.Processors.Processor[0].State.ThreadCnt), target))
			m = append(m, prometheus.MustNewConstMetric(cpuFreq, prometheus.GaugeValue, stringToFloat(comp.CPU.State.Processors.Processor[0].State.Freq)*1000000, target))
			// F5OS reports CPU utilization as a 0-100 percentage; divide by
			// 100 to expose a true 0-1 ratio, per Prometheus naming
			// conventions.
			m = append(m, prometheus.MustNewConstMetric(cpuUtilization, prometheus.GaugeValue, float64(comp.CPU.State.CPUUtil.Current)/100, target))
			m = append(m, prometheus.MustNewConstMetric(cpuUtilization5SecAvg, prometheus.GaugeValue, float64(comp.CPU.State.CPUUtil.FiveSecAvg)/100, target))
			m = append(m, prometheus.MustNewConstMetric(cpuUtilization1MinAvg, prometheus.GaugeValue, float64(comp.CPU.State.CPUUtil.OneMinAvg)/100, target))
			m = append(m, prometheus.MustNewConstMetric(cpuUtilization5MinAvg, prometheus.GaugeValue, float64(comp.CPU.State.CPUUtil.FiveMinAvg)/100, target))
			// CacheSize is reported in KiB; multiply by 1024 to expose the
			// Prometheus base unit (bytes).
			m = append(m, prometheus.MustNewConstMetric(cpuCacheSize, prometheus.GaugeValue, stringToFloat(comp.CPU.State.Processors.Processor[0].State.CacheSize)*1024, target))
			for _, thread := range comp.CPU.State.CPUThreads.CPUThread {
				m = append(m, prometheus.MustNewConstMetric(cpuThreadUtilizationCurrent, prometheus.GaugeValue, float64(thread.Current)/100, target, strconv.Itoa(thread.ThreadIndex), thread.Thread))
				m = append(m, prometheus.MustNewConstMetric(cpuThreadUtilization5SecAvg, prometheus.GaugeValue, float64(thread.FiveSecAvg)/100, target, strconv.Itoa(thread.ThreadIndex), thread.Thread))
				m = append(m, prometheus.MustNewConstMetric(cpuThreadUtilization1MinAvg, prometheus.GaugeValue, float64(thread.OneMinAvg)/100, target, strconv.Itoa(thread.ThreadIndex), thread.Thread))
				m = append(m, prometheus.MustNewConstMetric(cpuThreadUtilization5MinAvg, prometheus.GaugeValue, float64(thread.FiveMinAvg)/100, target, strconv.Itoa(thread.ThreadIndex), thread.Thread))
			}
			for fanName, fanSpeedValue := range comp.FanTray.GetFans() {
				m = append(m, prometheus.MustNewConstMetric(fanSpeed, prometheus.GaugeValue, fanSpeedValue, target, fanName))
			}

		case "psu-1":
			m = append(m, prometheus.MustNewConstMetric(psuCurrentIn, prometheus.GaugeValue, stringToFloat(comp.PSUStats.CurrentIn), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuCurrentOut, prometheus.GaugeValue, stringToFloat(comp.PSUStats.CurrentOut), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuVoltageIn, prometheus.GaugeValue, stringToFloat(comp.PSUStats.VoltageIn), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuVoltageOut, prometheus.GaugeValue, stringToFloat(comp.PSUStats.VoltageOut), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuTemperature1, prometheus.GaugeValue, stringToFloat(comp.PSUStats.Temperature1), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuTemperature2, prometheus.GaugeValue, stringToFloat(comp.PSUStats.Temperature2), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuTemperature3, prometheus.GaugeValue, stringToFloat(comp.PSUStats.Temperature3), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuFan1Speed, prometheus.GaugeValue, float64(comp.PSUStats.Fan1Speed), target, comp.Name))
		case "psu-2":
			m = append(m, prometheus.MustNewConstMetric(psuCurrentIn, prometheus.GaugeValue, stringToFloat(comp.PSUStats.CurrentIn), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuCurrentOut, prometheus.GaugeValue, stringToFloat(comp.PSUStats.CurrentOut), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuVoltageIn, prometheus.GaugeValue, stringToFloat(comp.PSUStats.VoltageIn), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuVoltageOut, prometheus.GaugeValue, stringToFloat(comp.PSUStats.VoltageOut), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuTemperature1, prometheus.GaugeValue, stringToFloat(comp.PSUStats.Temperature1), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuTemperature2, prometheus.GaugeValue, stringToFloat(comp.PSUStats.Temperature2), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuTemperature3, prometheus.GaugeValue, stringToFloat(comp.PSUStats.Temperature3), target, comp.Name))
			m = append(m, prometheus.MustNewConstMetric(psuFan1Speed, prometheus.GaugeValue, float64(comp.PSUStats.Fan1Speed), target, comp.Name))
		}
	}

	return m, true
}

var numberRegex = regexp.MustCompile(`^\s*([+-]?\d+(?:\.\d+)?)`)

func stringToFloat(value string) float64 {
	if value == "" {
		return 0.0
	}
	matches := numberRegex.FindStringSubmatch(value)
	if len(matches) > 1 {
		f, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			log.Printf("Error converting string to float: %v", err)
			return 0.0
		}
		return float64(f)
	}
	log.Printf("Error converting string to float: No valid number found value is: %q", value)
	return 0.0
}
