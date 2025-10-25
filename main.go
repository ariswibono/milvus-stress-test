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

func main() {
	// --- Command-line flags for traffic levels and connection ---
	milvusAddr := flag.String("milvus-addr", "localhost:19530", "The address of the Milvus instance.")
	testNormal := flag.Bool("test", false, "Run with normal traffic settings.")
	testMedium := flag.Bool("test-medium", false, "Run with medium traffic settings.")
	testHigh := flag.Bool("test-high", false, "Run with high traffic settings.")
	testVeryHigh := flag.Bool("test-very-high", false, "Run with very high traffic settings.")
	flag.Parse()

	// --- Simulation Settings ---
	var numConcurrentGoroutines, batchSize, numBatchesPerGoroutine int
	var trafficLevel string

	if *testMedium {
		trafficLevel = "MEDIUM"
		numConcurrentGoroutines = 20
		batchSize = 2000
		numBatchesPerGoroutine = 20
	} else if *testHigh {
		trafficLevel = "HIGH"
		numConcurrentGoroutines = 50
		batchSize = 5000
		numBatchesPerGoroutine = 20
	} else if *testVeryHigh {
		trafficLevel = "VERY HIGH"
		numConcurrentGoroutines = 100
		batchSize = 10000
		numBatchesPerGoroutine = 20
	} else { // Default to normal traffic if no other flag is set
		// This block now handles both the explicit `--test` flag and the default case
		if *testNormal {
			trafficLevel = "NORMAL"
		} else {
			trafficLevel = "NORMAL (default)"
		}
		numConcurrentGoroutines = 10
		batchSize = 1000
		numBatchesPerGoroutine = 10
	}

	// --- Simulation Strategy Details ---
	totalInserts := numConcurrentGoroutines * numBatchesPerGoroutine * batchSize
	fmt.Printf(">> Running with %s traffic settings <<\n", trafficLevel)
	fmt.Println("\n--- Simulation Strategy ---")
	fmt.Printf(" - Milvus Address:                  %s\n", *milvusAddr)
	fmt.Printf(" - Concurrent Workers (Goroutines): %d\n", numConcurrentGoroutines)
	fmt.Printf(" - Batch Size (Vectors per Insert): %d\n", batchSize)
	fmt.Printf(" - Batches per Worker:              %d\n", numBatchesPerGoroutine)
	fmt.Printf(" ---------------------------------------\n")
	fmt.Printf(" - Total Vectors to Insert:         %d\n", totalInserts)
	fmt.Println("---------------------------")

	totalStartTime := time.Now()
	ctx := context.Background()

	// Variables to track metrics for summary table
	var (
		connectionTime time.Duration
		insertionTime  time.Duration
		flushTime      time.Duration
		indexTime      time.Duration
		loadTime       time.Duration
		searchTime     time.Duration
		cleanupTime    time.Duration
		insertsPerSec  float64
		searchesPerSec float64
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

	// 4. Insert data concurrently
	fmt.Printf("\n--- Step 4: Starting concurrent data insertion ---\n")

	var wg sync.WaitGroup
	insertionStartTime := time.Now()

	for i := 0; i < numConcurrentGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			fmt.Printf("[Goroutine %d] Starting to insert %d batches of %d vectors...\n", goroutineID, numBatchesPerGoroutine, batchSize)
			rand.Seed(time.Now().UnixNano() + int64(goroutineID))

			for j := 0; j < numBatchesPerGoroutine; j++ {
				vectors := make([][]float32, batchSize)
				for k := 0; k < batchSize; k++ {
					vec := make([]float32, embeddingDim)
					for l := 0; l < embeddingDim; l++ {
						vec[l] = rand.Float32()
					}
					vectors[k] = vec
				}
				embeddingColumn := entity.NewColumnFloatVector(embeddingField, embeddingDim, vectors)
				_, err := milvusClient.Insert(ctx, collectionName, "", embeddingColumn)
				if err != nil {
					log.Printf("[Goroutine %d] Failed to insert batch %d: %v", goroutineID, j, err)
					return
				}
			}
			fmt.Printf("[Goroutine %d] Finished inserting all its batches.\n", goroutineID)
		}(i)
	}

	wg.Wait()
	insertionTime = time.Since(insertionStartTime)
	insertsPerSec = float64(totalInserts) / insertionTime.Seconds()

	fmt.Printf("✅ All goroutines finished inserting data in %s.\n", insertionTime)
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

	// 7. Perform concurrent searches
	fmt.Printf("\n--- Step 7: Perform %d concurrent searches ---\n", numConcurrentGoroutines)

	var searchWg sync.WaitGroup
	searchStartTime := time.Now()

	for i := 0; i < numConcurrentGoroutines; i++ {
		searchWg.Add(1)
		go func(goroutineID int) {
			defer searchWg.Done()
			fmt.Printf("[Search Goroutine %d] Executing search...\n", goroutineID)
			queryVectorData := make([]float32, embeddingDim)
			for j := range queryVectorData {
				queryVectorData[j] = rand.Float32()
			}
			queryVector := []entity.Vector{entity.FloatVector(queryVectorData)}
			searchParams, _ := entity.NewIndexIvfFlatSearchParam(10) // nprobe = 10

			_, err := milvusClient.Search(ctx, collectionName, []string{}, "", []string{}, queryVector, embeddingField, entity.L2, 3, searchParams)
			if err != nil {
				log.Printf("[Search Goroutine %d] Failed to perform search: %v", goroutineID, err)
			}
			fmt.Printf("[Search Goroutine %d] Search completed.\n", goroutineID)
		}(i)
	}
	searchWg.Wait()
	searchTime = time.Since(searchStartTime)
	searchesPerSec = float64(numConcurrentGoroutines) / searchTime.Seconds()

	fmt.Printf("✅ All search goroutines finished in %s.\n", searchTime)
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
	totalVectors := totalInserts
	totalDataMB := float64(totalVectors*embeddingDim*4) / (1024 * 1024) // 4 bytes per float32

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                           PERFORMANCE SUMMARY TABLE")
	fmt.Println(strings.Repeat("=", 80))

	// Configuration section
	fmt.Printf("│ %-25s │ %-50s │\n", "Configuration", "Value")
	fmt.Println("├" + strings.Repeat("─", 27) + "┼" + strings.Repeat("─", 52) + "┤")
	fmt.Printf("│ %-25s │ %-50s │\n", "Traffic Level", trafficLevel)
	fmt.Printf("│ %-25s │ %-50s │\n", "Milvus Address", *milvusAddr)
	fmt.Printf("│ %-25s │ %-50d │\n", "Concurrent Workers", numConcurrentGoroutines)
	fmt.Printf("│ %-25s │ %-50d │\n", "Batch Size", batchSize)
	fmt.Printf("│ %-25s │ %-50d │\n", "Batches per Worker", numBatchesPerGoroutine)
	fmt.Printf("│ %-25s │ %-50d │\n", "Total Vectors", totalVectors)
	fmt.Printf("│ %-25s │ %-50.2f MB │\n", "Total Data Size", totalDataMB)

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
