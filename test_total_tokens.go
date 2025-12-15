package main

// Small function with 3 tokens
func small() {
	println("small")
}

// Medium function with 5 tokens
func medium() {
	println("medium")
	println("medium")
}

// Large function with 8 tokens
func large() {
	println("large")
	println("large")
	println("large")
	println("large")
}

// Another small function (duplicate of first)
func small() {
	println("small")
}

// Another medium function (duplicate of second)
func medium() {
	println("medium")
	println("medium")
}