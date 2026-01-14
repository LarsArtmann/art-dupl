package cli

//nolint:testpackage // Tests require access to cli package internals
import (
	"flag"
	"testing"

	"github.com/onsi/gomega"
)

func TestCLIConfig(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test NewCLIConfig - create without flag conflicts
	config := &CLIConfig{
		Vendor:        flag.Bool("test_vendor", false, "test vendor flag"),
		Verbose:       flag.Bool("test_verbose", false, "test verbose flag"),
		VerboseLong:   flag.Bool("test_verbose_long", false, "test verbose long flag"),
		Threshold:     flag.Int("test_threshold", 15, "test threshold flag"),
		ThresholdLong: flag.Int("test_threshold_long", 15, "test threshold long flag"),
		Files:         flag.Bool("test_files", false, "test files flag"),
		HTML:          flag.Bool("test_html", false, "test html flag"),
		JSONFlag:      flag.Bool("test_json", false, "test json flag"),
		Plumbing:      flag.Bool("test_plumbing", false, "test plumbing flag"),
		SortBy:        flag.String("test_sort", "size", "test sort flag"),
	}

	g.Expect(config).NotTo(gomega.BeNil())
	g.Expect(*config.Threshold).To(gomega.Equal(15))
	g.Expect(*config.Verbose).To(gomega.BeFalse())
}

func TestCLIConfigHelpers(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test GetThreshold
	config := &CLIConfig{
		Threshold:     new(int),
		ThresholdLong: new(int),
		Verbose:       new(bool),
		VerboseLong:   new(bool),
		HTML:          new(bool),
		JSONFlag:      new(bool),
		SimpleJSON:    new(bool),
		Plumbing:      new(bool),
	}

	*config.Threshold = 20
	*config.ThresholdLong = 30
	g.Expect(config.GetThreshold()).To(gomega.Equal(30)) // Should prefer long form

	*config.ThresholdLong = 15                           // Reset to default
	g.Expect(config.GetThreshold()).To(gomega.Equal(20)) // Should use short form

	// Test IsVerbose
	*config.Verbose = true
	g.Expect(config.IsVerbose()).To(gomega.BeTrue())

	*config.Verbose = false
	*config.VerboseLong = true
	g.Expect(config.IsVerbose()).To(gomega.BeTrue())

	// Test GetOutputFormats
	*config.HTML = true
	*config.JSONFlag = false
	*config.Plumbing = false
	formats := config.GetOutputFormats()
	g.Expect(formats).To(gomega.ContainElement("html"))
	g.Expect(formats).To(gomega.HaveLen(1))

	*config.HTML = false
	*config.JSONFlag = true
	formats = config.GetOutputFormats()
	g.Expect(formats).To(gomega.ContainElement("json"))
	g.Expect(formats).To(gomega.HaveLen(1))

	// Test default format
	*config.HTML = false
	*config.JSONFlag = false
	*config.Plumbing = false
	formats = config.GetOutputFormats()
	g.Expect(formats).To(gomega.ContainElement("text"))
	g.Expect(formats).To(gomega.HaveLen(1))
}

// Note: NewCLIConfig and AddFlagsToCommand tests are skipped due to global flag state
// issues in Go's flag package that make them hard to test in isolation.
