package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/odigos-io/odigos/common"
)

const (
	elasticApmServerEndpoint = "ELASTIC_APM_SERVER_ENDPOINT"
)

type ElasticAPM struct{}

func (e *ElasticAPM) DestType() common.DestinationType {
	return common.ElasticAPMDestinationType
}

func (e *ElasticAPM) ModifyConfig(dest ExporterConfigurer, currentConfig *Config) ([]string, error) {
	isTlsDisabled := false
	if !e.requiredVarsExists(dest) {
		return nil, errors.New("ElasticAPM config is missing required variables")
	}

	isTlsDisabled = strings.Contains(dest.GetConfig()[elasticApmServerEndpoint], "http://")

	elasticApmEndpoint := e.parseEndpoint(dest.GetConfig()[elasticApmServerEndpoint])

	exporterName := "otlp/elastic-" + dest.GetID()
	currentConfig.Exporters[exporterName] = GenericMap{
		"endpoint": elasticApmEndpoint,
		"tls": GenericMap{
			"insecure": isTlsDisabled,
		},
		"headers": GenericMap{
			"authorization": "Bearer ${ELASTIC_APM_SECRET_TOKEN}",
		},
	}

	var pipelineNames []string
	if isTracingEnabled(dest) {
		tracesPipelineName := "traces/elastic-" + dest.GetID()
		currentConfig.Service.Pipelines[tracesPipelineName] = Pipeline{
			Exporters: []string{exporterName},
		}
		pipelineNames = append(pipelineNames, tracesPipelineName)
	}

	if isMetricsEnabled(dest) {
		metricsPipelineName := "metrics/elastic-" + dest.GetID()
		currentConfig.Service.Pipelines[metricsPipelineName] = Pipeline{
			Exporters: []string{exporterName},
		}
		pipelineNames = append(pipelineNames, metricsPipelineName)
	}

	if isLoggingEnabled(dest) {
		logsPipelineName := "logs/elastic-" + dest.GetID()
		currentConfig.Service.Pipelines[logsPipelineName] = Pipeline{
			Exporters: []string{exporterName},
		}
		pipelineNames = append(pipelineNames, logsPipelineName)
	}

	return pipelineNames, nil
}

func (e *ElasticAPM) requiredVarsExists(dest ExporterConfigurer) bool {
	if _, ok := dest.GetConfig()[elasticApmServerEndpoint]; !ok {
		return false
	}

	return true
}

func (e *ElasticAPM) parseEndpoint(endpoint string) string {
	var port = "8200"
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpointDetails := strings.Split(endpoint, ":")
	host := endpointDetails[0]
	if len(endpointDetails) > 1 {
		port = endpointDetails[1]
	}
	return fmt.Sprintf("%s:%s", host, port)
}
