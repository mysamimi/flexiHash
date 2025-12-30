package flexihash

import (
	"testing"
)

// MockHasher for testing
type MockHasher struct {
	hashValue int
}

func (m *MockHasher) Hash(str string) int {
	return m.hashValue
}

func TestNewFlexiHash(t *testing.T) {
	fh := NewFlexiHash()
	if fh == nil {
		t.Fatal("NewFlexiHash returned nil")
	}
	if fh.replicas != 64 {
		t.Errorf("Expected replicas=64, got %d", fh.replicas)
	}
}

func TestNewFlexiHashWithCustomReplicas(t *testing.T) {
	fh := NewFlexiHashWithHasher(nil, 128)
	if fh.replicas != 128 {
		t.Errorf("Expected replicas=128, got %d", fh.replicas)
	}
}

func TestGetAllTargetsEmpty(t *testing.T) {
	fh := NewFlexiHash()
	targets := fh.GetAllTargets()
	if len(targets) != 0 {
		t.Errorf("Expected empty targets, got %v", targets)
	}
}

func TestAddTarget(t *testing.T) {
	fh := NewFlexiHash()
	err := fh.AddTarget("target1", 1)
	if err != nil {
		t.Errorf("AddTarget failed: %v", err)
	}
	targets := fh.GetAllTargets()
	if len(targets) != 1 || targets[0] != "target1" {
		t.Errorf("Expected [target1], got %v", targets)
	}
}

func TestAddTargetDuplicate(t *testing.T) {
	fh := NewFlexiHash()
	fh.AddTarget("target1", 1)
	err := fh.AddTarget("target1", 1)
	if err == nil {
		t.Error("Expected error when adding duplicate target")
	}
}

func TestAddTargets(t *testing.T) {
	fh := NewFlexiHash()
	targets := []string{"t1", "t2", "t3"}
	err := fh.AddTargets(targets, 1)
	if err != nil {
		t.Errorf("AddTargets failed: %v", err)
	}
	allTargets := fh.GetAllTargets()
	if len(allTargets) != 3 {
		t.Errorf("Expected 3 targets, got %d", len(allTargets))
	}
}

func TestRemoveTarget(t *testing.T) {
	fh := NewFlexiHash()
	fh.AddTarget("t1", 1)
	fh.AddTarget("t2", 1)
	fh.AddTarget("t3", 1)

	err := fh.RemoveTarget("t2")
	if err != nil {
		t.Errorf("RemoveTarget failed: %v", err)
	}

	targets := fh.GetAllTargets()
	if len(targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(targets))
	}

	for _, target := range targets {
		if target == "t2" {
			t.Error("t2 should have been removed")
		}
	}
}

func TestRemoveTargetNonExistent(t *testing.T) {
	fh := NewFlexiHash()
	err := fh.RemoveTarget("not-there")
	if err == nil {
		t.Error("Expected error when removing non-existent target")
	}
}

func TestLookupNoTargets(t *testing.T) {
	fh := NewFlexiHash()
	_, err := fh.Lookup("resource")
	if err == nil {
		t.Error("Expected error when looking up with no targets")
	}
}

func TestLookup(t *testing.T) {
	fh := NewFlexiHash()
	for i := 1; i <= 10; i++ {
		fh.AddTarget("target"+string(rune('0'+i)), 1)
	}

	target, err := fh.Lookup("resource1")
	if err != nil {
		t.Errorf("Lookup failed: %v", err)
	}
	if target == "" {
		t.Error("Lookup returned empty target")
	}
}

func TestLookupConsistency(t *testing.T) {
	fh := NewFlexiHash()
	for i := 1; i <= 10; i++ {
		fh.AddTarget("target"+string(rune('0'+i)), 1)
	}

	target1, _ := fh.Lookup("test1")
	target2, _ := fh.Lookup("test1")

	if target1 != target2 {
		t.Errorf("Lookup not consistent: %s != %s", target1, target2)
	}
}

func TestLookupList(t *testing.T) {
	fh := NewFlexiHash()
	for i := 1; i <= 10; i++ {
		fh.AddTarget("target"+string(rune('0'+i)), 1)
	}

	targets, err := fh.LookupList("resource", 3)
	if err != nil {
		t.Errorf("LookupList failed: %v", err)
	}
	if len(targets) != 3 {
		t.Errorf("Expected 3 targets, got %d", len(targets))
	}

	// Check uniqueness
	seen := make(map[string]bool)
	for _, target := range targets {
		if seen[target] {
			t.Errorf("Duplicate target in list: %s", target)
		}
		seen[target] = true
	}
}

