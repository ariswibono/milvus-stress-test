## Milvus Stress Test (Go)

A simple, self-contained high-throughput insert/search simulation against a Milvus instance using the Go SDK. It spins up concurrent goroutines to insert randomly generated vectors, creates an IVF_FLAT index, loads the collection, performs concurrent searches, and then cleans up.

### Features
- Concurrent batched inserts with configurable workers and batch sizes
- Index creation (IVF_FLAT, L2) and collection loading
- Concurrent vector searches with configurable concurrency
- Clear step-by-step logs and throughput metrics

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

### Run the Stress Test
The program connects to Milvus, drops any existing test collection, inserts data concurrently, builds an index, loads the collection, runs concurrent searches, prints metrics, and drops the collection.

Basic run (defaults to NORMAL traffic):
```bash
go run main.go --milvus-addr localhost:19530
```

Traffic presets (mutually exclusive; choose one):
- `--test` (NORMAL)
- `--test-medium`
- `--test-high`
- `--test-very-high`

Examples:
```bash
# NORMAL (10 workers × 10 batches × 1000 vectors)
go run main.go --milvus-addr localhost:19530 --test

# MEDIUM (20 workers × 20 batches × 2000 vectors)
go run main.go --milvus-addr localhost:19530 --test-medium

# HIGH (50 workers × 20 batches × 5000 vectors)
go run main.go --milvus-addr localhost:19530 --test-high

# VERY HIGH (100 workers × 20 batches × 10000 vectors)
go run main.go --milvus-addr localhost:19530 --test-very-high
```

### Flags
- `--milvus-addr string`: Milvus gRPC address (default `localhost:19530`).
- `--test`: Run with NORMAL traffic.
- `--test-medium`: Run with MEDIUM traffic.
- `--test-high`: Run with HIGH traffic.
- `--test-very-high`: Run with VERY HIGH traffic.

If no traffic flag is provided, the program runs with NORMAL defaults and logs as `NORMAL (default)`.

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

### Sample Output (abridged)
```text
--- Step 4: Starting concurrent data insertion ---
✅ All goroutines finished inserting data in 512.877458ms.
   -> Throughput: 1559826.79 inserts/second
Flushing collection to seal segments...
✅ Data flushed successfully.
--- Step 5: Create index on field 'embedding' ---
✅ Index created successfully in 1.244369542s.
--- Step 6: Load collection into memory ---
✅ Collection loaded successfully in 3.613254333s.
--- Step 7: Perform 20 concurrent searches ---
✅ All search goroutines finished in 50.902042ms.
   -> Throughput: 392.91 searches/second
--- Simulation Summary ---
Total Vectors Inserted: 800000
```

Note: Actual timings depend on your machine and Milvus configuration.

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
