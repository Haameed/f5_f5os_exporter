package probe

import (
	"log"

	"github.com/Haameed/f5_f5os_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetLACPProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {

	type Config struct {
		Name     string `json:"name"`
		Interval string `json:"interval"`
		LACPMode string `json:"lacp-mode"`
	}

	type State struct {
		Name        string `json:"name"`
		Interval    string `json:"interval"`
		LACPMode    string `json:"lacp-mode"`
		SystemIDMAC string `json:"system-id-mac"`
	}

	type InterfaceStateCounter struct {
		LacpInPkts   string `json:"lacp-in-pkts"`
		LacpOutPkts  string `json:"lacp-out-pkts"`
		LacpRxErrors string `json:"lacp-rx-errors"`
	}
	type InterfaceState struct {
		Interface       string                 `json:"interface"`
		Activity        string                 `json:"activity"`
		Timeout         string                 `json:"timeout"`
		Synchronization string                 `json:"synchronization"`
		Aggregatable    bool                   `json:"aggregatable"`
		Collecting      bool                   `json:"collecting"`
		Distributing    bool                   `json:"distributing"`
		SystemID        string                 `json:"system-id"`
		OperKey         int                    `json:"oper-key"`
		PartnerID       string                 `json:"partner-id"`
		PartnerKey      int                    `json:"partner-key"`
		PortNum         int                    `json:"port-num"`
		PartnerPortNum  int                    `json:"partner-port-num"`
		Counters        *InterfaceStateCounter `json:"counters"`
	}

	type MemberInfo struct {
		Interface string         `json:"interface"`
		State     InterfaceState `json:"state"`
	}
	type Members struct {
		Member []MemberInfo `json:"member"`
	}

	type LACPInterface struct {
		Name    string  `json:"name"`
		Config  Config  `json:"config"`
		State   State   `json:"state"`
		Membres Members `json:"members"`
	}

	type OpenConfigLacpResponse struct {
		Interfaces []LACPInterface `json:"openconfig-lacp:interface"`
	}

	var (
		lacpInterfaceInfo = prometheus.NewDesc(
			"bigip_f5os_lacp_interface_info",
			"BigIP F5OS LACP interface info",
			[]string{"target", "lag_name", "interval", "lacp_mode", "system_id_mac"}, nil,
		)
		lacpMemberInfo = prometheus.NewDesc(
			"bigip_f5os_lacp_member_info",
			"BigIP F5OS LACP member info",
			[]string{"target", "lag_name", "interface", "activity", "timeout", "synchronization", "system_id", "partner_id"}, nil,
		)
		lacpMemberAggregatable = prometheus.NewDesc(
			"bigip_f5os_lacp_member_aggregatable",
			"BigIP F5OS LACP member aggregatable (1 true, 0 false)",
			[]string{"target", "lag_name", "interface"}, nil,
		)
		lacpMemberCollecting = prometheus.NewDesc(
			"bigip_f5os_lacp_member_collecting",
			"BigIP F5OS LACP member collecting (1 true, 0 false)",
			[]string{"target", "lag_name", "interface"}, nil,
		)
		lacpMemberDistributing = prometheus.NewDesc(
			"bigip_f5os_lacp_member_distributing",
			"BigIP F5OS LACP member distributing (1 true, 0 false)",
			[]string{"target", "lag_name", "interface"}, nil,
		)
		lacpMemberSynced = prometheus.NewDesc(
			"bigip_f5os_lacp_member_in_sync",
			"BigIP F5OS LACP member synchronization state (1 IN_SYNC, 0 otherwise)",
			[]string{"target", "lag_name", "interface"}, nil,
		)
		// lacpMemberOperKey = prometheus.NewDesc(
		// 	"bigip_f5os_lacp_member_oper_key",
		// 	"BigIP F5OS LACP member operational key",
		// 	[]string{"target", "lag_name", "interface"}, nil,
		// )
		// lacpMemberPartnerKey = prometheus.NewDesc(
		// 	"bigip_f5os_lacp_member_partner_key",
		// 	"BigIP F5OS LACP member partner key",
		// 	[]string{"target", "lag_name", "interface"}, nil,
		// )
		// lacpMemberPortNum = prometheus.NewDesc(
		// 	"bigip_f5os_lacp_member_port_num",
		// 	"BigIP F5OS LACP member port number",
		// 	[]string{"target", "lag_name", "interface"}, nil,
		// )
		// lacpMemberPartnerPortNum = prometheus.NewDesc(
		// 	"bigip_f5os_lacp_member_partner_port_num",
		// 	"BigIP F5OS LACP member partner port number",
		// 	[]string{"target", "lag_name", "interface"}, nil,
		// )
		lacpInPkts = prometheus.NewDesc(
			"bigip_f5os_lacp_in_packets_total",
			"BigIP F5OS LACP received packets",
			[]string{"target", "lag_name", "interface"}, nil,
		)
		lacpOutPkts = prometheus.NewDesc(
			"bigip_f5os_lacp_out_packets_total",
			"BigIP F5OS LACP transmitted packets",
			[]string{"target", "lag_name", "interface"}, nil,
		)
		lacpRxErrors = prometheus.NewDesc(
			"bigip_f5os_lacp_rx_errors_total",
			"BigIP F5OS LACP receive errors",
			[]string{"target", "lag_name", "interface"}, nil,
		)
	)

	var m []prometheus.Metric

	var resp OpenConfigLacpResponse
	if err := c.Get("/api/data/openconfig-lacp:lacp/interfaces/interface", &resp); err != nil {
		log.Printf("LACP stats unavailable for %s. error was %v", target, err)
	}

	for _, iface := range resp.Interfaces {
		m = append(m, prometheus.MustNewConstMetric(lacpInterfaceInfo, prometheus.GaugeValue, 0, target, iface.Name, iface.State.Interval, iface.State.LACPMode, iface.State.SystemIDMAC))
		for _, member := range iface.Membres.Member {
			s := member.State
			m = append(m, prometheus.MustNewConstMetric(lacpMemberInfo, prometheus.GaugeValue, 0, target, iface.Name, s.Interface, s.Activity, s.Timeout, s.Synchronization, s.SystemID, s.PartnerID))
			m = append(m, prometheus.MustNewConstMetric(lacpMemberAggregatable, prometheus.GaugeValue, boolToFloat(s.Aggregatable), target, iface.Name, s.Interface))
			m = append(m, prometheus.MustNewConstMetric(lacpMemberCollecting, prometheus.GaugeValue, boolToFloat(s.Collecting), target, iface.Name, s.Interface))
			m = append(m, prometheus.MustNewConstMetric(lacpMemberDistributing, prometheus.GaugeValue, boolToFloat(s.Distributing), target, iface.Name, s.Interface))
			m = append(m, prometheus.MustNewConstMetric(lacpMemberSynced, prometheus.GaugeValue, boolToFloat(s.Synchronization == "IN_SYNC"), target, iface.Name, s.Interface))
			// m = append(m, prometheus.MustNewConstMetric(lacpMemberOperKey, prometheus.GaugeValue, float64(s.OperKey), target, iface.Name, s.Interface))
			// m = append(m, prometheus.MustNewConstMetric(lacpMemberPartnerKey, prometheus.GaugeValue, float64(s.PartnerKey), target, iface.Name, s.Interface))
			// m = append(m, prometheus.MustNewConstMetric(lacpMemberPortNum, prometheus.GaugeValue, float64(s.PortNum), target, iface.Name, s.Interface))
			// m = append(m, prometheus.MustNewConstMetric(lacpMemberPartnerPortNum, prometheus.GaugeValue, float64(s.PartnerPortNum), target, iface.Name, s.Interface))

			if s.Counters != nil {
				m = append(m, prometheus.MustNewConstMetric(lacpInPkts, prometheus.CounterValue, stringToFloat(s.Counters.LacpInPkts), target, iface.Name, s.Interface))
				m = append(m, prometheus.MustNewConstMetric(lacpOutPkts, prometheus.CounterValue, stringToFloat(s.Counters.LacpOutPkts), target, iface.Name, s.Interface))
				m = append(m, prometheus.MustNewConstMetric(lacpRxErrors, prometheus.CounterValue, stringToFloat(s.Counters.LacpRxErrors), target, iface.Name, s.Interface))
			}
		}

	}

	return m, true
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
