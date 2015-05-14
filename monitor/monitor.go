package monitor

import (
	"fmt"
	"time"

	"code.google.com/p/goauth2/compute/serviceaccount"
	"google.golang.org/api/cloudmonitoring/v2beta2"
)

const (
	Prefix = "custom.cloudmonitoring.googleapis.com/"
)

var EmptyPoint = cloudmonitoring.Point{
	DoubleValue: 0,
	Start:       "",
	End:         "",
}

type MonitorOpts struct {
	ProjectID string
	Prefix    string
	Account   string
	Interval  time.Duration
}

type Monitor struct {
	MonitorOpts
	tsSvc   *cloudmonitoring.TimeseriesService
	mdSvc   *cloudmonitoring.MetricDescriptorsService
	request *cloudmonitoring.WriteTimeseriesRequest
}

type MetricOpts struct {
	Name   string
	Labels map[string]string
}

type Metric struct {
	MetricOpts
	*cloudmonitoring.TimeseriesPoint
}

func NewMonitor(opts MonitorOpts) (m *Monitor, err error) {
	svcOpts := &serviceaccount.Options{
		Account: opts.Account,
	}

	client, err := serviceaccount.NewClient(svcOpts)
	if err != nil {
		return nil, err
	}

	svc, err := cloudmonitoring.New(client)
	if err != nil {
		return nil, err
	}

	tsSvc := cloudmonitoring.NewTimeseriesService(svc)
	mdSvc := cloudmonitoring.NewMetricDescriptorsService(svc)
	request := &cloudmonitoring.WriteTimeseriesRequest{
		Timeseries: make([]*cloudmonitoring.TimeseriesPoint, 0),
	}

	m = &Monitor{opts, tsSvc, mdSvc, request}
	go m.start()

	return m, nil
}

func (m *Monitor) start() {
	time.Sleep(m.Interval / 2)

	for _ = range time.Tick(m.Interval) {
		err := m.flush()
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (m *Monitor) flush() (err error) {
	call := m.tsSvc.Write(m.ProjectID, m.request)

	_, err = call.Do()
	if err != nil {
		return err
	}

	return nil
}

func (m *Monitor) NewMetric(opts MetricOpts) (e *Metric, err error) {
	if err := m.create(opts); err != nil {
		return nil, err
	}

	lbls := map[string]string{}
	for lblName, value := range opts.Labels {
		key := Prefix + opts.Name + "-" + lblName
		lbls[key] = value
	}

	p := &cloudmonitoring.TimeseriesPoint{
		TimeseriesDesc: &cloudmonitoring.TimeseriesDescriptor{
			Metric:  Prefix + m.Prefix + opts.Name,
			Project: m.ProjectID,
			Labels:  lbls,
		},
		Point: &cloudmonitoring.Point{
			DoubleValue: 0,
			Start:       "",
			End:         "",
		},
	}

	m.request.Timeseries = append(m.request.Timeseries, p)

	return &Metric{opts, p}, nil
}

func (m *Monitor) create(opts MetricOpts) (err error) {
	lbls := []*cloudmonitoring.MetricDescriptorLabelDescriptor{}
	for lblName, _ := range opts.Labels {
		desc := &cloudmonitoring.MetricDescriptorLabelDescriptor{
			Key:         Prefix + opts.Name + "-" + lblName,
			Description: opts.Name + "-" + lblName,
		}

		lbls = append(lbls, desc)
	}

	call := m.mdSvc.Create(m.ProjectID, &cloudmonitoring.MetricDescriptor{
		Name:        Prefix + m.Prefix + opts.Name,
		Description: opts.Name,
		Project:     m.ProjectID,
		Labels:      lbls,
		TypeDescriptor: &cloudmonitoring.MetricDescriptorTypeDescriptor{
			MetricType: "gauge",
			ValueType:  "double",
		},
	})

	_, err = call.Do()
	if err != nil {
		return err
	}

	return nil
}

func (m *Metric) Measure(value float64) {
	now := time.Now().Format(time.RFC3339)
	m.Point.DoubleValue = value
	m.Point.Start = now
	m.Point.End = now
}
