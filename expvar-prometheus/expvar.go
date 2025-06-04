// Package expvar_prometheus provides a Prometheus collector for Go expvar. It exposes all variables
// it can automatically.
package expvar_prometheus

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bradfitz/iter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	prometheus.MustRegister(NewExpvarCollector())
	http.Handle("/debug/prometheus_default", promhttp.Handler())
}

// A Prometheus collector that exposes all Go expvars.
type Collector struct {
	descs map[int]*prometheus.Desc
}

func NewCollector() Collector {
	// This could probably be a global instance.
	return Collector{
		descs: make(map[int]*prometheus.Desc),
	}
}

const (
	fqName = "go_expvar"
	help   = "All expvars"
)

var desc = prometheus.NewDesc(fqName, help, nil, nil)

// Describe implements Collector.
func (e Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- desc
}

// Collect implements Collector.
func (e Collector) Collect(ch chan<- prometheus.Metric) {
	expvar.Do(func(kv expvar.KeyValue) {
		// I think this is very noisy, and there seems to be good support for exporting its
		// information in a more structured way.
		if kv.Key == "memstats" {
			return
		}
		expvarVisitor{
			f: func(m prometheus.Metric) {
				ch <- m
			},
			labelValues: []string{kv.Key},
			descs:       e.descs,
		}.collectVar(kv.Value)
	})
}

func labels(n int) (ls []string) {
	for i := range iter.N(n) {
		ls = append(ls, "key"+strconv.FormatInt(int64(i), 10))
	}
	return
}

// Walks expvar.Vars emitting metrics into f. Extends labelValues and creates new value instances as
// it descends.
type expvarVisitor struct {
	f           func(prometheus.Metric)
	labelValues []string
	descs       map[int]*prometheus.Desc
}

func (c expvarVisitor) newMetric(f float64) {
	c.f(prometheus.MustNewConstMetric(
		c.desc(),
		prometheus.UntypedValue,
		f,
		c.labelValues...))
}

func (c expvarVisitor) desc() *prometheus.Desc {
	d, ok := c.descs[len(c.labelValues)]
	if !ok {
		d = prometheus.NewDesc(fqName, "", labels(len(c.labelValues)), nil)
		c.descs[len(c.labelValues)] = d
	}
	return d
}

func (c expvarVisitor) metricError(err error) {
	c.f(prometheus.NewInvalidMetric(c.desc(), err))
}

func (c expvarVisitor) withLabelValue(lv string) expvarVisitor {
	//if !utf8.ValidString(lv) {
	//	lv = strconv.Quote(lv)
	//}
	c.labelValues = append(c.labelValues, lv)
	return c
}

func (c expvarVisitor) collectJsonValue(v interface{}) {
	switch v := v.(type) {
	case float64:
		c.newMetric(v)
	case map[string]interface{}:
		for k, v := range v {
			c.withLabelValue(k).collectJsonValue(v)
		}
	case bool:
		if v {
			c.newMetric(1)
		} else {
			c.newMetric(0)
		}
	case string:
		c.f(prometheus.MustNewConstMetric(
			prometheus.NewDesc("go_expvar", "",
				append(labels(len(c.labelValues)), "value"),
				nil),
			prometheus.UntypedValue,
			1,
			append(c.labelValues, v)...,
		))
	case []interface{}:
		for i, v := range v {
			c.withLabelValue(strconv.FormatInt(int64(i), 10)).collectJsonValue(v)
		}
	default:
		c.metricError(fmt.Errorf("unhandled json value type %T", v))
	}
}

func (c expvarVisitor) collectVar(v expvar.Var) {
	//switch _v := v.(type) {
	//case *expvar.Map:
	//	_v.Do(func(kv expvar.KeyValue) {
	//		c.withLabelValue(kv.Key).collectVar(kv.Value)
	//	})
	//	return
	//}
	var jv interface{}
	if err := json.Unmarshal([]byte(v.String()), &jv); err != nil {
		c.metricError(fmt.Errorf("error unmarshaling Var json: %w", err))
	}
	c.collectJsonValue(jv)
}
