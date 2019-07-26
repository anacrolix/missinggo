package xprometheus

import (
	"encoding/json"
	"expvar"
	"fmt"
	"strconv"

	"github.com/bradfitz/iter"
	"github.com/prometheus/client_golang/prometheus"
)

// A Prometheus collector that exposes all vars.
type expvarCollector struct {
	descs map[int]*prometheus.Desc
}

func NewExpvarCollector() expvarCollector {
	return expvarCollector{
		descs: make(map[int]*prometheus.Desc),
	}
}

const (
	fqName = "go_expvar"
	help   = "All expvars"
)

var desc = prometheus.NewDesc(fqName, help, nil, nil)

// Describe implements Collector.
func (e expvarCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- desc
}

// Collect implements Collector.
func (e expvarCollector) Collect(ch chan<- prometheus.Metric) {
	expvar.Do(func(kv expvar.KeyValue) {
		collector{
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

type collector struct {
	f           func(prometheus.Metric)
	labelValues []string
	descs       map[int]*prometheus.Desc
}

func (c *collector) newMetric(f float64) {
	c.f(prometheus.MustNewConstMetric(
		c.desc(),
		prometheus.UntypedValue,
		float64(f),
		c.labelValues...))
}

func (c collector) desc() *prometheus.Desc {
	d, ok := c.descs[len(c.labelValues)]
	if !ok {
		d = prometheus.NewDesc(fqName, "", labels(len(c.labelValues)), nil)
		c.descs[len(c.labelValues)] = d
	}
	return d
}

func (c collector) metricError(err error) {
	c.f(prometheus.NewInvalidMetric(c.desc(), err))
}

func (c collector) withLabelValue(lv string) collector {
	//if !utf8.ValidString(lv) {
	//	lv = strconv.Quote(lv)
	//}
	c.labelValues = append(c.labelValues, lv)
	return c
}

func (c collector) collectJsonValue(v interface{}) {
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

func (c collector) collectVar(v expvar.Var) {
	//switch _v := v.(type) {
	//case *expvar.Map:
	//	_v.Do(func(kv expvar.KeyValue) {
	//		c.withLabelValue(kv.Key).collectVar(kv.Value)
	//	})
	//	return
	//}
	var jv interface{}
	if err := json.Unmarshal([]byte(v.String()), &jv); err != nil {
		c.metricError(fmt.Errorf("error unmarshaling Var json: %s", err))
	}
	c.collectJsonValue(jv)
}
