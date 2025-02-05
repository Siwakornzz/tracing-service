
# Tracing Service

This project sets up a **tracing service** using **OpenTelemetry** with Jaeger for distributed tracing. It provides a service that receives trace data from other services (acting as a third party) and stores/aggregates that trace data to provide insights into distributed system flows.

## Project Structure

```
📁tracing-service
  └── 📁poc
        └── index.js
        └── package-lock.json
        └── package.json
  └── .gitignore
  └── docker-compose.yml
  └── Dockerfile
  └── go.mod
  └── go.sum
  └── LICENSE
  └── main.go
  └── README.md
```

## How It Works

### 1. **Tracing Service** (`main.go`)

The core service receives trace data from other backend services. It works with OpenTelemetry and sends traces to **Jaeger**.

- **/trace endpoint**: This API endpoint accepts a POST request with trace information. It either:
  - Creates a new root span (if no trace ID exists).
  - Creates a child span linked to a parent span using the provided `trace_id`.
  - It stores spans and trace information to **Jaeger**.

### 2. **Jaeger Integration**

- **Jaeger** is used as the tracing back-end system to aggregate and visualize distributed traces.
- Traces are sent to Jaeger via a configured endpoint.

### 3. **poc** (Proof of Concept)

In this folder, there is a Node.js-based script (`index.js`) that simulates how a backend service might call the tracing service and submit trace data for operations.

## Getting Started

### Prerequisites

- Go 1.18+ (for backend service)
- Node.js and npm (for the poc)
- Docker (to run Jaeger and the services together)

### Setup & Run the Service

#### 1. **Run Jaeger using Docker**

To run Jaeger locally via Docker, run the following command:

```bash
docker-compose up -d
```

This will bring up Jaeger's UI at [localhost:16686](http://localhost:16686) where you can view the traces.

#### 2. **Run Tracing Service**

Navigate to the root folder and run the following command to start the Go service:

```bash
go run main.go
```

#### 3. **Testing the Service (using Node.js Proof of Concept)**

In the `poc` folder, run the Node.js script (`index.js`) to simulate sending traces to the Go tracing service:

```bash
node poc/index.js
```

You should now be able to see traces in the Jaeger UI.

## Configuration

- **Jaeger URL**: The Go tracing service sends traces to the Jaeger collector at `http://localhost:14268/api/traces` by default. This can be modified as needed.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

