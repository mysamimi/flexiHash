package main

import (
	"fmt"
	"log"

	flexihash "github.com/mysamimi/flexiHash"
)

func main() {
	// Create a new FlexiHash instance
	hash := flexihash.NewFlexiHash()

	// Add cache servers (targets)
	targets := []string{"cache-1", "cache-2", "cache-3"}
	err := hash.AddTargets(targets, 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Basic Usage ===")
	// Simple lookup
	target, err := hash.Lookup("object-a")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("object-a -> %s\n", target)

	target, err = hash.Lookup("object-b")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("object-b -> %s\n", target)

	// Lookup with fallback (for redundant writes)
	fmt.Println("\n=== Lookup with fallback ===")
	targets2, err := hash.LookupList("object-c", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("object-c -> %v (primary and fallback)\n", targets2)

	// Add and remove targets
	fmt.Println("\n=== Adding and removing targets ===")
	err = hash.AddTarget("cache-4", 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Added cache-4")

	err = hash.RemoveTarget("cache-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Removed cache-1")

	target, _ = hash.Lookup("object-a")
	fmt.Printf("object-a now maps to -> %s\n", target)

	// Get all targets
	fmt.Println("\n=== All targets ===")
	allTargets := hash.GetAllTargets()
	fmt.Printf("Current targets: %v\n", allTargets)

	// Weighted targets
	fmt.Println("\n=== Weighted targets ===")
	hash2 := flexihash.NewFlexiHash()
	hash2.AddTarget("light-server", 1)
	hash2.AddTarget("heavy-server", 2) // 2x capacity

	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		target, _ := hash2.Lookup(fmt.Sprintf("object-%d", i))
		counts[target]++
	}
	fmt.Printf("Distribution with weights (1000 objects):\n")
	for target, count := range counts {
		fmt.Printf("  %s: %d (%.1f%%)\n", target, count, float64(count)/10)
	}

	// Custom replicas
	fmt.Println("\n=== Custom replicas ===")
	hash3 := flexihash.NewFlexiHashWithHasher(nil, 128) // 128 replicas instead of 64
	hash3.AddTargets([]string{"server-1", "server-2", "server-3"}, 1)
	target, _ = hash3.Lookup("test-object")
	fmt.Printf("With 128 replicas: test-object -> %s\n", target)

	// Demonstrating consistency
	fmt.Println("\n=== Consistency demonstration ===")
	hash4 := flexihash.NewFlexiHash()
	hash4.AddTargets([]string{"s1", "s2", "s3", "s4", "s5"}, 1)

	// Initial lookups
	initialMappings := make(map[string]string)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		target, _ := hash4.Lookup(key)
		initialMappings[key] = target
	}

	// Remove one server
	hash4.RemoveTarget("s3")

	// Check how many keys moved
	moved := 0
	for key, initialTarget := range initialMappings {
		newTarget, _ := hash4.Lookup(key)
		if initialTarget != newTarget {
			moved++
		}
	}

	fmt.Printf("After removing 1 of 5 servers:\n")
	fmt.Printf("  Keys moved: %d/100 (%.0f%%)\n", moved, float64(moved))
	fmt.Printf("  Expected: ~20%% (only keys on removed server should move)\n")
}
