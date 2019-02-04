package mpjolokia

import (
	"fmt"
	"regexp"
	"strings"

	mp "github.com/y-kuno/go-mackerel-plugin"

	"github.com/y-kuno/mackerel-plugin-jolokia/http"
)

var reg = regexp.MustCompile(`java.lang.name=(.+),type=(.+)`)

// JVM MBean List
const (
	MBeanClassLoading    = "java.lang:type=ClassLoading"
	MBeanMemory          = "java.lang:type=Memory"
	MBeanOperatingSystem = "java.lang:type=OperatingSystem"
	MBeanThreading       = "java.lang:type=Threading"
)

// JVM Attribute List
const (
	AttributeClassLoading     = "LoadedClassCount,UnloadedClassCount,TotalLoadedClassCount"
	AttributeGarbageCollector = "CollectionCount,CollectionTime"
	AttributeMemory           = "HeapMemoryUsage,NonHeapMemoryUsage"
	AttributeMemoryPool       = "Usage"
	AttributeOperatingSystem  = "ProcessCpuLoad,SystemCpuLoad"
	AttributeThreading        = "DaemonThreadCount,PeakThreadCount,ThreadCount"
)

func (p *JolokiaPlugin) jvmGraphDef(graphdef map[string]mp.Graphs) map[string]mp.Graphs {
	labelPrefix := strings.Title(p.LabelPrefix())
	graphdef["jvm.memory.heap"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s JVM Heap Memory", labelPrefix)),
		Unit:  mp.UnitBytes,
		Metrics: []mp.Metrics{
			{Name: "init", Match: true},
			{Name: "committed", Match: true},
			{Name: "max", Match: true},
			{Name: "used", Match: true},
		},
	}
	graphdef["jvm.memory.non_heap"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s JVM Non-Heap Memory", labelPrefix)),
		Unit:  mp.UnitBytes,
		Metrics: []mp.Metrics{
			{Name: "init", Match: true},
			{Name: "committed", Match: true},
			{Name: "max", Match: true},
			{Name: "used", Match: true},
		},
	}
	graphdef["jvm.class_load"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s JVM Class Loaders", labelPrefix)),
		Unit:  mp.UnitInteger,
		Metrics: []mp.Metrics{
			{Name: "LoadedClassCount", Label: "Loaded"},
			{Name: "UnloadedClassCount", Label: "Unloaded"},
			{Name: "TotalLoadedClassCount", Label: "Total"},
		},
	}
	graphdef["jvm.threads"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s JVM Threads", labelPrefix)),
		Unit:  mp.UnitInteger,
		Metrics: []mp.Metrics{
			{Name: "DaemonThreadCount", Label: "Daemon"},
			{Name: "PeakThreadCount", Label: "Peak"},
			{Name: "ThreadCount", Label: "Count"},
		},
	}
	graphdef["jvm.ops.cpu_load"] = mp.Graphs{
		Label: strings.TrimSpace(fmt.Sprintf("%s JVM CPU Load", labelPrefix)),
		Unit:  mp.UnitPercentage,
		Metrics: []mp.Metrics{
			{Name: "ProcessCpuLoad", Label: "Process", Scale: 100},
			{Name: "SystemCpuLoad", Label: "System", Scale: 100},
		},
	}

	// gc graphs
	var gcCountsMetrics []mp.Metrics
	var gcTimeMetrics []mp.Metrics
	var gcTimePercentageMetrics []mp.Metrics
	for _, v := range p.JVM.MBeanGC {
		beanName := reg.FindStringSubmatch(v)[1]
		metricName := strings.Replace(beanName, " ", "", -1)

		gcCountsMetrics = append(gcCountsMetrics,
			mp.Metrics{
				Name:  metricName,
				Label: beanName,
				Diff:  true,
				Match: true})
		gcTimeMetrics = append(gcTimeMetrics,
			mp.Metrics{
				Name:  metricName,
				Label: beanName,
				Diff:  true,
				Match: true})
		gcTimePercentageMetrics = append(gcTimePercentageMetrics,
			mp.Metrics{
				Name:  metricName,
				Label: beanName,
				Diff:  true,
				Scale: 100.0 / 60,
				Match: true})
	}
	graphdef["jvm.gc.counts"] = mp.Graphs{
		Label:   strings.TrimSpace(fmt.Sprintf("%s JVM GC Counts", labelPrefix)),
		Unit:    mp.UnitInteger,
		Metrics: gcCountsMetrics,
	}
	graphdef["jvm.gc.time"] = mp.Graphs{
		Label:   strings.TrimSpace(fmt.Sprintf("%s JVM GC Time", labelPrefix)),
		Unit:    mp.UnitFloat,
		Metrics: gcTimeMetrics,
	}
	// gc.time_percentage is the percentage of gc time to 60 sec
	graphdef["jvm.gc.time_percentage"] = mp.Graphs{
		Label:   strings.TrimSpace(fmt.Sprintf("%s JVM GC Time Percentage", labelPrefix)),
		Unit:    mp.UnitPercentage,
		Metrics: gcTimePercentageMetrics,
	}

	// memory-pool graphs
	for _, v := range p.JVM.MBeanMemoryPool {
		beanName := reg.FindStringSubmatch(v)[1]
		graphdef[fmt.Sprintf("jvm.memory.%s",
			strings.ToLower(strings.ToLower(strings.Replace(beanName, " ", "_", -1))))] = mp.Graphs{
			Label: strings.TrimSpace(fmt.Sprintf("%s JVM %s", labelPrefix, beanName)),
			Unit:  mp.UnitBytes,
			Metrics: []mp.Metrics{
				{Name: "init", Match: true},
				{Name: "committed", Match: true},
				{Name: "max", Match: true},
				{Name: "used", Match: true},
			},
		}
	}
	return graphdef
}

