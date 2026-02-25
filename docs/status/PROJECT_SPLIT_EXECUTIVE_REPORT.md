# Project Split Executive Report: art-dupl

This report outlines a proposal to split the `art-dupl` project into several smaller, more focused, and reusable Go projects. This approach aims to improve modularity, reduce coupling, enhance maintainability, and promote independent development and versioning of core functionalities.

## Current Project Overview (art-dupl)

The existing `art-dupl` project is a comprehensive tool for Go code duplication detection. It currently encompasses the CLI, configuration management, core detection algorithms, AST processing, job orchestration, output formatting, and various utilities within a single repository.

## Rationale for Splitting

The primary motivations for this split are:

- **Modularity:** Decoupling distinct functionalities into independent modules.
- **Reusability:** Allowing core libraries (detection, printing) to be used by other tools or projects.
- **Maintainability:** Easier to manage, test, and develop individual components.
- **Clear Boundaries:** Enforcing clear API contracts between components.
- **Independent Evolution:** Each project can evolve and be versioned independently.

## Proposed Project Structure

The `art-dupl` project can be logically split into the following highly focused projects:

---

### 1. `go-clones-cli` (CLI Application)

- **Purpose:** The main user-facing command-line application that orchestrates the clone detection process by integrating the core libraries.
- **Original Packages:** `cmd/`, `cli/`. It would also import and utilize `go-clones-config`, `go-clones-core`, and `go-clones-printer`.
- **Key Features:**
  - Command-line interface parsing and flag handling.
  - Orchestration of the clone detection workflow.
  - Loading and applying configuration.
  - Presenting results to the user using the output formatter.
- **Dependencies:** `go-clones-config`, `go-clones-core`, `go-clones-printer`.

---

### 2. `go-clones-core` (Detection Algorithms & AST Processing Library)

- **Purpose:** A reusable library containing the core logic for abstract syntax tree (AST) processing and clone detection algorithms. This would be the "engine" of code duplication analysis.
- **Original Packages:** `suffixtree/`, `hash/`, `detection/`, `syntax/`, `job/`, `domain/`, `types/`, `errors/`.
- **Key Features:**
  - Parsing Go source files into ASTs.
  - Tokenization and serialization of ASTs.
  - Suffix tree algorithm for structural clone detection.
  - Rolling hash algorithm for content-based detection.
  - Multi-method detection coordination.
  - Core domain models (Clone, CloneGroup, StringPool).
  - Robust error handling for detection-related issues.
- **Dependencies:** Minimal, potentially `go-clones-utils` for common helpers.

---

### 3. `go-clones-printer` (Output Formatting Library)

- **Purpose:** A dedicated library for generating various output formats for clone detection results. This component would be responsible solely for presentation logic.
- **Original Packages:** `printer/`, `adapter/`.
- **Key Features:**
  - Interfaces for different output formats (text, HTML, JSON, plumbing, CSV).
  - Implementations for each output format, handling the rendering of `domain.CloneGroup` objects.
  - Abstraction layer for easy extension with new formats.
- **Dependencies:** `go-clones-core` (specifically `domain/` package for input types), potentially `go-clones-utils`.

---

### 4. `go-clones-config` (Configuration Management Library)

- **Purpose:** A standalone library for managing application configuration, including loading from files, validation, and migration between versions.
- **Original Packages:** `config/`, `migration/`.
- **Key Features:**
  - Loading configuration from JSON files or other sources.
  - Validation of configuration parameters.
  - Migration utilities for handling configuration changes across different tool versions.
  - Type-safe configuration structures.
- **Dependencies:** Minimal.

---

### 5. `go-clones-utils` (General Utilities Library)

- **Purpose:** A common library for shared utility functions and helper code that are not tightly coupled to any specific domain (e.g., logging, filtering, position tracking).
- **Original Packages:** `pkg/`, `internal/`, `lib/`. Specific parts of these packages would be moved here.
- **Key Features:**
  - Shared logging infrastructure (`pkg/logger`).
  - File filtering logic (`pkg/filter`).
  - Position tracking (`pkg/position`).
  - Other general-purpose helper functions.
- **Dependencies:** None (it should be foundational).

---

## Conclusion

This proposed project split into `go-clones-cli`, `go-clones-core`, `go-clones-printer`, `go-clones-config`, and `go-clones-utils` creates a more robust, maintainable, and extensible architecture. Each new project would have a clear responsibility, fostering independent development and allowing for greater flexibility in how the core clone detection capabilities are consumed.
