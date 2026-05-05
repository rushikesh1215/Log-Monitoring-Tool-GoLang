# 🚀 Distributed Log Monitoring System

A high-performance, fault-tolerant log collection and monitoring system built with **Go**. This project utilizes a **Write-Ahead Log (WAL)** backed agent to ensure zero log loss even during network partitions or server downtime.



## 🏗 System Architecture

The system consists of two primary components:

1.  **Log Agent (Sidecar):**
    *   **Tails** local application logs using the `hpcloud/tail` library.
    *   **Persists** logs into disk-backed segments (WAL) before sending.
    *   **Retries** with exponential backoff until the server acknowledges receipt.
    *   **Metadata Injection:** Automatically tags logs with Node IP and Node Name.
2.  **Log Server (Backend):**
    *   **Gin-based API** to receive log batches.
    *   **Parses** pipe-separated strings (`IP | Node | Level | Message`).
    *   **Provides** endpoints for initial log fetching and real-time streaming.

---

## 🛠 Features

*   **Zero Log Loss:** Logs are committed to a local WAL before network transmission.
*   **Stateful Tailing:** Remembers the last read position in the source file.
*   **Intelligent Parsing:** Automatically detects log levels (INFO, WARN, ERROR, DEBUG) if not explicitly provided.
*   **Efficient Batching:** Transmits logs in configurable batches (100 lines or 512KB) to reduce HTTP overhead.
*   **Security:** Agent-to-Server communication is protected via `X-Agent-Access-Key` validation.



---

## 🚀 Getting Started

### 1. Prerequisites
*   Go 1.21+
*   Shared volume setup (if running in Kubernetes/Docker)

### 2. Running the Server
```bash
# Set environment variables
export PORT=8080
export AGENT_ACCESS_KEY=your_secret_key

go run cmd/server/main.go
```

### 3. Running the Agent
The agent requires environment variables to identify its source.
```bash
export SERVER_URL="http://localhost:8080/api/v1/logs"
export SERVICE_NAME="payment-service"
export LOG_FILE_PATH="/var/log/app/app.log"
export AGENT_ACCESS_KEY=your_secret_key
export NODE_IP="10.0.0.5"
export NODE_NAME="worker-node-01"

go run cmd/agent/main.go
```

---

## 📡 API Reference

### Receive Logs
`POST /api/v1/logs`
*   **Header:** `X-Agent-Access-Key: <key>`
*   **Body:**
```json
{
  "service_name": "upi-gateway",
  "logs": [
    "10.0.0.5 | node-01 | ERROR | Connection timed out",
    "10.0.0.5 | node-01 | INFO | Retry successful"
  ]
}
```
## 📦 Deployment (Kubernetes)

To deploy as a sidecar, ensure both containers share a volume for the log file:

```yaml
volumes:
- name: shared-logs
  emptyDir: {}

# App Container
volumeMounts:
- name: shared-logs
  mountPath: /var/log/app

# Agent Container
volumeMounts:
- name: shared-logs
  mountPath: /mnt/log
env:
- name: LOG_FILE_PATH
  value: "/mnt/log/app.log"
```

---

### Why the WAL-based approach?
Unlike simple log forwarders, this agent is designed for **reliability**. If the central log server goes down for maintenance, the agent buffers logs to the local disk. Once the server returns, the agent resumes transmission exactly where it left off, ensuring your audit trail remains unbroken.
