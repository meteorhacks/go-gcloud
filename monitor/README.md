# Monitoring Custom Metrics

Monitor custom metrics with Google Cloud Monitoring service. At the moment only works for `gauges` and only on Google Compute Engine vms with access to "https://www.googleapis.com/auth/monitoring".

## Example

```
package main

import (
  "github.com/meteorhacks/go-gcloud/monitor"
)

func main() {
  m, err := monitor.NewMonitor(monitor.MonitorOpts{
    ProjectID: "my-project-id",
    Prefix:    "test_",
  })

  if err != nil {
    panic(err)
  }

  mSomething := m.NewMetric(monitor.MetricOpts{
    Name:   "something"
    Labels: map[string]string{"foo": "bar"}
  })

  mSomething.Measure(12.34)

  err = m.Flush()
  if err != nil {
    panic(err)
  }
}
```

## TODO

 - Lots of stuff
