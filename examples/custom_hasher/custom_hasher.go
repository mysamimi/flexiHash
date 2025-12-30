package main

import (
	"fmt"
	"log"
	"math"

	flexihash "github.com/mysamimi/flexiHash"
)

// CustomHasher demonstrates implementing a custom hash function
type CustomHasher struct{}

func (h *CustomHasher) Hash(str string) int {
	// Simple custom hash (not recommended for production!)
	hash := 0
	for _, c := range str {
		hash = hash*31 + int(c)
	}
	return hash
}

// MockHasher for predictable testing
type MockHasher struct {
	hashValue int
}

func (m *MockHasher) Hash(str string) int {
	return m.hashValue
}

func main() {
	fmt.Println("=== Custom Hasher Example ===")

	// Using default CRC32 hasher
	fmt.Println("1. Default CRC32 Hasher:")
	hash1 := flexihash.NewFlexiHash()
	if err := hash1.AddTargets([]string{"server-1", "server-2", "server-3"}, 1); err != nil {
		log.Fatal(err)
	}

	target1, _ := hash1.Lookup("test-key")
	fmt.Printf("   test-key -> %s\n", target1)

	// Using MD5 hasher (built-in)
	fmt.Println("\n2. MD5 Hasher (built-in):")
	md5Hasher := &flexihash.Md5Hasher{}
	hash2 := flexihash.NewFlexiHashWithHasher(md5Hasher, 64)
	if err := hash2.AddTargets([]string{"server-1", "server-2", "server-3"}, 1); err != nil {
		log.Fatal(err)
	}

	target2, _ := hash2.Lookup("test-key")
	fmt.Printf("   test-key -> %s\n", target2)
	fmt.Printf("   (May differ from CRC32 due to different hash function)\n")

	// Using custom hasher
	fmt.Println("\n3. Custom Hasher:")
	customHasher := &CustomHasher{}
	hash3 := flexihash.NewFlexiHashWithHasher(customHasher, 64)
	if err := hash3.AddTargets([]string{"server-1", "server-2", "server-3"}, 1); err != nil {
		log.Fatal(err)
	}

	target3, _ := hash3.Lookup("test-key")
	fmt.Printf("   test-key -> %s\n", target3)

	// Custom replicas
	fmt.Println("\n4. Custom Replicas (higher = better distribution):")

	replicas := []int{16, 64, 256}
	for _, r := range replicas {
		hash := flexihash.NewFlexiHashWithHasher(nil, r)
		if err := hash.AddTargets([]string{"s1", "s2", "s3"}, 1); err != nil {
			log.Fatal(err)
		}

		counts := make(map[string]int)
		for i := 0; i < 1000; i++ {
			target, _ := hash.Lookup(fmt.Sprintf("key-%d", i))
			counts[target]++
		}

		fmt.Printf("   Replicas=%3d: ", r)
		for _, count := range counts {
			fmt.Printf("%d ", count)
		}

		// Calculate standard deviation to show distribution quality
		mean := 1000.0 / 3.0
		variance := 0.0
		for _, count := range counts {
			diff := float64(count) - mean
			variance += diff * diff
		}
		stddev := math.Sqrt(variance / 3.0)
		fmt.Printf("(stddev: %.1f)\n", stddev)
	}

	// MockHasher for predictable testing
	fmt.Println("\n5. Mock Hasher for Testing:")
	mockHasher := &MockHasher{hashValue: 100}
	hash4 := flexihash.NewFlexiHashWithHasher(mockHasher, 1)

	mockHasher.hashValue = 10
	if err := hash4.AddTarget("target-a", 1); err != nil {
		log.Fatal(err)
	}

	mockHasher.hashValue = 20
	if err := hash4.AddTarget("target-b", 1); err != nil {
		log.Fatal(err)
	}

	mockHasher.hashValue = 30
	if err := hash4.AddTarget("target-c", 1); err != nil {
		log.Fatal(err)
	}

	mockHasher.hashValue = 15 // Between target-a (10) and target-b (20)
	target, _ := hash4.Lookup("test")
	fmt.Printf("   Hash=15 maps to: %s (expected: target-b)\n", target)

	targets, _ := hash4.LookupList("test", 3)
	fmt.Printf("   Fallback order: %v\n", targets)

	// Compare hash distributions
	fmt.Println("\n6. Comparing Hash Algorithms:")
	testKey := "sample-key-123"

	crc32Hash := &flexihash.Crc32Hasher{}
	md5Hash := &flexihash.Md5Hasher{}
	customHash := &CustomHasher{}

	fmt.Printf("   Key: '%s'\n", testKey)
	fmt.Printf("   CRC32:  %d\n", crc32Hash.Hash(testKey))
	fmt.Printf("   MD5:    %d\n", md5Hash.Hash(testKey))
	fmt.Printf("   Custom: %d\n", customHash.Hash(testKey))

	// Demonstrate why custom hashers might be useful
	fmt.Println("\n7. Use Cases for Custom Hashers:")
	fmt.Println("   - Testing: MockHasher for predictable tests")
	fmt.Println("   - Performance: Faster hash for specific data patterns")
	fmt.Println("   - Compatibility: Match hash algorithm from other systems")
	fmt.Println("   - Security: Use cryptographic hash if needed")
	fmt.Println("   - PHP Compatibility: MD5Hasher matches PHP Flexihash MD5")
}
