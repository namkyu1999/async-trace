package utils

type Configurations struct {
	ParentServicePort        string `split_words:"true" default:"8080"`
	OtelExporterOtlpEndpoint string `split_words:"true"`
	OtelServiceName          string `split_words:"true" default:"parentservice"`
}

var Config Configurations
