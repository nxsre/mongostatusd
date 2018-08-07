# mongostatusd
MongoDB server status daemon

It requires you to have a `config.yaml` in the same working directory as it is running.

In the `config.yaml`, you can specify desired exporters e.g.

```yaml
metrics_report_period: 5s

stackdriver:
    project_id: census-demos
    metric_prefix: mongostatusd

mongodb_uri: mongodb://localhost:27017
mongodb_name: test

prometheus:
    port: 8787
```

which will export the stats to Prometheus and Stackdriver Monitoring

### Installing it
```go
go get -u -v github.com/opencensus-integrations/mongostatusd/cmd/mongostatusd
```

### Running it

Before beginning it, you'll need to setup any of the desired exporters

Exporter|Link
---|---
Stackdriver Monitoring|https://opencensus.io/codelabs/stackdriver/
Prometheus|https://opencensus.io/codelabs/prometheus/
DataDog|https://docs.datadoghq.com/agent/

Once you've setup any of the desired backends and having successfully installed `mongostatusd`, please run

```shell
mongostatusd
```
