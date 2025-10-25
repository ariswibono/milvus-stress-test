## Milvus Load Testing Tool (Go)

A professional-grade load testing tool for Milvus vector databases. Features duration-based testing, configurable load intensity, gradual load increase, and real-time monitoring. Perfect for production readiness testing and performance benchmarking.

### Features
- **Duration-based testing**: Run tests for specified time periods
- **Configurable load intensity**: Low, medium, high, extreme pressure levels
- **Gradual load increase**: Ramp-up testing to find performance limits
- **Real-time monitoring**: Live throughput and performance metrics
- **Production-ready**: Comprehensive testing scenarios for production deployment
- **Professional output**: Clean, engineering-focused reporting

### Prerequisites
- Go 1.21+ (tested with Go 1.25)
- Docker and Docker Compose (or Podman with compose compatibility)

### Start Milvus with Docker Compose
This repo includes a `compose.yml` that brings up Milvus and dependencies (MinIO, etcd) with persisted volumes in `volumes/`.

```bash
docker compose -f compose.yml up -d
# or with Podman
podman compose -f compose.yml up -d
```

Milvus should become available at `localhost:19530` by default.

To stop:
```bash
docker compose -f compose.yml down
```

### Quick Start

Show help and available options:
```bash
go run main.go --help
```

Basic 30-second medium load test:
```bash
go run main.go
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--milvus-addr` | Milvus server address | `localhost:19530` |
| `--duration` | Test duration (30s, 2m, 1h) | `30s` |
| `--pressure` | Load intensity (low, medium, high, extreme) | `medium` |
| `--ramp-up` | Gradually increase load from 10% to 100% | `false` |
| `--real-time` | Display real-time throughput metrics | `false` |
| `--help` | Show detailed help information | - |

### Load Intensity Levels

| Level | Workers | Batch Size | Use Case |
|-------|---------|------------|----------|
| `low` | 5 | 500 | Development, baseline testing |
| `medium` | 20 | 2000 | Production simulation |
| `high` | 50 | 5000 | Peak load testing |
| `extreme` | 100 | 10000 | Stress testing, breaking points |

## 🏭 Production-Ready Testing

### Essential Test Scenarios

#### 1. Baseline Performance Test
```bash
# Establish baseline metrics
go run main.go --duration 5m --pressure medium --real-time
```

#### 2. Peak Load Test
```bash
# Test maximum sustained load
go run main.go --duration 10m --pressure high --real-time
```

#### 3. Stress Test (Find Breaking Point)
```bash
# Gradually increase load to find limits
go run main.go --duration 15m --pressure high --ramp-up --real-time
```

#### 4. Endurance Test
```bash
# Long-term stability test
go run main.go --duration 1h --pressure medium --real-time
```

#### 5. Extreme Load Test
```bash
# Maximum possible load
go run main.go --duration 5m --pressure extreme --real-time
```

### Production-Specific Scenarios

#### E-commerce (Black Friday Simulation)
```bash
go run main.go --duration 2h --pressure high --ramp-up --real-time
```

#### Real-time Analytics
```bash
go run main.go --duration 4h --pressure high --real-time
```

#### ML/AI Workloads
```bash
go run main.go --duration 1h --pressure extreme --real-time
```

### Key Metrics to Monitor

- **Throughput**: Operations per second
- **Latency**: Response times  
- **Error Rate**: Failed operations percentage
- **Resource Usage**: CPU, Memory, Disk I/O

### Red Flags to Watch For

- Throughput drops below 80% of baseline
- Latency increases by more than 50%
- Error rate exceeds 1%
- CPU usage consistently above 90%
- Memory usage growing continuously

### What the Program Does
1. Connects to Milvus using the Go SDK.
2. Drops the collection `go_high_throughput_collection` if it exists.
3. Creates the collection with schema:
   - `id` (Int64, primary key, AutoID)
   - `embedding` (FloatVector, dim=8)
4. Inserts randomly generated embeddings concurrently in batches.
5. Flushes the collection.
6. Creates IVF_FLAT index on `embedding` (L2, nlist=16) and waits for completion.
7. Loads the collection into memory.
8. Executes concurrent searches (topk=3, nprobe=10) using random query vectors.
9. Prints throughput metrics and a final summary.
10. Drops the collection to clean up.

### Sample Output

#### Basic Test
```bash
go run main.go --duration 30s --pressure medium --real-time
```

```
>> Starting Milvus Load Test: MEDIUM intensity for 30s <<

--- Test Configuration ---
 - Milvus Address:                  localhost:19530
 - Test Duration:                   30s
 - Load Intensity:                  MEDIUM
 - Concurrent Workers:              20
 - Batch Size (Vectors per Insert): 2000
 - Test Mode:                       Continuous load until duration expires
----------------------------------------

--- Step 4: Starting continuous data insertion for 30s ---
[Worker 0] Starting continuous insertion...
[Worker 1] Starting continuous insertion...
...
📊 [5s] Batch Size: 2000, Throughput: 1250.3 ops/sec
📊 [10s] Batch Size: 2000, Throughput: 2100.7 ops/sec
✅ All workers finished inserting data in 30.123456s.
   -> Total vectors inserted: 1250000
   -> Throughput: 41500.00 inserts/second
```

#### Professional Summary Table
```
================================================================================
                        LOAD TEST PERFORMANCE SUMMARY
================================================================================
│ Configuration              │ Value                                            │
├─────────────────────────────┼──────────────────────────────────────────────────┤
│ Test Duration               │ 30s                                              │
│ Load Intensity              │ MEDIUM                                           │
│ Milvus Address              │ localhost:19530                                 │
│ Concurrent Workers          │ 20                                               │
│ Batch Size                  │ 2000                                             │
│ Vectors Inserted            │ 1250000                                          │
│ Data Size Inserted          │ 38.15 MB                                         │
│ Searches Performed          │ 500                                              │
├─────────────────────────────┼──────────────────────────────────────────────────┤
│ Performance Metrics         │ Value                                            │
├─────────────────────────────┼──────────────────────────────────────────────────┤
│ Total Elapsed Time          │ 45.123456s                                       │
│ Data Insertion Time         │ 30.123456s                                       │
│ Insert Throughput           │ 41500.00                                         │
│ Search Throughput            │ 125.50                                          │
================================================================================
```

Note: Actual performance depends on your hardware and Milvus configuration.

### Implementation Notes
- Embedding dimension is `8` for speed and demonstration.
- Index used: IVF_FLAT (L2), `nlist=16`; searches use `nprobe=10` and `topk=3`.
- Data is random float32 vectors; total inserted vectors are derived from:
  `(numWorkers × batchesPerWorker × batchSize)`.

### Troubleshooting
- Ensure Milvus is healthy and reachable at `--milvus-addr`.
- If using Podman, confirm the socket/compose plugin is configured.
- If you see timeouts, give Milvus extra startup time after `compose up -d`.
- For very high settings, ensure sufficient CPU/RAM.

### Cleanup
The program drops the test collection at the end. If it terminates early, you can manually drop `go_high_throughput_collection` via your Milvus client/UI.
