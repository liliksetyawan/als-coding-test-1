package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	itemIDs := []string{"A", "B", "C", "D", "E", "F"}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, errs := FetchAndAggregate(ctx, itemIDs, 3, 2*time.Second)

	fmt.Println("Hasil:")
	for id, item := range results {
		fmt.Printf("- %s: %s | Rp%.2f\n", id, item.Name, item.Price)
	}

	fmt.Println("\nErrors:")
	for _, e := range errs {
		fmt.Println("-", e)
	}
}

// ItemDetails
type ItemDetails struct {
	ID          string
	Name        string
	Description string
	Price       float64
}

// simulateFetchItemDetails simulates an external API call to fetch item details.
// It introduces random delays and occasional errors.
// It respects the context for cancellation.
func simulateFetchItemDetails(ctx context.Context, itemID string) (*ItemDetails, error) {
	// Simulate network latency
	delay := time.Duration(500+rand.Intn(1500)) * time.Millisecond // 0.5s to 2s
	select {
	case <-time.After(delay):
		// Continue
	case <-ctx.Done():
		log.Printf("Context cancelled for item %s, aborting fetch.", itemID)
		return nil, ctx.Err() // Context cancelled
	}

	// Simulate occasional API errors
	if rand.Intn(100) < 15 { // 15% chance of error
		log.Printf("Simulated API error for item %s", itemID)
		return nil, fmt.Errorf("simulated API error for item %s: service unavailable", itemID)
	}

	details := &ItemDetails{
		ID:          itemID,
		Name:        fmt.Sprintf("Product %s", itemID),
		Description: fmt.Sprintf("Detailed description for product %s.", itemID),
		Price:       rand.Float64() * 100,
	}
	log.Printf("Successfully fetched details for item %s", itemID)
	return details, nil
}

func FetchAndAggregate(ctx context.Context, itemIDs []string, maxConcurrent int, perItemTimeout time.Duration) (results map[string]ItemDetails, errs []error) {
	results = make(map[string]ItemDetails)
	var wg sync.WaitGroup

	var mu sync.Mutex // gunakan mutext untuk locking var results and errs

	sem := make(chan struct{}, maxConcurrent)

	for _, id := range itemIDs {
		wg.Add(1)

		id := id

		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			// per item timeout
			itemCtx, cancel := context.WithTimeout(ctx, perItemTimeout)
			defer cancel()

			details, err := simulateFetchItemDetails(itemCtx, id)
			mu.Lock()

			defer mu.Unlock()

			if err != nil {
				errs = append(errs, err)
				return
			}

			results[id] = *details

		}()

	}

	wg.Wait()

	return results, errs
}
