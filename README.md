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

| Level | Workers | Batch Size | Real-World Equivalent | Use Case |
|-------|---------|------------|----------------------|----------|
| `low` | 5 | 500 | Small startup, development environment | Development, baseline testing |
| `medium` | 20 | 2000 | Mid-size application, typical production | Production simulation, normal operations |
| `high` | 50 | 5000 | Large enterprise, high-traffic application | Peak load testing, traffic spikes |
| `extreme` | 100 | 10000 | Massive scale, Black Friday traffic | Stress testing, breaking points |

### What These Levels Mean

**`low`** - **Development/Startup Scale**
- Similar to: Small startup with < 1000 users
- Typical throughput: 2,500-5,000 operations/second
- Use for: Development testing, baseline establishment

**`medium`** - **Standard Production Scale**  
- Similar to: Mid-size SaaS application with 10K-100K users
- Typical throughput: 40,000-80,000 operations/second
- Use for: Normal production load simulation

**`high`** - **Enterprise Scale**
- Similar to: Large enterprise with 100K+ users, e-commerce platform
- Typical throughput: 250,000-500,000 operations/second  
- Use for: Peak traffic testing, enterprise workloads

**`extreme`** - **Massive Scale**
- Similar to: Black Friday traffic, social media platforms, global applications
- Typical throughput: 1,000,000+ operations/second
- Use for: Finding absolute limits, stress testing

> **Note**: These levels are relative benchmarks. Actual performance depends on your hardware, Milvus configuration, and data characteristics. Use them as starting points and adjust based on your specific requirements.

## 🧪 Understanding Load Testing Types

### What Each Test Type Actually Means

#### **Baseline Testing**
**What it is**: Establishing your system's normal performance under typical conditions.
**Why it matters**: You need to know your "normal" before you can identify problems.
**What to measure**: 
- Normal throughput (operations/second)
- Typical response times
- Resource usage under normal load
- Error rates during normal operations

**Real-world analogy**: Like checking your car's normal fuel efficiency before a road trip.

#### **Load Testing** 
**What it is**: Testing your system under expected maximum load.
**Why it matters**: Ensures your system can handle your actual production traffic.
**What to measure**:
- Performance under expected peak load
- Whether response times stay acceptable
- If error rates remain low under normal peak conditions

**Real-world analogy**: Like testing your car's performance on a highway at speed limit.

#### **Stress Testing**
**What it is**: Gradually increasing load until you find the breaking point.
**Why it matters**: Discovers your system's limits before users do.
**What to measure**:
- Maximum sustainable load
- Where performance starts to degrade
- How the system behaves near its limits
- Recovery time after overload

**Real-world analogy**: Like testing how fast your car can go before the engine struggles.

#### **Endurance Testing**
**What it is**: Running sustained load for extended periods.
**Why it matters**: Finds memory leaks, resource exhaustion, and long-term stability issues.
**What to measure**:
- Performance consistency over time
- Memory usage trends
- Resource exhaustion
- System stability under sustained load

**Real-world analogy**: Like testing your car's reliability on a long road trip.

#### **Spike Testing**
**What it is**: Sudden, dramatic increases in load (like Black Friday traffic).
**Why it matters**: Real-world traffic isn't smooth - it has sudden spikes.
**What to measure**:
- How quickly the system can scale
- Performance during sudden load increases
- Recovery time after traffic spikes
- Whether the system can handle traffic bursts

**Real-world analogy**: Like testing how your car handles sudden acceleration.

### When to Use Each Test Type

#### **Start Here: Baseline Testing**
- **When**: Before any other testing, after system setup
- **Duration**: 5-10 minutes
- **Goal**: Establish your "normal" performance
- **Success criteria**: Consistent, predictable performance

#### **Then: Load Testing**
- **When**: After baseline is established
- **Duration**: 10-30 minutes  
- **Goal**: Verify system handles expected production load
- **Success criteria**: Performance stays within acceptable limits

#### **Next: Stress Testing**
- **When**: After load testing passes
- **Duration**: 15-30 minutes with ramp-up
- **Goal**: Find your system's limits
- **Success criteria**: Identify breaking point without system failure

