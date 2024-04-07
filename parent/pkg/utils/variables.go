package utils

type Configurations struct {
	HttpPort                 string `split_words:"true" default:"8080"`
	OtelExporterOtlpEndpoint string `split_words:"true" default:"otel-collector.observability.svc.cluster.local:4317"`
}

var Config Configurations
