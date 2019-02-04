package mpjolokia

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	mp "github.com/y-kuno/go-mackerel-plugin"
)

// Custom is custom definition
type Custom struct {
	Graphs []struct {
		Key     string
		Label   string
		Unit    string
		Metrics []mp.Metrics
	}
	JMX []JMX
}

// JMX is jmx metrics definition
type JMX struct {
	MBean     string
	Attribute []struct {
		Name   string
		Prefix string
	}
	Scope []string
}

func (p *JolokiaPlugin) customGraphDef(graphdef map[string]mp.Graphs) map[string]mp.Graphs {
	for _, v := range p.Custom.Graphs {
		graphdef[v.Key] = mp.Graphs{
			Label:   v.Label,
			Unit:    v.Unit,
			Metrics: v.Metrics,
		}
	}
	return graphdef
}

func (p *JolokiaPlugin) fetchCustomMetrics(stats map[string]float64) error {
	jmx := p.Custom.JMX
	if len(jmx) == 0 {
		return fmt.Errorf("empty custom jmx metrics")
	}

	for _, v := range jmx {
		attr := v.Attribute
		if len(attr) == 0 {
			if err := FetchValueWithScope(stats, p.BaseURL, v.MBean, "", "", v.Scope); err != nil {
				return err
			}
			continue
		}

		for _, a := range attr {
			if err := FetchValueWithScope(stats, p.BaseURL, v.MBean, a.Name, a.Prefix, v.Scope); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *JolokiaPlugin) readCustomFile() error {
	buf, err := ioutil.ReadFile(p.CustomFile)
	if err != nil {
		return err
	}

	var c Custom
	if err := yaml.Unmarshal(buf, &c); err != nil {
		return err
	}
	p.Custom = c

	return nil
}
