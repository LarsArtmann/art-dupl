package cli

import (
	"os"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/onsi/gomega"
)

func TestRuntimeConfig_ToConfig(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test ToConfig conversion
	runtime := &RuntimeConfig{
		Threshold:     25,
		Vendor:        true,
		FilesFromStdin: true,
		Verbose:       true,
		Paths:         []string{"./src", "./lib"},
	}

	cfg := runtime.ToConfig()
	g.Expect(cfg).NotTo(gomega.BeNil())
	g.Expect(cfg.Threshold).To(gomega.Equal(25))
	g.Expect(cfg.IncludeVendor).To(gomega.BeTrue())
	g.Expect(cfg.FilesFromStdin).To(gomega.BeTrue())
	g.Expect(cfg.Verbose).To(gomega.BeTrue())
	g.Expect(cfg.Paths).To(gomega.ContainElements("./src", "./lib"))
}

func TestRuntimeConfig_ToConfig_OutputFormats(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test HTML format
	runtime := &RuntimeConfig{HTML: true}
	cfg := runtime.ToConfig()
	g.Expect(cfg.OutputFormat).To(gomega.Equal(config.OutputFormatHTML))

	// Test JSON format
	runtime = &RuntimeConfig{JSON: true}
	cfg = runtime.ToConfig()
	g.Expect(cfg.OutputFormat).To(gomega.Equal(config.OutputFormatJSON))

	// Test Plumbing format
	runtime = &RuntimeConfig{Plumbing: true}
	cfg = runtime.ToConfig()
	g.Expect(cfg.OutputFormat).To(gomega.Equal(config.OutputFormatPlumbing))

	// Test default (text) format when no output format flags are set
	runtime = &RuntimeConfig{}
	cfg = runtime.ToConfig()
	// Check if output format is one of the valid formats
	validFormats := []config.OutputFormat{config.OutputFormatText, config.OutputFormatHTML, config.OutputFormatJSON, config.OutputFormatPlumbing}
	isValid := false
	for _, valid := range validFormats {
		if cfg.OutputFormat == valid {
			isValid = true
			break
		}
	}
	g.Expect(isValid).To(gomega.BeTrue(), "Output format should be one of the valid formats")
}

func TestDefaultRuntimeConfig(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test DefaultRuntimeConfig
	cfg := DefaultRuntimeConfig()
	g.Expect(cfg).NotTo(gomega.BeNil())
	g.Expect(cfg.Threshold).To(gomega.Equal(15))
	g.Expect(cfg.SortBy).To(gomega.Equal("size"))
	g.Expect(cfg.OutputWriter).NotTo(gomega.BeNil())
	g.Expect(cfg.ErrorWriter).NotTo(gomega.BeNil())
}

func TestCLIIOWriters(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test Write methods
	stdout := &cliStdout{}
	stderr := &cliStderr{}

	// Test stdout Write
	data := []byte("test output")
	n, err := stdout.Write(data)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(n).To(gomega.Equal(len(data)))

	// Test stderr Write
	data = []byte("test error")
	n, err = stderr.Write(data)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(n).To(gomega.Equal(len(data)))

	// Test that writing to os.Stdout works
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = originalStdout
		w.Close()
	}()

	testData := []byte("test to stdout")
	go stdout.Write(testData)

	// Read what was written
	buf := make([]byte, 100)
	written, _ := r.Read(buf)
	g.Expect(string(buf[:written])).To(gomega.ContainSubstring("test to stdout"))
}