func (p *JolokiaPlugin) fetchJVMMetrics(stats map[string]float64) error {
	// fetch class loading
	if err := FetchValue(stats, p.BaseURL, MBeanClassLoading, AttributeClassLoading, ""); err != nil {
		return err
	}
	// fetch operating system
	if err := FetchValue(stats, p.BaseURL, MBeanOperatingSystem, AttributeOperatingSystem, ""); err != nil {
		return err
	}
	// fetch threading
	if err := FetchValue(stats, p.BaseURL, MBeanThreading, AttributeThreading, ""); err != nil {
		return err
	}

	// fetch garbage collector
	if err := p.fetchGarbageCollector(stats); err != nil {
		return err
	}
	// fetch memory
	if err := p.fetchMemory(stats); err != nil {
		return err
	}
	// fetch memory pool
	if err := p.fetchMemoryPool(stats); err != nil {
		return err
	}

	return nil
}

func (p *JolokiaPlugin) fetchGarbageCollector(stats map[string]float64) error {
	for _, v := range p.JVM.MBeanGC {
		beanName := reg.FindStringSubmatch(v)[1]
		metricName := strings.Replace(beanName, " ", "", -1)

		url := fmt.Sprintf("%s/read/%s/%s", p.BaseURL, strings.Replace(v, " ", "%20", -1), AttributeGarbageCollector)
		resp, err := http.DoRequest(url)
		if err != nil {
			return err
		}

		value := resp.Value.(map[string]interface{})
		stats[fmt.Sprintf("jvm.gc.counts.%s", metricName)] = value["CollectionCount"].(float64)
		// CollectionTime is approximate accumulated collection elapsed time in milliseconds.
		gcTime := value["CollectionTime"].(float64) / 1000.0
		stats[fmt.Sprintf("jvm.gc.time.%s", metricName)] = gcTime
		stats[fmt.Sprintf("jvm.gc.time_percentage.%s", metricName)] = gcTime
	}
	return nil
}

func (p *JolokiaPlugin) fetchMemory(stats map[string]float64) error {
	resp, err := http.DoRequest(fmt.Sprintf("%s/read/%s/%s", p.BaseURL, MBeanMemory, AttributeMemory))
	if err != nil {
		return err
	}

	value := resp.Value.(map[string]interface{})
	// heap memory
	heap := value["HeapMemoryUsage"].(map[string]interface{})
	for k, v := range heap {
		f := v.(float64)
		if f > 0 {
			stats[fmt.Sprintf("jvm.memory.heap.%s", k)] = f
		}
	}
	// non-heap memory
	nonHeap := value["NonHeapMemoryUsage"].(map[string]interface{})
	for k, v := range nonHeap {
		f := v.(float64)
		if f > 0 {
			stats[fmt.Sprintf("jvm.memory.non_heap.%s", k)] = f
		}
	}
	return nil
}

func (p *JolokiaPlugin) fetchMemoryPool(stats map[string]float64) error {
	for _, v := range p.JVM.MBeanMemoryPool {
		url := fmt.Sprintf("%s/read/%s/%s", p.BaseURL, strings.Replace(v, " ", "%20", -1), AttributeMemoryPool)
		resp, err := http.DoRequest(url)
		if err != nil {
			return err
		}

		beanName := reg.FindStringSubmatch(v)[1]
		key := fmt.Sprintf("jvm.memory.%s", strings.ToLower(strings.ToLower(strings.Replace(beanName, " ", "_", -1))))
		for k, v := range resp.Value.(map[string]interface{}) {
			f := v.(float64)
			if f > 0 {
				stats[fmt.Sprintf("%s.%s", key, k)] = f
			}
		}
	}
	return nil
}

func (p *JolokiaPlugin) searchMBean() error {
	url := fmt.Sprintf("%s/search/java.lang:*", p.BaseURL)
	resp, err := http.DoRequest(url)
	if err != nil {
		return err
	}

	/*
		"java.lang:name=Metaspace,type=MemoryPool",
		"java.lang:name=PS Old Gen,type=MemoryPool",
		"java.lang:name=PS Scavenge,type=GarbageCollector",
		"java.lang:name=PS Eden Space,type=MemoryPool",
		"java.lang:type=Runtime",
		"java.lang:type=Threading",
		"java.lang:type=OperatingSystem",
		"java.lang:name=Code Cache,type=MemoryPool",
		"java.lang:type=Compilation",
		"java.lang:name=CodeCacheManager,type=MemoryManager",
		"java.lang:name=Compressed Class Space,type=MemoryPool",
		"java.lang:type=Memory",
		"java.lang:name=PS Survivor Space,type=MemoryPool",
		"java.lang:type=ClassLoading",
		"java.lang:name=Metaspace Manager,type=MemoryManager",
		"java.lang:name=PS MarkSweep,type=GarbageCollector"
	*/

	var gc []string
	var memory []string
	for _, v := range resp.Value.([]interface{}) {
		str := v.(string)
		mbean := reg.FindStringSubmatch(str)
		if len(mbean) == 3 {
			switch mbean[2] {
			case "GarbageCollector":
				gc = append(gc, str)
			case "MemoryPool":
				memory = append(memory, str)
			}
		}
	}
	p.JVM.MBeanGC = gc
	p.JVM.MBeanMemoryPool = memory
	return nil
}
