package testutils

import (
	"math/rand"
)

// UniqueTestHelper provides unique function for test data generation
func UniqueTestHelper() string {
	return "unique_" + generateRandomSuffix()
}

// generateRandomSuffix generates a random suffix for uniqueness
func generateRandomSuffix() string {
	// As of Go 1.20, rand is automatically seeded, no need to call Seed()
	// Generate 3-letter suffix for 26^3 = 17,576 possible combinations
	const suffixLength = 3
	suffix := make([]byte, suffixLength)
	for i := range suffixLength {
		suffix[i] = byte('a' + rand.Intn(26))
	}
	return string(suffix)
}

// UniqueFunction returns a unique function template
func UniqueFunction() string {
	return `func unique` + UniqueTestHelper() + `() {
	// This is unique
	return nil
}`
}
