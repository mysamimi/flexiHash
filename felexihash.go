package flexihash

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
)

// Hasher is the interface for hash functions
type Hasher interface {
	Hash(string) int
}

// Crc32Hasher uses CRC32 to hash values (matches PHP behavior)
type Crc32Hasher struct{}

// Hash returns a signed 32-bit CRC32 hash (matches PHP crc32 behavior)
func (h *Crc32Hasher) Hash(str string) int {
	return int(int32(crc32.ChecksumIEEE([]byte(str))))
}

// Md5Hasher uses MD5 to hash values (matches PHP Flexihash MD5 hasher)
// Uses first 8 hexits (32 bits) of MD5 hash
type Md5Hasher struct{}

// Hash returns a 32-bit hash from MD5 (matches PHP Flexihash behavior)
func (h *Md5Hasher) Hash(str string) int {
	hash := md5.Sum([]byte(str))
	hexStr := hex.EncodeToString(hash[:])
	// Take first 8 characters (32 bits) and convert to int
	var result int64
	for i := 0; i < 8 && i < len(hexStr); i++ {
		result = result*16 + int64(hexDigitToInt(hexStr[i]))
	}
	return int(result)
}

func hexDigitToInt(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'f' {
		return int(c - 'a' + 10)
	}
	if c >= 'A' && c <= 'F' {
		return int(c - 'A' + 10)
	}
	return 0
}

// FlexiHash implements consistent hashing
type FlexiHash struct {
	replicas               int
	hasher                 Hasher
	targetCount            int
	positionToTarget       map[int]string
	targetToPositions      map[string][]int
	positionToTargetSorted bool
	sortedPositions        []int
	positionCount          int
}

// NewFlexiHash creates a new FlexiHash instance with default settings
func NewFlexiHash() *FlexiHash {
	return NewFlexiHashWithHasher(nil, 0)
}

// NewFlexiHashWithHasher creates a FlexiHash with custom hasher and replicas
func NewFlexiHashWithHasher(hasher Hasher, replicas int) *FlexiHash {
	if hasher == nil {
		hasher = &Crc32Hasher{}
	}
	if replicas == 0 {
		replicas = 64
	}
	return &FlexiHash{
		replicas:          replicas,
		hasher:            hasher,
		positionToTarget:  make(map[int]string),
		targetToPositions: make(map[string][]int),
	}
}

// AddTarget adds a target to the hash ring with optional weight
func (fh *FlexiHash) AddTarget(target string, weight float64) error {
	if weight == 0 {
		weight = 1
	}
	if _, exists := fh.targetToPositions[target]; exists {
		return errors.New("Target '" + target + "' already exists.")
	}
	fh.targetToPositions[target] = []int{}

	// Hash the target into multiple positions
	replicaCount := int(float64(fh.replicas) * weight)
	for i := 0; i < replicaCount; i++ {
		position := fh.hasher.Hash(target + strconv.Itoa(i))
		fh.positionToTarget[position] = target
		fh.targetToPositions[target] = append(fh.targetToPositions[target], position)
		fh.positionCount++
	}

	fh.positionToTargetSorted = false
	fh.targetCount++
	return nil
}

// AddTargets adds multiple targets with optional weight
func (fh *FlexiHash) AddTargets(targets []string, weight float64) error {
	if weight == 0 {
		weight = 1
	}
	for _, target := range targets {
		if err := fh.AddTarget(target, weight); err != nil {
			return err
		}
	}
	return nil
}

// RemoveTarget removes a target from the hash ring
func (fh *FlexiHash) RemoveTarget(target string) error {
	positions, exists := fh.targetToPositions[target]
	if !exists {
		return errors.New("Target '" + target + "' does not exist.")
	}

	for _, position := range positions {
		delete(fh.positionToTarget, position)
	}
	delete(fh.targetToPositions, target)

	fh.positionToTargetSorted = false
	fh.targetCount--
	return nil
}

// GetAllTargets returns a list of all potential targets
func (fh *FlexiHash) GetAllTargets() []string {
	var targets []string
	for target := range fh.targetToPositions {
		targets = append(targets, target)
	}
	return targets
}

// Lookup finds the target for a given resource
func (fh *FlexiHash) Lookup(resource string) (string, error) {
	targets, err := fh.LookupList(resource, 1)
	if err != nil {
		return "", err
	}
	if len(targets) == 0 {
		return "", errors.New("No targets exist")
	}
	return targets[0], nil
}

// LookupList returns a list of targets for the resource, in order of precedence
func (fh *FlexiHash) LookupList(resource string, requestedCount int) ([]string, error) {
	if requestedCount < 1 {
		return nil, errors.New("Invalid count requested")
	}

	// Handle no targets
	if len(fh.positionToTarget) == 0 {
		return []string{}, nil
	}

	// Optimize single target
	if fh.targetCount == 1 {
		// Return unique targets only
		result := []string{}
		seen := make(map[string]bool)
		for _, target := range fh.positionToTarget {
			if !seen[target] {
				result = append(result, target)
				seen[target] = true
			}
		}
		return result, nil
	}

	// Hash resource to a position
	resourcePosition := fh.hasher.Hash(resource)

	var results []string

	fh.sortPositionTargets()
	positions := fh.sortedPositions

	// Binary search for the first position greater than resource position
	low := 0
	high := fh.positionCount - 1
	notfound := false

	for high >= low || notfound {
		probe := (high + low) / 2

		if !notfound && positions[probe] <= resourcePosition {
			low = probe + 1
		} else if probe == 0 || resourcePosition > positions[probe-1] || notfound {
			if notfound {
				// Binary search failed to find any position greater than resource position
				// In this case, wrap around to first position
				probe = 0
			}

			results = append(results, fh.positionToTarget[positions[probe]])

			if requestedCount > 1 {
				for i := requestedCount - 1; i > 0; i-- {
					probe++
					if probe > fh.positionCount-1 {
						probe = 0 // cycle
					}
					results = append(results, fh.positionToTarget[positions[probe]])
				}
			}
			break
		} else {
			high = probe - 1
		}
	}

	// Return unique targets
	seen := make(map[string]bool)
	uniqueResults := []string{}
	for _, target := range results {
		if !seen[target] {
			uniqueResults = append(uniqueResults, target)
			seen[target] = true
		}
	}

	return uniqueResults, nil
}

// sortPositionTargets sorts the internal mapping by position
func (fh *FlexiHash) sortPositionTargets() {
	if !fh.positionToTargetSorted {
		fh.sortedPositions = make([]int, 0, len(fh.positionToTarget))
		for pos := range fh.positionToTarget {
			fh.sortedPositions = append(fh.sortedPositions, pos)
		}
		// Sort by position
		sort.Ints(fh.sortedPositions)
		fh.positionToTargetSorted = true
		fh.positionCount = len(fh.sortedPositions)
	}
}
