package middleware

import "go.opentelemetry.io/otel/propagation"

func propagatorForTest() propagation.TextMapPropagator {
	return propagation.TraceContext{}
}
