package mpjolokia

import (
	mp "github.com/y-kuno/go-mackerel-plugin"

	"github.com/y-kuno/mackerel-plugin-jolokia/integration"
)

// integration list
const (
	Tomcat = "tomcat"
)

func (p *JolokiaPlugin) integrationGraphDef(graphdef map[string]mp.Graphs) map[string]mp.Graphs {
	switch p.Integration {
	case Tomcat:
		graphdef = integration.TomcatGraphDef(graphdef, p.LabelPrefix())
	}
	return graphdef
}

func (p *JolokiaPlugin) fetchIntegrationMetrics(stats map[string]float64) error {
	switch p.Integration {
	case Tomcat:
		if err := integration.FetchTomcatMetrics(stats, p.BaseURL); err != nil {
			return err
		}
	}
	return nil
}
