package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer      trace.Tracer
	traceGroups sync.Map
)

func main() {
	tp, err := setupTracerProvider("http://jaeger:14268/api/traces")
	if err != nil {
		log.Fatalf("failed to setup TracerProvider: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer = tp.Tracer("tracing-service")

	app := fiber.New()

	app.Post("/trace", func(c *fiber.Ctx) error {
		traceData := make(map[string]interface{})
		if err := c.BodyParser(&traceData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		traceID := fmt.Sprintf("%v", traceData["trace_id"])
		parentSpanID := fmt.Sprintf("%v", traceData["parent_span_id"])

		log.Printf("🟡 Received Trace ID: %s, Parent Span ID: %s", traceID, parentSpanID)

		if traceID == "<nil>" || traceID == "" {
			traceID = generateNewTraceID()
			traceData["trace_id"] = traceID
		}

		var ctx context.Context
		var parentSpan trace.Span

		root, exists := traceGroups.Load(traceID)
		if exists {
			parentSpan = root.(trace.Span)
			ctx = trace.ContextWithSpan(context.Background(), parentSpan)
		} else {
			ctx, parentSpan = tracer.Start(context.Background(), "RootTrace-"+traceID)
			traceGroups.Store(traceID, parentSpan)
		}

		opts := []trace.SpanStartOption{}
		if parentSpanID != "<nil>" && parentSpanID != "" {
			spanID, err := trace.SpanIDFromHex(parentSpanID)
			if err != nil {
				log.Printf("Error parsing SpanID from hex: %v", err)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid parent span ID"})
			}

			opts = append(opts, trace.WithLinks(trace.Link{
				SpanContext: trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: parentSpan.SpanContext().TraceID(),
					SpanID:  spanID,
					Remote:  true,
				}),
			}))
		}

		_, span := tracer.Start(ctx, fmt.Sprintf("%v", traceData["operation"]), opts...)
		defer span.End()

		for key, value := range traceData {
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
		}

		return c.JSON(fiber.Map{
			"status":   "traced!",
			"trace_id": traceID,
			"span_id":  span.SpanContext().SpanID().String(),
		})
	})

	log.Println("🚀 Tracing Service started at :5001")
	log.Fatal(app.Listen(":5001"))
}

func setupTracerProvider(jaegerURL string) (*sdktrace.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("tracing-service"),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func generateNewTraceID() string {
	tid := trace.TraceID{}
	rand.Read(tid[:])
	return tid.String()
}

func generateNewSpanID() string {
	sid := trace.SpanID{}
	rand.Read(sid[:])
	return sid.String()
}
