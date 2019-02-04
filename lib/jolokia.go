package mpjolokia

import (
	"flag"
	"fmt"
	"log"
	"os"

	mp "github.com/y-kuno/go-mackerel-plugin"

	"github.com/y-kuno/mackerel-plugin-jolokia/http"
)

// JolokiaPlugin mackerel plugin
type JolokiaPlugin struct {
	BaseURL     string
	Prefix      string
	ExcludeJVM  bool
	Integration string
	CustomFile  string
	JVM         struct {
		MBeanGC         []string
		MBeanMemoryPool []string
	}
	Custom Custom
}

// MetricKeyPrefix is metrics prefix
func (p *JolokiaPlugin) MetricKeyPrefix() string {
	if p.Prefix == "" {
		if p.Integration == "" {
			return "jolokia"
		}
		return p.Integration
	}
	return p.Prefix
}

// GraphDefinition interface for mackerel plugin
func (p *JolokiaPlugin) GraphDefinition() map[string]mp.Graphs {
	graphdef := make(map[string]mp.Graphs)
	if !p.ExcludeJVM {
		graphdef = p.jvmGraphDef(graphdef)
	}

	if p.Integration != "" {
		graphdef = p.integrationGraphDef(graphdef)
	}

	if p.CustomFile != "" {
		graphdef = p.customGraphDef(graphdef)
	}
	return graphdef
}

// LabelPrefix is graphs label prefix
func (p *JolokiaPlugin) LabelPrefix() string {
	if p.Prefix == "" {
		return p.Integration
	}
	return p.Prefix
}

// FetchMetrics interface for mackerel plugin
func (p *JolokiaPlugin) FetchMetrics() (map[string]float64, error) {
	stats := make(map[string]float64)
	if !p.ExcludeJVM {
		if err := p.fetchJVMMetrics(stats); err != nil {
			return nil, err
		}
	}

	if p.Integration != "" {
		if err := p.fetchIntegrationMetrics(stats); err != nil {
			return nil, err
		}
	}

	if p.CustomFile != "" {
		if err := p.fetchCustomMetrics(stats); err != nil {
			return nil, err
		}
	}
	return stats, nil
}

// FetchValue is fetch jolokia value
func FetchValue(stats map[string]float64, url, mbaen, attribute, prefix string) error {
	return FetchValueWithScope(stats, url, mbaen, attribute, prefix, []string{})
}

// FetchValueWithScope is fetch jolokia value with contains scope
func FetchValueWithScope(stats map[string]float64, url, mbaen, attribute, prefix string, scope []string) error {
	resp, err := http.DoRequest(fmt.Sprintf("%s/read/%s/%s", url, mbaen, attribute))
	if err != nil {
		return err
	}

	switch value := resp.Value.(type) {
	case int:
		if contains(scope, attribute) {
			stats[keyPrefix(prefix, attribute)] = float64(value)
		}
	case float64:
		if contains(scope, attribute) {
			stats[keyPrefix(prefix, attribute)] = value
		}
	case map[string]interface{}:
		for k, val := range value {
			if !contains(scope, k) {
				continue
			}

			switch v := val.(type) {
			case int:
				stats[keyPrefix(prefix, k)] = float64(v)
			case float64:
				stats[keyPrefix(prefix, k)] = v
			}
		}
	}
	return nil
}

func keyPrefix(prefix, key string) string {
	if prefix != "" {
		return fmt.Sprintf("%s.%s", prefix, key)
	}
	return key
}

func contains(path []string, key string) bool {
	if len(path) == 0 {
		return true
	}

	m := make(map[string]string)
	for _, s := range path {
		m[s] = s
	}

	_, ok := m[key]
	return ok
}

// Do the plugin
func Do() {
	optHost := flag.String("host", "localhost", "Hostname")
	optPort := flag.String("port", "8778", "Port")
	optPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	optExcludeJVM := flag.Bool("exclude-jvm-metrics", false, "Exclude JVM metrics")
	optIntegration := flag.String("integration", "", "Integration Name")
	optCustomFile := flag.String("custom-metrics-file", "", "Custom metrics file")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	p := &JolokiaPlugin{
		BaseURL:     fmt.Sprintf("http://%s:%s/jolokia", *optHost, *optPort),
		Prefix:      *optPrefix,
		ExcludeJVM:  *optExcludeJVM,
		Integration: *optIntegration,
		CustomFile:  *optCustomFile,
	}

	if !p.ExcludeJVM {
		// search JVM Mbean
		if err := p.searchMBean(); err != nil {
			log.Fatalf("serch JVM Mbean: %s", err)
			os.Exit(1)
		}
	}

	if p.CustomFile != "" {
		if err := p.readCustomFile(); err != nil {
			log.Fatalf("read custom mterics file: %s", err)
		}
	}

	plugin := mp.NewMackerelPlugin(p)
	plugin.Tempfile = *optTempfile
	plugin.Run()
}
