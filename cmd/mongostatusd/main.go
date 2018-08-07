// Copyright 2018, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/DataDog/opencensus-go-exporter-datadog"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats/view"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/opencensus-integrations/mongostatusd"
)

func main() {
	yamlBlob, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to read the config.yaml file: %v", err)
	}

	config := new(Config)
	if err := yaml.Unmarshal(yamlBlob, config); err != nil {
		log.Fatalf("Failed to parse yaml file content: %v", err)
	}

	flushFn, err := enableOpenCensus(config)
	if err != nil {
		log.Fatalf("Failed to enable OpenCensus: %v", err)
	}
	defer func() {
		log.Println("About to flush OpenCensus exporters")
		flushFn()
		log.Println("Flushed OpenCensus exporters")
	}()

	mongoDBURI := config.MongoDBURI
	if mongoDBURI == "" {
		mongoDBURI = "mongodb://localhost:27017"
	} else if !strings.HasPrefix(mongoDBURI, "mongodb://") {
		mongoDBURI = "mongodb://" + mongoDBURI
	}
	mc, err := mongo.NewClient(mongoDBURI)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}
	if err := mc.Connect(context.Background()); err != nil {
		log.Fatalf("Failed to connect to the MongoDB server: %v", err)
	}
	defer func() {
		log.Println("Disconnecting from Mongo...")
		// As of `Mon  6 Aug 2018 16:07:50 PDT` and
		// MongoDB Go driver commit: 44fa48dcf49c6ab707da1359c640383fc0c42e86
		// (*Client).Disconnect takes an indefinite time to return
		// so throw it in a goroutine
		go mc.Disconnect(context.Background())
		log.Println("Disconnected from Mongo!")
	}()

	cfg := &mongostatusd.Config{MongoDBName: config.MongoDBName}

	stopChan := make(chan bool, 1)
	go func() {
		// Watch for any signal e.g. Ctrl-C, then stop the watcher
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan)
		notif := <-sigChan
		log.Printf("Signal notified: %v", notif)

		stopChan <- true
		close(stopChan)
		log.Printf("Successfully sent stop signal")
	}()

	errsChan := cfg.Watch(mc, stopChan)
	for err := range errsChan {
		if err != nil {
			log.Printf("Encountered error: %v", err)
		}
	}
}

type Config struct {
	MongoDBName string `yaml:"mongodb_name"`
	MongoDBURI  string `yaml:"mongodb_uri"`

	MetricsReportPeriod time.Duration `yaml:"metrics_report_period"`

	DataDog     *DataDog     `yaml:"datadog"`
	Prometheus  *Prometheus  `yaml:"prometheus"`
	Stackdriver *Stackdriver `yaml:"stackdriver"`
	Zipkin      *Zipkin      `yaml:"zipkin"`
}

type Prometheus struct {
	Port      int    `yaml:"port"`
	Namespace string `yaml:"namespace"`
}

type Stackdriver struct {
	MetricPrefix string `yaml:"metric_prefix"`
	ProjectID    string `yaml:"project_id"`
}

type Zipkin struct {
	LocalEndpointURI string `yaml:"local_endpoint"`
	ReporterURI      string `yaml:"reporter_uri"`
	ServiceName      string `yaml:"service_name"`
}

type DataDog struct {
	Namespace       string   `yaml:"namespace"`
	Service         string   `yaml:"service_name"`
	StatsAddressURI string   `yaml:"stats_addr"`
	Tags            []string `yaml:"tags"`
	TraceAddressURI string   `yaml:"trace_addr"`
}

func enableOpenCensus(cfg *Config) (flushFn func(), err error) {
	if err := view.Register(mongostatusd.AllViews...); err != nil {
		return nil, err
	}

	var flushFns []func()

	defer func() {
		if len(flushFns) > 0 {
			flushFn = func() {
				for _, fn := range flushFns {
					fn()
				}
			}
		}
	}()

	if scf := cfg.Stackdriver; scf != nil {
		sd, err := stackdriver.NewExporter(stackdriver.Options{
			MetricPrefix: scf.MetricPrefix,
			ProjectID:    scf.ProjectID,
		})
		if err != nil {
			return nil, err
		}
		view.RegisterExporter(sd)
		flushFns = append(flushFns, sd.Flush)
	}

	metricsReportingPeriod := cfg.MetricsReportPeriod
	if metricsReportingPeriod <= 0 {
		metricsReportingPeriod = 5 * time.Second
	}
	view.SetReportingPeriod(metricsReportingPeriod)

	if pcf := cfg.Prometheus; pcf != nil {
		var pe *prometheus.Exporter
		pe, err = prometheus.NewExporter(prometheus.Options{
			Namespace: pcf.Namespace,
		})
		if err != nil {
			return
		}
		var ln net.Listener
		ln, err = net.Listen("tcp", fmt.Sprintf("localhost:%d", pcf.Port))
		if err != nil {
			return
		}

		go func() {
			defer ln.Close()

			mux := http.NewServeMux()
			mux.Handle("/metrics", pe)
			if err := http.Serve(ln, mux); err != nil {
				log.Fatalf("Prometheus exporter serve error: %v", err)
			}
		}()
		// If flush is invoked, kill the Prometheus server/listener
		flushFns = append(flushFns, func() { ln.Close() })
		view.RegisterExporter(pe)
	}

	if dcf := cfg.DataDog; dcf != nil {
		de := datadog.NewExporter(datadog.Options{
			Namespace: dcf.Namespace,
			Service:   dcf.Service,
			TraceAddr: dcf.TraceAddressURI,
			StatsAddr: dcf.StatsAddressURI,
			Tags:      dcf.Tags,
		})
		view.RegisterExporter(de)
		flushFns = append(flushFns, de.Stop)
	}

	return
}