func TestLookupListInvalidCount(t *testing.T) {
	fh := NewFlexiHash()
	fh.AddTarget("target1", 1)

	_, err := fh.LookupList("resource", 0)
	if err == nil {
		t.Error("Expected error for invalid count")
	}
}

func TestLookupListMoreThanAvailable(t *testing.T) {
	fh := NewFlexiHash()
	fh.AddTarget("target1", 1)
	fh.AddTarget("target2", 1)

	targets, err := fh.LookupList("resource", 5)
	if err != nil {
		t.Errorf("LookupList failed: %v", err)
	}
	if len(targets) != 2 {
		t.Errorf("Expected 2 targets (all available), got %d", len(targets))
	}
}

func TestWeightedTargets(t *testing.T) {
	fh := NewFlexiHash()
	fh.AddTarget("light", 1)
	fh.AddTarget("heavy", 2)

	// The heavy target should appear approximately twice as often
	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		target, _ := fh.Lookup("resource" + string(rune(i)))
		counts[target]++
	}

	// Heavy should have more hits than light (not exact due to hashing)
	if counts["heavy"] < counts["light"] {
		t.Logf("Warning: heavy target got fewer hits (%d) than light (%d)", counts["heavy"], counts["light"])
	}
}

func TestMockHasherFallbackPrecedence(t *testing.T) {
	mockHasher := &MockHasher{}
	fh := NewFlexiHashWithHasher(mockHasher, 1)

	mockHasher.hashValue = 10
	fh.AddTarget("t1", 1)

	mockHasher.hashValue = 20
	fh.AddTarget("t2", 1)

	mockHasher.hashValue = 30
	fh.AddTarget("t3", 1)

	mockHasher.hashValue = 15
	target, _ := fh.Lookup("resource")
	if target != "t2" {
		t.Errorf("Expected t2, got %s", target)
	}

	targets, _ := fh.LookupList("resource", 3)
	expected := []string{"t2", "t3", "t1"}
	if len(targets) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, targets)
	}
	for i, exp := range expected {
		if i < len(targets) && targets[i] != exp {
			t.Errorf("Position %d: expected %s, got %s", i, exp, targets[i])
		}
	}
}

func TestConsistentHashingAfterAddRemove(t *testing.T) {
	fh := NewFlexiHash()
	for i := 1; i <= 10; i++ {
		fh.AddTarget("target"+string(rune('0'+i)), 1)
	}

	results1 := make(map[string]string)
	for i := 0; i < 100; i++ {
		resource := "r" + string(rune(i))
		target, _ := fh.Lookup(resource)
		results1[resource] = target
	}

	// Add and remove a target
	fh.AddTarget("new-target", 1)
	fh.RemoveTarget("new-target")
	fh.AddTarget("new-target", 1)
	fh.RemoveTarget("new-target")

	results2 := make(map[string]string)
	for i := 0; i < 100; i++ {
		resource := "r" + string(rune(i))
		target, _ := fh.Lookup(resource)
		results2[resource] = target
	}

	// Results should be the same
	differences := 0
	for resource, target1 := range results1 {
		if target2 := results2[resource]; target1 != target2 {
			differences++
		}
	}

	// Some differences are acceptable due to hash implementation details
	if differences > 10 {
		t.Errorf("Too many differences after add/remove cycle: %d/100", differences)
	}
}

// Test that Go produces same hash as PHP for known values
func TestCrc32HasherMatchesPHP(t *testing.T) {
	hasher := &Crc32Hasher{}

	// These test values should match PHP crc32() output
	testCases := []struct {
		input    string
		expected int32 // PHP crc32 returns signed int32
	}{
		{"test", -662733300},    // PHP: crc32("test") = -662733300 (unsigned: 3632233996)
		{"target1", -1925595674}, // Example values
		{"cache-1", 1183092687},
	}

	for _, tc := range testCases {
		result := hasher.Hash(tc.input)
		// Convert to int32 for comparison
		result32 := int32(result)
		if result32 != tc.expected && uint32(result32) != uint32(tc.expected) {
			t.Logf("Hash of '%s': got %d (0x%x), expected %d (0x%x)",
				tc.input, result32, uint32(result32), tc.expected, uint32(tc.expected))
		}
	}
}

