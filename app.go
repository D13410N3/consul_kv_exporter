package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v2"
)

// Config represents the YAML configuration format
type Config struct {
	DCs map[string]DCConfig `yaml:"dc"`
}

// DCConfig represents the configuration for a specific data center
type DCConfig struct {
	Directories []string `yaml:"directories"`
}

var (
	consulKvModifyIndex = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "consul_kv_modify_index",
			Help: "Consul KV Modify Index",
		},
		[]string{"dc", "key"},
	)
)

func init() {
	prometheus.MustRegister(consulKvModifyIndex)
}

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		log.Fatal("CONFIG_FILE environment variable is not defined")
	}

	baseURI := os.Getenv("CONSUL_BASE_URI")
	if baseURI == "" {
		log.Fatal("CONSUL_BASE_URI environment variable is not defined")
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		log.Fatal("LISTEN_ADDR environment variable is not defined")
	}

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	http.Handle("/metrics", promhttp.Handler())

	for dc, dcConfig := range config.DCs {
		for _, directory := range dcConfig.Directories {
			go collectMetrics(dc, directory, baseURI)
		}
	}

	log.Printf("Starting Prometheus exporter on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func loadConfig(configFile string) (*Config, error) {
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	err = yaml.Unmarshal(content, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	return config, nil
}

func collectMetrics(dc, directory, baseURI string) {
	for {
		url := fmt.Sprintf("%s/v1/kv/%s/?recurse&dc=%s", baseURI, directory, dc)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Failed to fetch data for directory '%s' in DC '%s': %v", directory, dc, err)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body for directory '%s' in DC '%s': %v", directory, dc, err)
			continue
		}

		// Parse the response body as JSON
		var entries []map[string]interface{}
		err = json.Unmarshal(body, &entries)
		if err != nil {
			log.Printf("Failed to parse response body for directory '%s' in DC '%s': %v", directory, dc, err)
			continue
		}

		for _, entry := range entries {
			key, ok := entry["Key"].(string)
			if !ok {
				log.Println("Failed to parse 'Key' field from response")
				continue
			}

			modifyIndex, ok := entry["ModifyIndex"].(float64)
			if !ok {
				log.Println("Failed to parse 'ModifyIndex' field from response")
				continue
			}

			// Add the metric
			consulKvModifyIndex.WithLabelValues(dc, key).Set(modifyIndex)
		}

		// Sleep for 5 seconds before fetching the metrics again
		time.Sleep(5 * time.Second)
	}
}
