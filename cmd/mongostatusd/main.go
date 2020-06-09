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
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	"contrib.go.opencensus.io/exporter/ocagent"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"

	"github.com/opencensus-integrations/mongostatusd"
)

func main() {
	configPath := flag.String("config", "", "the path to the YAML configuration file")
	flag.Parse()
	yamlBlob, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Failed to read the config.yaml file: %v", err)
	}

	config := new(Config)
	if err := yaml.Unmarshal(yamlBlob, config); err != nil {
		log.Fatalf("Failed to parse yaml file content: %v", err)
	}

	if err := enableOpenCensus(config); err != nil {
		log.Fatalf("Failed to enable OpenCensus: %v", err)
	}

	mongoDBURI := config.MongoDBURI
	if mongoDBURI == "" {
		mongoDBURI = "mongodb://localhost:27017"
	} else if !strings.HasPrefix(mongoDBURI, "mongodb://") {
		mongoDBURI = "mongodb://" + mongoDBURI
	}
	mc, err := mongo.NewClient(options.Client().ApplyURI(mongoDBURI))
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
	// Watch for any signal e.g. Ctrl-C, then stop the watcher
	if false {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan)
		go func() {
			notif := <-sigChan
			log.Printf("Signal notified: %d, %v", notif, notif)

			stopChan <- true
			close(stopChan)
			log.Printf("Successfully sent stop signal")
		}()
	}

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
}

func enableOpenCensus(cfg *Config) error {
	if err := view.Register(mongostatusd.AllViews...); err != nil {
		return err
	}
	oce, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithAddress("localhost:55678"),
		ocagent.WithServiceName("mongostatusd"),
	)
	if err != nil {
		return err
	}

	metricsReportingPeriod := cfg.MetricsReportPeriod
	if metricsReportingPeriod <= 0 {
		metricsReportingPeriod = 5 * time.Second
	}
	view.SetReportingPeriod(metricsReportingPeriod)
	view.RegisterExporter(oce)
	trace.RegisterExporter(oce)

	return nil
}
