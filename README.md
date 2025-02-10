# Tracing Service

This project is a Proof of Concept (PoC) implementation of a tracing service using OpenTelemetry and Jaeger for distributed tracing. It provides an example of how trace data can be collected, stored, and visualized in a distributed system. This is not a production release.stem flows.

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

 ## Endpoints:
- /start-trace → Starts a new root trace.
- /add-trace → Adds a child span to an existing trace.
- /stop-trace → Ends a span and removes it from memory.

### 2. **Jaeger Integration**

- **Jaeger** is used as the tracing back-end system to aggregate and visualize distributed traces.
- The tracing data is collected and sent to Jaeger for visualization.
- The Jaeger UI is available at http://localhost:16686 for viewing trace data.

### 3. **Node.js Proof of Concept (poc/index.js)**

In this folder, there is a Node.js-based script (`index.js`) that simulates how a backend service might call the tracing service and submit trace data for operations.
- This is a simple script that simulates a backend service interacting with the tracing service.
- It creates traces for user creation, database insertion, and email confirmation steps.
- Not intended for production use.

## Getting Started

### Prerequisites

- Go 1.18+ (for backend service)
- Node.js and npm (for the poc)
- Docker (for running Jaeger and related services)

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
- **Modify**:  main.go if a different collector is required.

## Limitations
- This project is a Proof of Concept (PoC) and not production-ready.
- There is no persistence for trace data beyond what is stored in Jaeger.
- It is only intended for demonstration and testing purposes.
## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
