package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

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
	tp, err := setupTracerProvider("http://localhost:14268/api/traces")
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
		operation := fmt.Sprintf("%v", traceData["operation"])

		var ctx context.Context
		var parentSpan trace.Span

		// 🟢 ถ้าหา Trace ID ไม่เจอ ให้สร้างใหม่
		if traceID == "<nil>" || traceID == "" {
			traceID = generateNewTraceID()
			traceData["trace_id"] = traceID
		}

		// 🟢 เช็คว่ามี Parent Trace ไหม?
		root, exists := traceGroups.Load(traceID)
		if exists {
			parentSpan = root.(trace.Span)
			ctx = trace.ContextWithSpan(context.Background(), parentSpan)
		} else {
			// 🟢 ถ้าไม่มี Parent → เป็น Root Span
			ctx, parentSpan = tracer.Start(context.Background(), "RootTrace-"+traceID)
			traceGroups.Store(traceID, parentSpan)
		}

		// 🕒 อ่าน Start Time & End Time จาก Request
		startTime, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", traceData["start_time"]))
		endTime, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", traceData["end_time"]))

		// 🟡 กำหนด Parent Span ให้ Context
		if parentSpan.SpanContext().IsValid() {
			ctx = trace.ContextWithSpan(ctx, parentSpan)
		}

		// 🟡 Start Child Span
		opts := []trace.SpanStartOption{
			trace.WithTimestamp(startTime),
		}

		_, span := tracer.Start(ctx, operation, opts...)
		defer span.End(trace.WithTimestamp(endTime))

		// 📌 Set Attributes ของ Span
		for key, value := range traceData {
			span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
		}

		// ✅ เก็บ Parent Span ไว้ เพื่อให้ Request ถัดไปใช้
		traceGroups.Store(traceID, span)

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
