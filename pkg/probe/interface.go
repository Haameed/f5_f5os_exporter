package probe

import (
	"log"

	"github.com/Haameed/f5_f5os_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func GetInterfaceProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	type InterfaceCounters struct {
		InOctets         string `json:"in-octets"`
		InUnicastPkts    string `json:"in-unicast-pkts"`
		InBroadcastPkts  string `json:"in-broadcast-pkts"`
		InMulticastPkts  string `json:"in-multicast-pkts"`
		InDiscards       string `json:"in-discards"`
		InErrors         string `json:"in-errors"`
		InFCSErrors      string `json:"in-fcs-errors"`
		OutOctets        string `json:"out-octets"`
		OutUnicastPkts   string `json:"out-unicast-pkts"`
		OutBroadcastPkts string `json:"out-broadcast-pkts"`
		OutMulticastPkts string `json:"out-multicast-pkts"`
		OutDiscards      string `json:"out-discards"`
		OutErrors        string `json:"out-errors"`
	}
	type EthernetCounters struct {
		InMacControlFrames  string `json:"in-mac-control-frames"`
		InMacPauseFrames    string `json:"in-mac-pause-frames"`
		InOversizeFrames    string `json:"in-oversize-frames"`
		InJabberFrames      string `json:"in-jabber-frames"`
		InFragmentFrames    string `json:"in-fragment-frames"`
		In8021qFrames       string `json:"in-8021q-frames"`
		InCRCErrors         string `json:"in-crc-errors"`
		OutMacControlFrames string `json:"out-mac-control-frames"`
		OutMacPauseFrames   string `json:"out-mac-pause-frames"`
		Out8021qFrames      string `json:"out-8021q-frames"`
	}
	type FlowControl struct {
		Rx string `json:"rx,omitempty"`
		Tx string `json:"tx,omitempty"`
	}
	type EthernetConfig struct {
		AutoNegotiate bool   `json:"auto-negotiate,omitempty"`
		DuplexMode    string `json:"duplex-mode,omitempty"`
		PortSpeed     string `json:"port-speed,omitempty"`
		AggregateID   string `json:"openconfig-if-aggregate:aggregate-id,omitempty"`
	}

	type EthernetState struct {
		AutoNegotiate       bool              `json:"auto-negotiate,omitempty"`
		DuplexMode          string            `json:"duplex-mode,omitempty"`
		PortSpeed           string            `json:"port-speed,omitempty"`
		HwMacAddress        string            `json:"hw-mac-address,omitempty"`
		NegotiatedDuplex    string            `json:"negotiated-duplex-mode,omitempty"`
		NegotiatedPortSpeed string            `json:"negotiated-port-speed,omitempty"`
		Counters            *EthernetCounters `json:"counters,omitempty"`
		FlowControl         *FlowControl      `json:"f5-if-ethernet:flow-control,omitempty"`
	}
	type InterfaceConfig struct {
		Name    string `json:"name"`
		Type    string `json:"type,omitempty"`
		Enabled bool   `json:"enabled,omitempty"`
	}
	type Ethernet struct {
		Config EthernetConfig `json:"config,omitempty"`
		State  EthernetState  `json:"state,omitempty"`
	}
	type InterfaceState struct {
		Name       string             `json:"name"`
		Type       string             `json:"type,omitempty"`
		MTU        int                `json:"mtu,omitempty"`
		Enabled    bool               `json:"enabled,omitempty"`
		OperStatus string             `json:"oper-status,omitempty"`
		Counters   *InterfaceCounters `json:"counters,omitempty"`
		FEC        string             `json:"f5-interface:forward-error-correction,omitempty"`
		LacpState  string             `json:"f5-lacp:lacp_state,omitempty"`
	}

	type Interface struct {
		Name     string          `json:"name"`
		Config   InterfaceConfig `json:"config,omitempty"`
		State    InterfaceState  `json:"state,omitempty"`
		Ethernet *Ethernet       `json:"openconfig-if-ethernet:ethernet,omitempty"`
	}
	type OpenconfigInterfaces struct {
		Interface []Interface `json:"interface"`
	}
	type InterfacesResponse struct {
		OpenconfigInterfaces `json:"openconfig-interfaces:interfaces"`
	}

	var (
		interfaceInfo = prometheus.NewDesc(
			"bigip_f5os_interface_info",
			"BigIP F5OS interface info",
			[]string{"target", "interface", "type", "oper_status", "fec", "lacp_state"}, nil,
		)
		interfaceEnabled = prometheus.NewDesc(
			"bigip_f5os_interface_enabled",
			"BigIP F5OS interface admin enabled (1 enabled, 0 disabled)",
			[]string{"target", "interface"}, nil,
		)
		interfaceOperStatus = prometheus.NewDesc(
			"bigip_f5os_interface_oper_status_up",
			"BigIP F5OS interface operational status (1 UP, 0 otherwise)",
			[]string{"target", "interface"}, nil,
		)
		interfaceMTU = prometheus.NewDesc(
			"bigip_f5os_interface_mtu_bytes",
			"BigIP F5OS interface MTU in bytes",
			[]string{"target", "interface"}, nil,
		)
		inOctets = prometheus.NewDesc(
			"bigip_f5os_interface_in_octets_total",
			"BigIP F5OS interface received octets",
			[]string{"target", "interface"}, nil,
		)
		inUnicastPkts = prometheus.NewDesc(
			"bigip_f5os_interface_in_unicast_packets_total",
			"BigIP F5OS interface received unicast packets",
			[]string{"target", "interface"}, nil,
		)
		inBroadcastPkts = prometheus.NewDesc(
			"bigip_f5os_interface_in_broadcast_packets_total",
			"BigIP F5OS interface received broadcast packets",
			[]string{"target", "interface"}, nil,
		)
		inMulticastPkts = prometheus.NewDesc(
			"bigip_f5os_interface_in_multicast_packets_total",
			"BigIP F5OS interface received multicast packets",
			[]string{"target", "interface"}, nil,
		)
		inDiscards = prometheus.NewDesc(
			"bigip_f5os_interface_in_discards_total",
			"BigIP F5OS interface received discards",
			[]string{"target", "interface"}, nil,
		)
		inErrors = prometheus.NewDesc(
			"bigip_f5os_interface_in_errors_total",
			"BigIP F5OS interface received errors",
			[]string{"target", "interface"}, nil,
		)
		inFCSErrors = prometheus.NewDesc(
			"bigip_f5os_interface_in_fcs_errors_total",
			"BigIP F5OS interface received FCS errors",
			[]string{"target", "interface"}, nil,
		)
		outOctets = prometheus.NewDesc(
			"bigip_f5os_interface_out_octets_total",
			"BigIP F5OS interface transmitted octets",
			[]string{"target", "interface"}, nil,
		)
		outUnicastPkts = prometheus.NewDesc(
			"bigip_f5os_interface_out_unicast_packets_total",
			"BigIP F5OS interface transmitted unicast packets",
			[]string{"target", "interface"}, nil,
		)
		outBroadcastPkts = prometheus.NewDesc(
			"bigip_f5os_interface_out_broadcast_packets_total",
			"BigIP F5OS interface transmitted broadcast packets",
			[]string{"target", "interface"}, nil,
		)
		outMulticastPkts = prometheus.NewDesc(
			"bigip_f5os_interface_out_multicast_packets_total",
			"BigIP F5OS interface transmitted multicast packets",
			[]string{"target", "interface"}, nil,
		)
		outDiscards = prometheus.NewDesc(
			"bigip_f5os_interface_out_discards_total",
			"BigIP F5OS interface transmitted discards",
			[]string{"target", "interface"}, nil,
		)
		outErrors = prometheus.NewDesc(
			"bigip_f5os_interface_out_errors_total",
			"BigIP F5OS interface transmitted errors",
			[]string{"target", "interface"}, nil,
		)

		ethernetInfo = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_info",
			"BigIP F5OS ethernet interface info",
			[]string{"target", "interface", "port_speed", "negotiated_port_speed", "duplex_mode", "negotiated_duplex_mode", "hw_mac_address", "aggregate_id"}, nil,
		)
		ethernetAutoNegotiate = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_auto_negotiate",
			"BigIP F5OS ethernet interface auto-negotiate (1 true, 0 false)",
			[]string{"target", "interface"}, nil,
		)
		ethernetFlowControlRx = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_flow_control_rx",
			"BigIP F5OS ethernet interface RX flow control (1 on, 0 off)",
			[]string{"target", "interface"}, nil,
		)
		ethernetFlowControlTx = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_flow_control_tx",
			"BigIP F5OS ethernet interface TX flow control (1 on, 0 off)",
			[]string{"target", "interface"}, nil,
		)

		ethInMacControlFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_in_mac_control_frames_total",
			"BigIP F5OS ethernet received MAC control frames",
			[]string{"target", "interface"}, nil,
		)
		ethInMacPauseFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_in_mac_pause_frames_total",
			"BigIP F5OS ethernet received MAC pause frames",
			[]string{"target", "interface"}, nil,
		)
		ethInOversizeFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_in_oversize_frames_total",
			"BigIP F5OS ethernet received oversize frames",
			[]string{"target", "interface"}, nil,
		)
		ethInJabberFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_in_jabber_frames_total",
			"BigIP F5OS ethernet received jabber frames",
			[]string{"target", "interface"}, nil,
		)
		ethInFragmentFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_in_fragment_frames_total",
			"BigIP F5OS ethernet received fragment frames",
			[]string{"target", "interface"}, nil,
		)
		ethIn8021qFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_in_8021q_frames_total",
			"BigIP F5OS ethernet received 802.1q frames",
			[]string{"target", "interface"}, nil,
		)
		ethInCRCErrors = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_in_crc_errors_total",
			"BigIP F5OS ethernet received CRC errors",
			[]string{"target", "interface"}, nil,
		)
		ethOutMacControlFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_out_mac_control_frames_total",
			"BigIP F5OS ethernet transmitted MAC control frames",
			[]string{"target", "interface"}, nil,
		)
		ethOutMacPauseFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_out_mac_pause_frames_total",
			"BigIP F5OS ethernet transmitted MAC pause frames",
			[]string{"target", "interface"}, nil,
		)
		ethOut8021qFrames = prometheus.NewDesc(
			"bigip_f5os_interface_ethernet_out_8021q_frames_total",
			"BigIP F5OS ethernet transmitted 802.1q frames",
			[]string{"target", "interface"}, nil,
		)
	)

	var m []prometheus.Metric

	var resp InterfacesResponse
	if err := c.Get("/api/data/openconfig-interfaces:interfaces", &resp); err != nil {
		log.Printf("Interface stats unavailable for %s. error was %v", target, err)
	}

	for _, iface := range resp.Interface {
		name := iface.Name
		st := iface.State

		m = append(m, prometheus.MustNewConstMetric(interfaceInfo, prometheus.GaugeValue, 0, target, name, st.Type, st.OperStatus, st.FEC, st.LacpState))
		m = append(m, prometheus.MustNewConstMetric(interfaceEnabled, prometheus.GaugeValue, boolToFloat(st.Enabled), target, name))
		m = append(m, prometheus.MustNewConstMetric(interfaceOperStatus, prometheus.GaugeValue, boolToFloat(st.OperStatus == "UP"), target, name))
		m = append(m, prometheus.MustNewConstMetric(interfaceMTU, prometheus.GaugeValue, float64(st.MTU), target, name))

		if st.Counters != nil {
			ct := st.Counters
			m = append(m, prometheus.MustNewConstMetric(inOctets, prometheus.CounterValue, stringToFloat(ct.InOctets), target, name))
			m = append(m, prometheus.MustNewConstMetric(inUnicastPkts, prometheus.CounterValue, stringToFloat(ct.InUnicastPkts), target, name))
			m = append(m, prometheus.MustNewConstMetric(inBroadcastPkts, prometheus.CounterValue, stringToFloat(ct.InBroadcastPkts), target, name))
			m = append(m, prometheus.MustNewConstMetric(inMulticastPkts, prometheus.CounterValue, stringToFloat(ct.InMulticastPkts), target, name))
			m = append(m, prometheus.MustNewConstMetric(inDiscards, prometheus.CounterValue, stringToFloat(ct.InDiscards), target, name))
			m = append(m, prometheus.MustNewConstMetric(inErrors, prometheus.CounterValue, stringToFloat(ct.InErrors), target, name))
			m = append(m, prometheus.MustNewConstMetric(inFCSErrors, prometheus.CounterValue, stringToFloat(ct.InFCSErrors), target, name))
			m = append(m, prometheus.MustNewConstMetric(outOctets, prometheus.CounterValue, stringToFloat(ct.OutOctets), target, name))
			m = append(m, prometheus.MustNewConstMetric(outUnicastPkts, prometheus.CounterValue, stringToFloat(ct.OutUnicastPkts), target, name))
			m = append(m, prometheus.MustNewConstMetric(outBroadcastPkts, prometheus.CounterValue, stringToFloat(ct.OutBroadcastPkts), target, name))
			m = append(m, prometheus.MustNewConstMetric(outMulticastPkts, prometheus.CounterValue, stringToFloat(ct.OutMulticastPkts), target, name))
			m = append(m, prometheus.MustNewConstMetric(outDiscards, prometheus.CounterValue, stringToFloat(ct.OutDiscards), target, name))
			m = append(m, prometheus.MustNewConstMetric(outErrors, prometheus.CounterValue, stringToFloat(ct.OutErrors), target, name))
		}

		if iface.Ethernet != nil {
			es := iface.Ethernet.State
			m = append(m, prometheus.MustNewConstMetric(ethernetInfo, prometheus.GaugeValue, 0, target, name, es.PortSpeed, es.NegotiatedPortSpeed, es.DuplexMode, es.NegotiatedDuplex, es.HwMacAddress, iface.Ethernet.Config.AggregateID))
			m = append(m, prometheus.MustNewConstMetric(ethernetAutoNegotiate, prometheus.GaugeValue, boolToFloat(es.AutoNegotiate), target, name))

			if es.FlowControl != nil {
				m = append(m, prometheus.MustNewConstMetric(ethernetFlowControlRx, prometheus.GaugeValue, boolToFloat(es.FlowControl.Rx == "on"), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethernetFlowControlTx, prometheus.GaugeValue, boolToFloat(es.FlowControl.Tx == "on"), target, name))
			}

			if es.Counters != nil {
				ec := es.Counters
				m = append(m, prometheus.MustNewConstMetric(ethInMacControlFrames, prometheus.CounterValue, stringToFloat(ec.InMacControlFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethInMacPauseFrames, prometheus.CounterValue, stringToFloat(ec.InMacPauseFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethInOversizeFrames, prometheus.CounterValue, stringToFloat(ec.InOversizeFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethInJabberFrames, prometheus.CounterValue, stringToFloat(ec.InJabberFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethInFragmentFrames, prometheus.CounterValue, stringToFloat(ec.InFragmentFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethIn8021qFrames, prometheus.CounterValue, stringToFloat(ec.In8021qFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethInCRCErrors, prometheus.CounterValue, stringToFloat(ec.InCRCErrors), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethOutMacControlFrames, prometheus.CounterValue, stringToFloat(ec.OutMacControlFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethOutMacPauseFrames, prometheus.CounterValue, stringToFloat(ec.OutMacPauseFrames), target, name))
				m = append(m, prometheus.MustNewConstMetric(ethOut8021qFrames, prometheus.CounterValue, stringToFloat(ec.Out8021qFrames), target, name))
			}
		}
	}

	return m, true
}
