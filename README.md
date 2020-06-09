# mongostatusd
MongoDB server status daemon

It requires you to have a `config.yaml` in the same working directory as it is running.

In the `config.yaml`, you can specify desired exporters e.g.

```yaml
metrics_report_period: 5s

mongodb_uri: mongodb://localhost:27017
mongodb_name: test
```

which will export to an [OpenCensus agent](https://opencensus.io/agent/) which by default should be at "localhost:55678".

### Installing it
```go
go get -u -v github.com/opencensus-integrations/mongostatusd/cmd/mongostatusd
```

### Running it

```shell
mongostatusd --config config.yaml
```
