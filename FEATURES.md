# art-dupl Feature Documentation

> **Last Updated:** 2025-12-15  
> **Version:** Analysis of main branch

## Overview

**art-dupl** is a Go tool for finding code clones using suffix tree algorithms and hash-based detection. It analyzes abstract syntax trees (ASTs) to find structural code clones while ignoring literal values.

## 🚀 Core Features

### Detection Methods

| Feature                              | Status           | Description                                                                     |
| ------------------------------------ | ---------------- | ------------------------------------------------------------------------------- |
| **Suffix Tree Detection (art-dupl)** | FULLY_FUNCTIONAL | Original detection method using suffix tree algorithms on serialized ASTs       |
| **Hash-Based Detection**             | FULLY_FUNCTIONAL | Alternative detection method using SHA1 hashing for faster clone identification |
| **Multi-Detection Mode**             | FULLY_FUNCTIONAL | Run both detection methods simultaneously for comprehensive analysis            |

### Output Formats

| Feature                      | Status           | Description                                                             |
| ---------------------------- | ---------------- | ----------------------------------------------------------------------- |
| **Text Output**              | FULLY_FUNCTIONAL | Human-readable clone listing with file paths and line numbers (default) |
| **HTML Output**              | FULLY_FUNCTIONAL | Detailed report with syntax-highlighted code fragments                  |
| **JSON Output**              | FULLY_FUNCTIONAL | Structured data with metadata, statistics, and summary information      |
| **Plumbing Output**          | FULLY_FUNCTIONAL | Machine-readable format for script integration and CI/CD                |
| **Batch Generation (--all)** | FULLY_FUNCTIONAL | Generate all output formats for all detection methods at once           |

### Sorting Options

| Feature                  | Status               | Description                                                                                                          |
| ------------------------ | -------------------- | -------------------------------------------------------------------------------------------------------------------- |
| **Size Sorting**         | FULLY_FUNCTIONAL     | Sort clone groups by token count (largest first)                                                                     |
| **Occurrence Sorting**   | FULLY_FUNCTIONAL     | Sort clone groups by number of files (most widespread first)                                                         |
| **Hash Sorting**         | FULLY_FUNCTIONAL     | Sort clone groups by hash value (alphabetical)                                                                       |
| **Total Tokens Sorting** | PARTIALLY_FUNCTIONAL | Implementation exists but not exposed in config (SortClonesByTotalTokens function exists but not in AllSortCriteria) |

## 🔧 Configuration Features

| Feature                      | Status           | Description                                          |
| ---------------------------- | ---------------- | ---------------------------------------------------- |
| **Command-Line Flags**       | FULLY_FUNCTIONAL | All settings configurable via CLI flags              |
| **JSON Configuration Files** | FULLY_FUNCTIONAL | Persistent configuration in JSON format              |
| **Configuration Merging**    | FULLY_FUNCTIONAL | Intelligent merging of file and CLI configurations   |
| **Threshold Control**        | FULLY_FUNCTIONAL | Adjustable minimum token sequence size (default: 15) |
| **Vendor Directory Control** | FULLY_FUNCTIONAL | Option to include/exclude vendor directory           |
| **File Input from Stdin**    | FULLY_FUNCTIONAL | Read file paths from stdin with `-files` flag        |

## 🖥️ Professional CLI Features (via Fang)

| Feature                     | Status           | Description                                                      |
| --------------------------- | ---------------- | ---------------------------------------------------------------- |
| **Styled Help Output**      | FULLY_FUNCTIONAL | Rich, themed help text with examples                             |
| **Auto-Completion**         | FULLY_FUNCTIONAL | Tab completion for bash, zsh, fish, and powershell shells        |
| **Man Page Generation**     | FULLY_FUNCTIONAL | Generate manual pages for documentation                          |
| **Version Information**     | FULLY_FUNCTIONAL | Detailed version, commit, and build information                  |
| **Error Handling**          | FULLY_FUNCTIONAL | Context-aware error messages with suggestions                    |
| **Completion Descriptions** | FULLY_FUNCTIONAL | Option to disable completion descriptions with --no-descriptions |

## 🔍 Advanced Features

| Feature                     | Status               | Description                                                  |
| --------------------------- | -------------------- | ------------------------------------------------------------ |
| **Verbose Logging**         | FULLY_FUNCTIONAL     | Detailed progress information with `-v` flag                 |
| **Custom Output Directory** | FULLY_FUNCTIONAL     | Specify output directory for batch generation                |
| **Performance Profiling**   | PARTIALLY_FUNCTIONAL | Hidden `--profile` flag exists but implementation incomplete |
| **Execution Timeout**       | PARTIALLY_FUNCTIONAL | Hidden `--timeout` flag exists but implementation incomplete |

## 📊 Analysis Features

| Feature                | Status           | Description                                      |
| ---------------------- | ---------------- | ------------------------------------------------ |
| **File Counting**      | FULLY_FUNCTIONAL | Counts and reports total files analyzed          |
| **Clone Grouping**     | FULLY_FUNCTIONAL | Groups duplicates by hash signature              |
| **Statistics Summary** | FULLY_FUNCTIONAL | JSON output includes comprehensive statistics    |
| **Fragment Analysis**  | FULLY_FUNCTIONAL | Analyzes code fragments with start/end positions |

