package middleware

import (
	"fmt"

	"ai-api-gateway/internal/tracing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates tracing middleware
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// Start span
		spanName := c.Request.Method + " " + c.FullPath()
		ctx, span := tracing.StartSpan(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(c.Request.Method),
				semconv.URLFullKey.String(c.Request.URL.String()),
				semconv.HTTPRouteKey.String(c.FullPath()),
				attribute.String("http.user_agent", c.Request.UserAgent()),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		// Store context in request
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Set span attributes from response
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(c.Writer.Status()),
			attribute.Int("http.response.size", c.Writer.Size()),
		)

		// Set span status
		if c.Writer.Status() >= 500 {
			span.RecordError(fmt.Errorf("HTTP %d", c.Writer.Status()))
			span.SetStatus(codes.Error, "Internal Server Error")
		} else if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "Client Error")
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
	}
}
