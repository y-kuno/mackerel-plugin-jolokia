package integration

import (
	"fmt"
	"strings"

	mp "github.com/y-kuno/go-mackerel-plugin"

	"github.com/y-kuno/mackerel-plugin-jolokia/http"
)

// Tomcat MBean List
const (
	MBeanGlobalRequestProcessor = "Catalina:name=*,type=GlobalRequestProcessor"
	MBeanThreadPool             = "Catalina:name=*,type=ThreadPool"
)

// Tomcat MBean List
const (
	AttributeThreadPool = "maxThreads,currentThreadCount,currentThreadsBusy"
)

// TomcatGraphDef is tomcat graph definitions
func TomcatGraphDef(graphdef map[string]mp.Graphs, prefix string) map[string]mp.Graphs {
	labelPrefix := strings.Title(prefix)
	graphdef["request.bytes.#"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s Request Bytes", labelPrefix)),
		Unit:  mp.UnitBytes,
		Metrics: []mp.Metrics{
			{Name: "bytesReceived", Label: "Received", Diff: true},
			{Name: "bytesSent", Label: "Sent", Diff: true},
		},
	}
	graphdef["request.count.#"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s Request Count", labelPrefix)),
		Unit:  mp.UnitInteger,
		Metrics: []mp.Metrics{
			{Name: "requestCount", Label: "Request", Diff: true},
			{Name: "errorCount", Label: "Error", Diff: true},
		},
	}
	graphdef["request.time.#"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s Request Time", labelPrefix)),
		Unit:  mp.UnitFloat,
		Metrics: []mp.Metrics{
			{Name: "maxTime", Label: "Max"},
			{Name: "processingTime", Label: "Processing", Diff: true},
		},
	}
	graphdef["threads.#"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s Threads", labelPrefix)),
		Unit:  mp.UnitInteger,
		Metrics: []mp.Metrics{
			{Name: "maxThreads", Label: "Max"},
			{Name: "currentThreadCount", Label: "Count"},
			{Name: "currentThreadsBusy", Label: "Busy"},
		},
	}
	return graphdef
}

// FetchTomcatMetrics is fetch tomcat metrics
func FetchTomcatMetrics(stats map[string]float64, url string) error {
	// fetch global request processor
	if err := fetchGlobalRequestProcessor(stats, url); err != nil {
		return err
	}
	// fetch thread pool
	if err := fetchThreadPool(stats, url); err != nil {
		return err
	}
	return nil
}

func fetchGlobalRequestProcessor(stats map[string]float64, url string) error {
	resp, err := http.DoRequest(fmt.Sprintf("%s/read/%s", url, MBeanGlobalRequestProcessor))
	if err != nil {
		return err
	}

	for k, v := range resp.Value.(map[string]interface{}) {
		value := v.(map[string]interface{})
		arr := strings.Split(strings.Split(k, "\"")[1], "-")
		p := fmt.Sprintf("%s_%s", arr[0], arr[2])
		// request byte
		stats[fmt.Sprintf("request.bytes.%s.bytesReceived", p)] = value["bytesReceived"].(float64)
		stats[fmt.Sprintf("request.bytes.%s.bytesSent", p)] = value["bytesSent"].(float64)
		// request count
		stats[fmt.Sprintf("request.count.%s.requestCount", p)] = value["requestCount"].(float64)
		stats[fmt.Sprintf("request.count.%s.errorCount", p)] = value["errorCount"].(float64)
		// processing time
		stats[fmt.Sprintf("request.time.%s.maxTime", p)] = value["maxTime"].(float64)
		stats[fmt.Sprintf("request.time.%s.processingTime", p)] = value["processingTime"].(float64)
	}
	return nil
}

func fetchThreadPool(stats map[string]float64, url string) error {
	resp, err := http.DoRequest(fmt.Sprintf("%s/read/%s/%s", url, MBeanThreadPool, AttributeThreadPool))
	if err != nil {
		return err
	}

	for k, v := range resp.Value.(map[string]interface{}) {
		value := v.(map[string]interface{})
		arr := strings.Split(strings.Split(k, "\"")[1], "-")
		p := fmt.Sprintf("%s_%s", arr[0], arr[2])
		for key, val := range value {
			stats[fmt.Sprintf("threads.%s.%s", p, key)] = val.(float64)
		}
	}
	return nil
}
