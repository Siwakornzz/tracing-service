package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
	traceGroups sync.Map // เก็บ trace_id และ spans
)

func main() {
	tp, err := setupTracerProvider("http://localhost:14268/api/traces")
	if err != nil {
		log.Fatalf("failed to setup TracerProvider: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer = tp.Tracer("tracing-service")
	app := fiber.New()

	app.Post("/start-trace", func(c *fiber.Ctx) error {
		traceData := make(map[string]interface{})
		if err := c.BodyParser(&traceData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		operation := fmt.Sprintf("%v", traceData["operation"])
		message := fmt.Sprintf("%v", traceData["message"])
		startTime, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", traceData["start_time"]))

		// สร้าง trace_id ใหม่
		traceID := uuid.New().String() // สร้าง trace_id ใหม่

		// สร้าง Root Span
		ctx := context.Background()
		_, span := tracer.Start(ctx, operation, trace.WithTimestamp(startTime))

		span.SetAttributes(attribute.String("message", message))

		spanID := span.SpanContext().SpanID().String() // ดึง span_id
		traceGroups.Store(spanID, span)                // เก็บ spanID ของ A

		log.Println("[Start] Span ID:", spanID)
		log.Println("[Start] traceID ID:", traceID)

		// ส่ง trace_id กลับไปพร้อม span_id
		return c.JSON(fiber.Map{
			"status":   "trace started",
			"trace_id": traceID, // ส่ง trace_id ที่สร้างใหม่
			"span_id":  spanID,
		})
	})

	// Add a new span (Child Span)
	app.Post("/add-trace", func(c *fiber.Ctx) error {
		traceData := make(map[string]interface{})
		if err := c.BodyParser(&traceData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		traceID := fmt.Sprintf("%v", traceData["trace_id"])
		parentSpanID := fmt.Sprintf("%v", traceData["parent_span_id"]) // ✅ ใช้ parent span ID
		operation := fmt.Sprintf("%v", traceData["operation"])
		startTime, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", traceData["start_time"]))
		message := fmt.Sprintf("%v", traceData["message"]) // เอา message มาจาก input

		// ✅ หาว่า parent span อยู่ที่ไหน
		parentSpan, exists := traceGroups.Load(parentSpanID)
		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "parent_span_id not found"})
		}

		// ✅ ใช้ parent span เป็น context
		ctx := trace.ContextWithSpan(context.Background(), parentSpan.(trace.Span))
		_, childSpan := tracer.Start(ctx, operation, trace.WithTimestamp(startTime))

		childSpan.SetAttributes(attribute.String("message", message)) // ใส่ message เป็น tag

		traceGroups.Store(childSpan.SpanContext().SpanID().String(), childSpan)

		spandId := childSpan.SpanContext().SpanID().String()

		log.Println("[ADD] Span ID:", spandId)
		log.Println("[ADD] traceID ID:", traceID)

		return c.JSON(fiber.Map{
			"status":   "span added",
			"trace_id": traceID,
			"span_id":  spandId,
		})
	})

	// Stop a span
	app.Post("/stop-trace", func(c *fiber.Ctx) error {
		traceData := make(map[string]interface{})
		if err := c.BodyParser(&traceData); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		spanID := fmt.Sprintf("%v", traceData["span_id"])
		endTime, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", traceData["end_time"]))

		span, exists := traceGroups.Load(spanID)
		if !exists {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "span_id not found"})
		}

		span.(trace.Span).End(trace.WithTimestamp(endTime))
		traceGroups.Delete(spanID) // ลบ span ออกจาก memory

		log.Println("[STOP] Span ID:", spanID)

		return c.JSON(fiber.Map{"status": "span stopped", "span_id": spanID})
	})

	log.Println("🚀 Tracing Service started at :5001")
	log.Fatal(app.Listen(":5001"))
}

// ตั้งค่า Tracer Provider
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
