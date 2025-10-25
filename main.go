package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	// Collection settings
	collectionName  = "go_high_throughput_collection"
	embeddingDim    = 8
	primaryKeyField = "id"
	embeddingField  = "embedding"
)

// calculateDynamicLoad calculates the current load based on elapsed time (like a real dyno)
func calculateDynamicLoad(elapsed time.Duration, totalDuration time.Duration, maxWorkers int, maxBatchSize int) (int, int) {
	if elapsed >= totalDuration {
		return maxWorkers, maxBatchSize
	}

	// Linear ramp-up: start at 10% of max, reach 100% at the end
	progress := float64(elapsed) / float64(totalDuration)
	if progress > 1.0 {
		progress = 1.0
	}

	// Start at 10% of max capacity
	startWorkers := max(1, maxWorkers/10)
	startBatchSize := max(100, maxBatchSize/10)

	currentWorkers := startWorkers + int(float64(maxWorkers-startWorkers)*progress)
	currentBatchSize := startBatchSize + int(float64(maxBatchSize-startBatchSize)*progress)

	return currentWorkers, currentBatchSize
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func showDetailedHelp() {
	fmt.Println("Milvus Load Testing Tool")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  go run main.go [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  --milvus-addr string")
	fmt.Println("        Milvus server address (default: localhost:19530)")
	fmt.Println("        Example: --milvus-addr 192.168.1.100:19530")
	fmt.Println()
	fmt.Println("  --duration duration")
	fmt.Println("        Test duration (default: 30s)")
	fmt.Println("        Examples: --duration 30s, --duration 2m, --duration 1h")
	fmt.Println()
	fmt.Println("  --pressure string")
	fmt.Println("        Load intensity level (default: medium)")
	fmt.Println("        Options: low, medium, high, extreme")
	fmt.Println("        - low:    5 workers, 500 vectors/batch")
	fmt.Println("        - medium: 20 workers, 2000 vectors/batch")
	fmt.Println("        - high:   50 workers, 5000 vectors/batch")
	fmt.Println("        - extreme: 100 workers, 10000 vectors/batch")
	fmt.Println()
	fmt.Println("  --ramp-up")
	fmt.Println("        Gradually increase load from 10% to 100% over duration")
	fmt.Println("        Useful for finding performance limits")
	fmt.Println()
	fmt.Println("  --real-time")
	fmt.Println("        Display real-time throughput metrics during test")
	fmt.Println()
	fmt.Println("  --help")
	fmt.Println("        Show this help information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic 30-second medium load test")
	fmt.Println("  go run main.go")
	fmt.Println()
	fmt.Println("  # High load test for 2 minutes")
	fmt.Println("  go run main.go --duration 2m --pressure high")
	fmt.Println()
	fmt.Println("  # Gradual load increase with real-time monitoring")
	fmt.Println("  go run main.go --duration 1m --pressure high --ramp-up --real-time")
	fmt.Println()
	fmt.Println("  # Extreme endurance test")
	fmt.Println("  go run main.go --duration 1h --pressure extreme --real-time")
	fmt.Println()
	fmt.Println("  # Custom Milvus server")
	fmt.Println("  go run main.go --milvus-addr 192.168.1.100:19530 --duration 5m")
}

func main() {
	// --- Command-line flags for load testing ---
	milvusAddr := flag.String("milvus-addr", "localhost:19530", "Milvus server address (host:port)")
	duration := flag.Duration("duration", 30*time.Second, "Test duration (e.g., 30s, 2m, 1h)")
	pressure := flag.String("pressure", "medium", "Load intensity: low, medium, high, extreme")
	rampUp := flag.Bool("ramp-up", false, "Gradually increase load from 10% to 100% over duration")
	realTime := flag.Bool("real-time", false, "Display real-time throughput metrics")
	showHelp := flag.Bool("help", false, "Show detailed help information")
	flag.Parse()

	// Show help if requested
	if *showHelp {
		showDetailedHelp()
		return
	}

	// --- Pressure Level Settings ---
	var numConcurrentGoroutines, batchSize int
	var pressureLevel string

	switch *pressure {
	case "low":
		pressureLevel = "LOW"
		numConcurrentGoroutines = 5
		batchSize = 500
	case "medium":
		pressureLevel = "MEDIUM"
		numConcurrentGoroutines = 20
		batchSize = 2000
	case "high":
		pressureLevel = "HIGH"
		numConcurrentGoroutines = 50
		batchSize = 5000
	case "extreme":
		pressureLevel = "EXTREME"
		numConcurrentGoroutines = 100
		batchSize = 10000
	default:
		pressureLevel = "MEDIUM (default)"
		numConcurrentGoroutines = 20
		batchSize = 2000
	}

	// --- Load Test Configuration ---
	fmt.Printf(">> Starting Milvus Load Test: %s intensity for %s <<\n", pressureLevel, *duration)
	fmt.Println("\n--- Test Configuration ---")
	fmt.Printf(" - Milvus Address:                  %s\n", *milvusAddr)
	fmt.Printf(" - Test Duration:                   %s\n", *duration)
	fmt.Printf(" - Load Intensity:                  %s\n", pressureLevel)
	fmt.Printf(" - Concurrent Workers:              %d\n", numConcurrentGoroutines)
	fmt.Printf(" - Batch Size (Vectors per Insert): %d\n", batchSize)
	fmt.Printf(" - Test Mode:                       Continuous load until duration expires\n")
	fmt.Println("----------------------------------------")

	totalStartTime := time.Now()
	ctx := context.Background()

	// Variables to track metrics for summary table
	var (
		connectionTime         time.Duration
		insertionTime          time.Duration
		flushTime              time.Duration
		indexTime              time.Duration
		loadTime               time.Duration
		searchTime             time.Duration
		cleanupTime            time.Duration
		insertsPerSec          float64
		searchesPerSec         float64
		totalVectorsInserted   int64
		totalSearchesPerformed int64
	)

	// 1. Connect to Milvus
	fmt.Println("\n--- Step 1: Connect to Milvus ---")
	fmt.Printf("Attempting to connect to Milvus at %s...\n", *milvusAddr)
	connectStart := time.Now()
	milvusClient, err := client.NewClient(ctx, client.Config{Address: *milvusAddr})
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}
	defer milvusClient.Close()
	connectionTime = time.Since(connectStart)
	fmt.Println("✅ Connected to Milvus successfully!")

	// 2. Clean up previous runs
	fmt.Printf("\n--- Step 2: Check for and drop existing collection '%s' ---\n", collectionName)
	has, err := milvusClient.HasCollection(ctx, collectionName)
	if err != nil {
		log.Fatalf("Failed to check if collection exists: %v", err)
	}
	if has {
		fmt.Printf("Collection '%s' already exists. Dropping it...\n", collectionName)
		if err := milvusClient.DropCollection(ctx, collectionName); err != nil {
			log.Fatalf("Failed to drop collection: %v", err)
		}
		fmt.Println("✅ Dropped existing collection.")
	} else {
		fmt.Println("Collection does not exist, proceeding.")
	}

	// 3. Create a collection
	fmt.Printf("\n--- Step 3: Create collection '%s' ---\n", collectionName)
	schema := &entity.Schema{
		CollectionName: collectionName,
		Fields: []*entity.Field{
			{Name: primaryKeyField, DataType: entity.FieldTypeInt64, PrimaryKey: true, AutoID: true},
			{Name: embeddingField, DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": fmt.Sprintf("%d", embeddingDim)}},
		},
	}
	if err := milvusClient.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		log.Fatalf("Failed to create collection: %v", err)
	}
	fmt.Println("✅ Collection created successfully.")

	// 4. Insert data continuously for the specified duration (with optional ramp-up)
	fmt.Printf("\n--- Step 4: Starting continuous data insertion for %s ---\n", *duration)
	if *rampUp {
		fmt.Println("📈 RAMP-UP MODE: Gradually increasing load from 10% to 100%...")
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	insertionStartTime := time.Now()
	testEndTime := insertionStartTime.Add(*duration)

	// Start all worker goroutines
	for i := 0; i < numConcurrentGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			fmt.Printf("[Worker %d] Starting continuous insertion...\n", goroutineID)
			rand.Seed(time.Now().UnixNano() + int64(goroutineID))

			batchCount := 0
			lastThroughput := 0.0

			for time.Now().Before(testEndTime) {
				// Calculate dynamic load if ramp-up is enabled
				currentBatchSize := batchSize
				if *rampUp {
					elapsed := time.Since(insertionStartTime)
					_, currentBatchSize = calculateDynamicLoad(elapsed, *duration, numConcurrentGoroutines, batchSize)
				}

				vectors := make([][]float32, currentBatchSize)
				for k := 0; k < currentBatchSize; k++ {
					vec := make([]float32, embeddingDim)
					for l := 0; l < embeddingDim; l++ {
						vec[l] = rand.Float32()
					}
					vectors[k] = vec
				}
				embeddingColumn := entity.NewColumnFloatVector(embeddingField, embeddingDim, vectors)
				_, err := milvusClient.Insert(ctx, collectionName, "", embeddingColumn)
				if err != nil {
					log.Printf("[Worker %d] Failed to insert batch %d: %v", goroutineID, batchCount, err)
					continue
				}

				// Update counters atomically
				mu.Lock()
				totalVectorsInserted += int64(currentBatchSize)
				mu.Unlock()

				// Real-time monitoring
				if *realTime && batchCount%10 == 0 {
					elapsed := time.Since(insertionStartTime)
					currentThroughput := float64(totalVectorsInserted) / elapsed.Seconds()
					if currentThroughput != lastThroughput {
						fmt.Printf("📊 [%s] Batch Size: %d, Throughput: %.1f ops/sec\n",
							elapsed.Round(time.Second), currentBatchSize, currentThroughput)
						lastThroughput = currentThroughput
					}
				}

				batchCount++
			}
			fmt.Printf("[Worker %d] Finished after %d batches.\n", goroutineID, batchCount)
		}(i)
	}

	wg.Wait()
	insertionTime = time.Since(insertionStartTime)
	insertsPerSec = float64(totalVectorsInserted) / insertionTime.Seconds()

	fmt.Printf("✅ All workers finished inserting data in %s.\n", insertionTime)
	fmt.Printf("   -> Total vectors inserted: %d\n", totalVectorsInserted)
	fmt.Printf("   -> Throughput: %.2f inserts/second\n", insertsPerSec)

	// Flush the collection
	fmt.Println("\nFlushing collection to seal segments...")
	flushStart := time.Now()
	if err := milvusClient.Flush(ctx, collectionName, false); err != nil {
		log.Fatalf("Failed to flush collection: %v", err)
	}
	flushTime = time.Since(flushStart)
	fmt.Println("✅ Data flushed successfully.")

	// 5. Create an index
	fmt.Printf("\n--- Step 5: Create index on field '%s' ---\n", embeddingField)
	index, _ := entity.NewIndexIvfFlat(entity.L2, 16)
	fmt.Println("Waiting for index to be built (this may take a while)...")
	indexStartTime := time.Now()
	if err := milvusClient.CreateIndex(ctx, collectionName, embeddingField, index, false); err != nil {
		log.Fatalf("Failed to create index: %v", err)
	}
	indexTime = time.Since(indexStartTime)
	fmt.Printf("✅ Index created successfully in %s.\n", indexTime)

	// 6. Load the collection
	fmt.Println("\n--- Step 6: Load collection into memory ---")
	loadStartTime := time.Now()
	if err := milvusClient.LoadCollection(ctx, collectionName, false); err != nil {
		log.Fatalf("Failed to load collection: %v", err)
	}
	loadTime = time.Since(loadStartTime)
	fmt.Printf("✅ Collection loaded successfully in %s.\n", loadTime)

	// 7. Perform continuous searches for a shorter duration
	searchDuration := *duration / 4 // Search for 1/4 of the total test duration
	fmt.Printf("\n--- Step 7: Perform continuous searches for %s ---\n", searchDuration)

	var searchWg sync.WaitGroup
	var searchMu sync.Mutex
	searchStartTime := time.Now()
	searchEndTime := searchStartTime.Add(searchDuration)

	for i := 0; i < numConcurrentGoroutines; i++ {
		searchWg.Add(1)
		go func(goroutineID int) {
			defer searchWg.Done()
			fmt.Printf("[Search Worker %d] Starting continuous searches...\n", goroutineID)
			rand.Seed(time.Now().UnixNano() + int64(goroutineID))

			searchCount := 0
			for time.Now().Before(searchEndTime) {
				queryVectorData := make([]float32, embeddingDim)
				for j := range queryVectorData {
					queryVectorData[j] = rand.Float32()
				}
				queryVector := []entity.Vector{entity.FloatVector(queryVectorData)}
				searchParams, _ := entity.NewIndexIvfFlatSearchParam(10) // nprobe = 10

				_, err := milvusClient.Search(ctx, collectionName, []string{}, "", []string{}, queryVector, embeddingField, entity.L2, 3, searchParams)
				if err != nil {
					log.Printf("[Search Worker %d] Failed to perform search %d: %v", goroutineID, searchCount, err)
					continue
				}

				// Update counters atomically
				searchMu.Lock()
				totalSearchesPerformed++
				searchMu.Unlock()

				searchCount++
			}
			fmt.Printf("[Search Worker %d] Finished after %d searches.\n", goroutineID, searchCount)
		}(i)
	}
	searchWg.Wait()
	searchTime = time.Since(searchStartTime)
	searchesPerSec = float64(totalSearchesPerformed) / searchTime.Seconds()

	fmt.Printf("✅ All search workers finished in %s.\n", searchTime)
	fmt.Printf("   -> Total searches performed: %d\n", totalSearchesPerformed)
	fmt.Printf("   -> Throughput: %.2f searches/second\n", searchesPerSec)

	// 8. Clean up
	fmt.Printf("\n--- Step 8: Clean up by dropping collection '%s' ---\n", collectionName)
	cleanupStart := time.Now()
	if err := milvusClient.DropCollection(ctx, collectionName); err != nil {
		log.Fatalf("Failed to drop collection: %v", err)
	}
	cleanupTime = time.Since(cleanupStart)
	fmt.Println("✅ Cleanup successful!")

	// --- Final Summary Table ---
	totalDuration := time.Since(totalStartTime)
	totalDataMB := float64(totalVectorsInserted*embeddingDim*4) / (1024 * 1024) // 4 bytes per float32

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                        LOAD TEST PERFORMANCE SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	// Configuration section
	fmt.Printf("│ %-25s │ %-50s │\n", "Configuration", "Value")
	fmt.Println("├" + strings.Repeat("─", 27) + "┼" + strings.Repeat("─", 52) + "┤")
	fmt.Printf("│ %-25s │ %-50s │\n", "Test Duration", *duration)
	fmt.Printf("│ %-25s │ %-50s │\n", "Pressure Level", pressureLevel)
	fmt.Printf("│ %-25s │ %-50s │\n", "Milvus Address", *milvusAddr)
	fmt.Printf("│ %-25s │ %-50d │\n", "Concurrent Workers", numConcurrentGoroutines)
	fmt.Printf("│ %-25s │ %-50d │\n", "Batch Size", batchSize)
	fmt.Printf("│ %-25s │ %-50d │\n", "Vectors Inserted", totalVectorsInserted)
	fmt.Printf("│ %-25s │ %-50.2f MB │\n", "Data Size Inserted", totalDataMB)
	fmt.Printf("│ %-25s │ %-50d │\n", "Searches Performed", totalSearchesPerformed)

	fmt.Println("├" + strings.Repeat("─", 27) + "┼" + strings.Repeat("─", 52) + "┤")

	// Performance metrics section
	fmt.Printf("│ %-25s │ %-50s │\n", "Performance Metrics", "Value")
	fmt.Println("├" + strings.Repeat("─", 27) + "┼" + strings.Repeat("─", 52) + "┤")
	fmt.Printf("│ %-25s │ %-50s │\n", "Total Elapsed Time", totalDuration.String())
	fmt.Printf("│ %-25s │ %-50s │\n", "Connection Time", connectionTime.String())
	fmt.Printf("│ %-25s │ %-50s │\n", "Data Insertion Time", insertionTime.String())
	fmt.Printf("│ %-25s │ %-50.2f │\n", "Insert Throughput", insertsPerSec)
	fmt.Printf("│ %-25s │ %-50s │\n", "Flush Time", flushTime.String())
	fmt.Printf("│ %-25s │ %-50s │\n", "Index Creation Time", indexTime.String())
	fmt.Printf("│ %-25s │ %-50s │\n", "Collection Load Time", loadTime.String())
	fmt.Printf("│ %-25s │ %-50s │\n", "Search Execution Time", searchTime.String())
	fmt.Printf("│ %-25s │ %-50.2f │\n", "Search Throughput", searchesPerSec)
	fmt.Printf("│ %-25s │ %-50s │\n", "Cleanup Time", cleanupTime.String())

	fmt.Println(strings.Repeat("=", 80))
}
