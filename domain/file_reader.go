package domain

// FileReaderFunc represents a function that reads file contents by filename.
// This is the canonical definition — printer and SDK packages reference this
// type via aliases to prevent drift between identical function-type definitions.
type FileReaderFunc func(filename string) ([]byte, error)
