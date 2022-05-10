package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"io"
)

type Config struct {
	ServiceName string `mapstructure:"serviceName" validate:"required"`
	HostPort    string `mapstructure:"hostPort" validate:"required"`
	Enable      bool   `mapstructure:"enable"`
	LogSpans    bool   `mapstructure:"logSpans"`
}

func NewJaegerTracer(jaegerConfig *Config) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		ServiceName: jaegerConfig.ServiceName,

		// "const" sampler is a binary sampling strategy: 0=never sample, 1=always sample.
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},

		// Log the emitted spans to stdout.
		Reporter: &config.ReporterConfig{
			LogSpans:           jaegerConfig.LogSpans,
			LocalAgentHostPort: jaegerConfig.HostPort,
		},
	}

	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	return cfg.NewTracer(
		config.Logger(jaeger.StdLogger),
		config.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		config.Injector(opentracing.TextMap, zipkinPropagator),
		config.Injector(opentracing.Binary, zipkinPropagator),
		config.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		config.Extractor(opentracing.TextMap, zipkinPropagator),
		config.Extractor(opentracing.Binary, zipkinPropagator),
		config.ZipkinSharedRPCSpan(false),
	)
}