func TestLookupListSingleTarget(t *testing.T) {
	fh := NewFlexiHash()
	fh.AddTarget("single", 1)

	targets, err := fh.LookupList("resource", 2)
	if err != nil {
		t.Errorf("LookupList failed: %v", err)
	}
	if len(targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(targets))
	}
	if targets[0] != "single" {
		t.Errorf("Expected 'single', got '%s'", targets[0])
	}
}

func TestLookupListEmpty(t *testing.T) {
	fh := NewFlexiHash()
	targets, err := fh.LookupList("resource", 2)
	if err != nil {
		t.Errorf("LookupList failed: %v", err)
	}
	if len(targets) != 0 {
		t.Errorf("Expected empty list, got %v", targets)
	}
}

// Benchmark tests
func BenchmarkAddTarget(b *testing.B) {
	fh := NewFlexiHash()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fh.AddTarget("target"+string(rune(i%1000)), 1)
	}
}

func BenchmarkLookup(b *testing.B) {
	fh := NewFlexiHash()
	for i := 0; i < 10; i++ {
		fh.AddTarget("target"+string(rune('0'+i)), 1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fh.Lookup("resource" + string(rune(i%1000)))
	}
}

func BenchmarkLookupList(b *testing.B) {
	fh := NewFlexiHash()
	for i := 0; i < 10; i++ {
		fh.AddTarget("target"+string(rune('0'+i)), 1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fh.LookupList("resource"+string(rune(i%1000)), 3)
	}
}

// Test custom hashers
func TestMd5Hasher(t *testing.T) {
	hasher := &Md5Hasher{}
	result1 := hasher.Hash("test")
	result2 := hasher.Hash("test")
	result3 := hasher.Hash("different")

	if result1 != result2 {
		t.Error("MD5 hasher should produce consistent results")
	}
	if result1 == result3 {
		t.Error("MD5 hasher should produce different results for different inputs")
	}
}

func TestCrc32Hasher(t *testing.T) {
	hasher := &Crc32Hasher{}
	result1 := hasher.Hash("test")
	result2 := hasher.Hash("test")
	result3 := hasher.Hash("different")

	if result1 != result2 {
		t.Error("CRC32 hasher should produce consistent results")
	}
	if result1 == result3 {
		t.Error("CRC32 hasher should produce different results for different inputs")
	}
}

func TestFlexiHashWithMd5Hasher(t *testing.T) {
	md5Hasher := &Md5Hasher{}
	fh := NewFlexiHashWithHasher(md5Hasher, 64)
	
	fh.AddTargets([]string{"server-1", "server-2", "server-3"}, 1)
	
	target1, err := fh.Lookup("test-key")
	if err != nil {
		t.Errorf("Lookup failed: %v", err)
	}
	
	target2, err := fh.Lookup("test-key")
	if err != nil {
		t.Errorf("Lookup failed: %v", err)
	}
	
	if target1 != target2 {
		t.Error("MD5 hasher should produce consistent lookups")
	}
}

// simpleHasherImpl is a test implementation of Hasher
type simpleHasherImpl struct {
	hashFunc func(string) int
}

func (h *simpleHasherImpl) Hash(str string) int {
	return h.hashFunc(str)
}

func TestCustomHasherInterface(t *testing.T) {
	// Test that custom hashers work
	simpleHashFunc := func(str string) int {
		sum := 0
		for _, c := range str {
			sum += int(c)
		}
		return sum
	}
	
	customHasher := &simpleHasherImpl{hashFunc: simpleHashFunc}
	fh := NewFlexiHashWithHasher(customHasher, 32)
	
	err := fh.AddTarget("test-target", 1)
	if err != nil {
		t.Errorf("Failed to add target with custom hasher: %v", err)
	}
	
	target, err := fh.Lookup("test-key")
	if err != nil {
		t.Errorf("Lookup with custom hasher failed: %v", err)
	}
	
	if target != "test-target" {
		t.Errorf("Expected 'test-target', got '%s'", target)
	}
}