## 🚫 Known Limitations

| Limitation         | Impact | Status                                             |
| ------------------ | ------ | -------------------------------------------------- |
| **Go Only**        | High   | Only analyzes Go source files (.go extension)      |
| **Large Files**    | Medium | May have performance issues with very large files  |
| **Memory Usage**   | Medium | Can consume significant memory for large codebases |
| **Hash Collision** | Low    | Theoretical possibility of SHA1 collisions         |

## 🔮 Experimental Features

| Feature                   | Status       | Description                                      |
| ------------------------- | ------------ | ------------------------------------------------ |
| **Performance Profiling** | EXPERIMENTAL | `--profile` flag exists but needs implementation |
| **Custom Timeouts**       | EXPERIMENTAL | `--timeout` flag exists but needs implementation |

## 📋 Feature Usage Examples

### Basic Usage

```bash
# Default analysis
art-dupl

# Higher threshold
art-dupl -t 50

# Specific paths
art-dupl ./src ./lib
```

### Output Formats

```bash
# HTML report
art-dupl -html > report.html

# JSON with statistics
art-dupl -json -t 20

# Machine-readable for scripts
art-dupl -plumbing -sort occurrence
```

### Professional Features

```bash
# Generate shell completions
art-dupl completion bash
art-dupl completion zsh
art-dupl completion fish

# Generate man page
art-dupl man > art-dupl.1

# Check version
art-dupl --version
```

### Advanced Usage

```bash
# Configuration file
art-dupl -config dupl.json

# Batch generation
art-dupl --all --output-dir ./reports

# Multiple detection methods
art-dupl -detection-methods "hash,art-dupl"

# Analyze test files only
find . -name '*_test.go' | art-dupl -files

# Generate completions with no descriptions
art-dupl completion bash --no-descriptions

# Powershell completion
art-dupl completion powershell
```

## 🏗️ Architecture Components

| Component       | Status           | Description                                      |
| --------------- | ---------------- | ------------------------------------------------ |
| **suffixtree/** | FULLY_FUNCTIONAL | Core suffix tree implementation                  |
| **syntax/**     | FULLY_FUNCTIONAL | AST handling, serialization, and node processing |
| **job/**        | FULLY_FUNCTIONAL | Orchestrates file parsing and tree building      |
| **printer/**    | FULLY_FUNCTIONAL | Output formatting for all supported formats      |
| **hash/**       | FULLY_FUNCTIONAL | Hash-based detection implementation              |
| **config/**     | FULLY_FUNCTIONAL | Configuration management with validation         |
| **detection/**  | FULLY_FUNCTIONAL | Multi-detector coordination                      |

## 🧪 Testing Status

| Test Type             | Coverage | Status                                     |
| --------------------- | -------- | ------------------------------------------ |
| **Unit Tests**        | Good     | Comprehensive unit tests for core packages |
| **Integration Tests** | Good     | End-to-end workflow testing                |
| **BDD Tests**         | Good     | Behavior-driven development tests          |
| **Performance Tests** | Limited  | Basic performance testing exists           |

## 📝 Documentation Quality

| Documentation         | Status    | Notes                          |
| --------------------- | --------- | ------------------------------ |
| **README.md**         | GOOD      | Comprehensive with examples    |
| **CLI Help**          | EXCELLENT | Rich, styled help with Fang    |
| **Code Comments**     | GOOD      | Adequate commenting throughout |
| **API Documentation** | LIMITED   | No generated API docs          |

## 🔮 Future Roadmap

### High Priority

- [ ] Complete performance profiling implementation
- [ ] Add execution timeout functionality
- [ ] Improve memory efficiency for large codebases
- [ ] Add support for additional languages (TypeScript, JavaScript)

### Medium Priority

- [ ] Generate comprehensive API documentation
- [ ] Add more sorting criteria options
- [ ] Implement clone similarity scoring
- [ ] Add duplicate suppression rules

### Low Priority

- [ ] Web UI for report visualization
- [ ] Integration with IDE plugins
- [ ] Historical trend analysis
- [ ] Clone impact analysis

## 📊 Overall Project Health

| Metric                   | Score | Notes                                           |
| ------------------------ | ----- | ----------------------------------------------- |
| **Feature Completeness** | 85%   | Most features fully implemented                 |
| **Code Quality**         | 85%   | Well-structured, good separation of concerns    |
| **Documentation**        | 80%   | Good user docs, needs API docs                  |
| **Testing**              | 75%   | Good coverage, could use more performance tests |
| **Production Readiness** | 85%   | Ready for production use                        |

## 🔚 Conclusion

art-dupl is a mature, production-ready tool with comprehensive code clone detection capabilities. The core functionality is robust, with multiple detection methods, output formats, and professional CLI features. While there are some incomplete experimental features, the main use cases are well-covered and reliable.

The tool successfully combines the simplicity of the original dupl tool with modern enhancements like JSON output, configuration files, and a professional CLI experience powered by Fang.

---

_This documentation reflects the state of the main branch as of 2025-12-15. For the latest information, check the repository directly._
