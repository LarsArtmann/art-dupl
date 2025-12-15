package testutils

import (
	"math/rand"
	"time"
)

// UniqueTestHelper provides unique function for test data generation
func UniqueTestHelper() string {
	return "unique_" + generateRandomSuffix()
}

// generateRandomSuffix generates a random suffix for uniqueness
func generateRandomSuffix() string {
	rand.Seed(time.Now().UnixNano())
	return string(rune('a' + rand.Intn(26)))
}

// UniqueFunction returns a unique function template
func UniqueFunction() string {
	return `func unique` + UniqueTestHelper() + `() {
	// This is unique
	return nil
}`
}