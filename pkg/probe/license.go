package probe

import (
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/Haameed/f5_f5os_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

type LicensingResponse struct {
	Licensing Licensing `json:"f5-system-licensing:licensing"`
}

type Licensing struct {
	State LicensingState `json:"state"`
}

type LicensingState struct {
	RegistrationKey RegistrationKey `json:"registration-key"`
	License         string          `json:"license"`
	RawLicense      string          `json:"raw-license"`
}

type RegistrationKey struct {
	Base string `json:"base"`
}

func licenseField(license, label string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(label) + `\s+(.+?)\s*$`)
	matches := re.FindStringSubmatch(license)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func parseLicenseDate(value string) float64 {
	if value == "" {
		return 0
	}
	t, err := time.Parse("2006/01/02", value)
	if err != nil {
		log.Printf("Error parsing license date %q: %v", value, err)
		return 0
	}
	return float64(t.Unix())
}

func GetLicenseProbe(c http.BigIPHTTP, target string) ([]prometheus.Metric, bool) {
	var (
		licenseInfo = prometheus.NewDesc(
			"bigip_f5os_license_info",
			"BigIP F5OS license info",
			[]string{"target", "licensed_version", "platform_id"}, nil,
		)
		licensedDate = prometheus.NewDesc(
			"bigip_f5os_license_licensed_date_seconds",
			"BigIP F5OS licensed date as Unix timestamp",
			[]string{"target"}, nil,
		)
		serviceCheckDate = prometheus.NewDesc(
			"bigip_f5os_license_service_check_date_seconds",
			"BigIP F5OS service check date as Unix timestamp",
			[]string{"target"}, nil,
		)
	)

	var m []prometheus.Metric

	var resp LicensingResponse
	if err := c.Get("/api/data/openconfig-system:system/f5-system-licensing:licensing", &resp); err != nil {
		log.Printf("License stats unavailable for %s. error was %v", target, err)
	}

	lic := resp.Licensing.State.License

	licensedVersion := licenseField(lic, "Licensed version")
	platformID := licenseField(lic, "Platform ID")
	licensedDateStr := licenseField(lic, "Licensed date")
	serviceCheckDateStr := licenseField(lic, "Service check date")

	m = append(m, prometheus.MustNewConstMetric(licenseInfo, prometheus.GaugeValue, 0, target, licensedVersion, platformID))
	m = append(m, prometheus.MustNewConstMetric(licensedDate, prometheus.GaugeValue, parseLicenseDate(licensedDateStr), target))
	m = append(m, prometheus.MustNewConstMetric(serviceCheckDate, prometheus.GaugeValue, parseLicenseDate(serviceCheckDateStr), target))

	return m, true
}
