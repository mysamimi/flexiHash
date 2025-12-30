package main

import (
	"encoding/json"
	"fmt"
	"log"

	flexihash "github.com/mysamimi/flexiHash"
)

// This example demonstrates how Go and PHP can work together
// by ensuring consistent hash results across both platforms

func main() {
	fmt.Println("=== Go & PHP Interoperability Demo ===")

	// Create FlexiHash instance with same configuration as PHP
	hash := flexihash.NewFlexiHash() // Default: 64 replicas, CRC32 hasher

	// Add targets (same as PHP example)
	targets := []string{"cache-1", "cache-2", "cache-3"}
	err := hash.AddTargets(targets, 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Target Configuration:")
	fmt.Printf("  Servers: %v\n", targets)
	fmt.Printf("  Replicas: 64 (default)\n")
	fmt.Printf("  Hasher: CRC32 (matches PHP crc32)\n\n")

	// Test objects that can be verified against PHP
	testObjects := []string{
		"object-a",
		"object-b",
		"object-c",
		"user-123",
		"session-456",
	}

	fmt.Println("Hash Distribution (verify these match PHP):")
	for _, obj := range testObjects {
		target, err := hash.Lookup(obj)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %-15s -> %s\n", obj, target)
	}

	// Demonstrate lookupList for redundancy
	fmt.Println("\nLookupList with fallbacks (2 servers):")
	for _, obj := range testObjects {
		targets, err := hash.LookupList(obj, 2)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %-15s -> %v\n", obj, targets)
	}

	// Export configuration for PHP verification
	fmt.Println("\n=== Configuration for PHP Verification ===")
	config := map[string]interface{}{
		"replicas": 64,
		"hasher":   "crc32",
		"targets":  targets,
	}
	jsonConfig, _ := json.MarshalIndent(config, "", "  ")
	fmt.Printf("Config:\n%s\n", jsonConfig)

	// Test consistency with weights
	fmt.Println("\n=== Testing Weighted Targets ===")
	hash2 := flexihash.NewFlexiHash()
	hash2.AddTarget("small-cache", 1)
	hash2.AddTarget("large-cache", 2)

	counts := make(map[string]int)
	testCount := 1000
	for i := 0; i < testCount; i++ {
		key := fmt.Sprintf("item-%d", i)
		target, _ := hash2.Lookup(key)
		counts[target]++
	}

	fmt.Printf("Distribution over %d items:\n", testCount)
	for target, count := range counts {
		percentage := float64(count) / float64(testCount) * 100
		fmt.Printf("  %-12s: %4d items (%.1f%%)\n", target, count, percentage)
	}

	// Generate test vectors for PHP
	fmt.Println("\n=== Test Vectors for PHP Verification ===")
	fmt.Println("Copy this to PHP to verify:")
	fmt.Println("<?php")
	fmt.Println("require_once 'vendor/autoload.php';")
	fmt.Println()
	fmt.Println("use Flexihash\\Flexihash;")
	fmt.Println()
	fmt.Println("$hash = new Flexihash();")
	fmt.Printf("$hash->addTargets(%s);\n", formatPHPArray(targets))
	fmt.Println()

	for _, obj := range testObjects {
		target, _ := hash.Lookup(obj)
		fmt.Printf("// Test: %s\n", obj)
		fmt.Printf("$result = $hash->lookup('%s');\n", obj)
		fmt.Printf("echo \"%-15s -> \" . $result . \" (expected: %s)\" . PHP_EOL;\n", obj, target)
		fmt.Printf("assert($result === '%s', '%s failed');\n", target, obj)
		fmt.Println()
	}
	fmt.Println("echo \"All tests passed!\" . PHP_EOL;")
	fmt.Println("?>")
}

func formatPHPArray(items []string) string {
	result := "["
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("'%s'", item)
	}
	result += "]"
	return result
}