#### **Finally: Endurance Testing**
- **When**: After stress testing identifies limits
- **Duration**: 1-4 hours
- **Goal**: Verify long-term stability
- **Success criteria**: No performance degradation over time

### What to Look For (Red Flags)

#### **Performance Degradation**
- Throughput drops below 80% of baseline
- Response times increase by more than 50%
- Error rates exceed 1%

#### **Resource Issues**
- CPU usage consistently above 90%
- Memory usage growing continuously (memory leaks)
- Disk I/O saturation
- Network bandwidth exhaustion

#### **Milvus-Specific Issues**
- Insert failures increasing over time
- Search timeouts becoming frequent
- Index creation failures
- Collection loading issues
- Query performance degradation

#### **System Behavior**
- Gradual slowdown over time
- Sudden performance drops
- Inconsistent response times
- System becoming unresponsive

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

## 🤖 RAG (Retrieval-Augmented Generation) Testing

### RAG-Specific Characteristics

RAG systems have unique performance patterns:
- **High search-to-insert ratio**: More searches than inserts
- **Low-latency requirements**: Sub-second search responses
- **Batch processing**: Document ingestion in batches
- **Real-time queries**: User-facing search requests
- **Vector similarity**: Semantic search performance

### Recommended RAG Test Scenarios

#### **1. Document Ingestion Testing**
```bash
# Simulate document processing pipeline
go run main.go --duration 30m --pressure medium --real-time
```
**Purpose**: Test document embedding insertion performance
**Typical workload**: Batch processing of documents, PDFs, text files

#### **2. Real-time Query Testing**
```bash
# Simulate user search queries
go run main.go --duration 15m --pressure high --real-time
```
**Purpose**: Test search performance under user load
**Typical workload**: User queries, semantic search, similarity matching

#### **3. RAG End-to-End Testing**
```bash
# Combined insert + search workload
go run main.go --duration 1h --pressure high --ramp-up --real-time
```
**Purpose**: Test complete RAG pipeline performance
**Typical workload**: Continuous document ingestion + user queries

#### **4. RAG Stress Testing**
```bash
# Find RAG system limits
go run main.go --duration 20m --pressure extreme --ramp-up --real-time
```
**Purpose**: Discover maximum RAG throughput
**Typical workload**: Peak document processing + high query volume

### RAG-Specific Metrics to Monitor

#### **Search Performance (Critical for RAG)**
- **Search Latency**: < 100ms for real-time queries
- **Search Throughput**: Queries per second
- **Search Accuracy**: Recall@K, Precision@K
- **Index Performance**: Index creation and loading time

#### **Insert Performance (Document Processing)**
- **Insert Latency**: Batch processing time
- **Insert Throughput**: Documents processed per second
- **Embedding Generation**: Time to create vector embeddings
- **Batch Processing**: Large document batch handling

#### **RAG-Specific Red Flags**
- Search latency > 200ms
- Search failures during peak load
- Index creation taking > 5 minutes
- Memory usage growing during search operations
- Search accuracy degradation under load

### RAG Production Recommendations

#### **Development/Testing Environment**
```bash
# Light RAG testing
go run main.go --duration 10m --pressure low --real-time
```

#### **Production RAG System**
```bash
# Production RAG load testing
go run main.go --duration 1h --pressure high --real-time
```

#### **High-Traffic RAG (ChatGPT-like)**
```bash
# High-volume RAG system
go run main.go --duration 2h --pressure extreme --ramp-up --real-time
```

### RAG Performance Benchmarks

| RAG Scale | Users | Documents/Day | Queries/Day | Recommended Test |
|-----------|-------|---------------|-------------|------------------|
| **Small RAG** | < 1K | < 10K | < 100K | `--pressure low` |
| **Medium RAG** | 1K-10K | 10K-100K | 100K-1M | `--pressure medium` |
| **Large RAG** | 10K-100K | 100K-1M | 1M-10M | `--pressure high` |
| **Enterprise RAG** | 100K+ | 1M+ | 10M+ | `--pressure extreme` |

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
