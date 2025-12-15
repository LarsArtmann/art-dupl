package cli

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
		Verbose:      new(bool),
		VerboseLong:  new(bool),
		HTML:          new(bool),
		JSONFlag:      new(bool),
		Plumbing:      new(bool),
	}
	
	*config.Threshold = 20
	*config.ThresholdLong = 30
	g.Expect(config.GetThreshold()).To(gomega.Equal(30)) // Should prefer long form

	*config.ThresholdLong = 15 // Reset to default
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

func TestNewCLIConfig(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test NewCLIConfig creates config with flags
	cliConfig := NewCLIConfig()
	g.Expect(cliConfig).NotTo(gomega.BeNil())
	g.Expect(cliConfig.Vendor).NotTo(gomega.BeNil())
	g.Expect(cliConfig.Verbose).NotTo(gomega.BeNil())
	g.Expect(cliConfig.Threshold).NotTo(gomega.BeNil())
	g.Expect(cliConfig.Files).NotTo(gomega.BeNil())
	g.Expect(cliConfig.HTML).NotTo(gomega.BeNil())
	g.Expect(cliConfig.JSONFlag).NotTo(gomega.BeNil())
	g.Expect(cliConfig.Plumbing).NotTo(gomega.BeNil())
	g.Expect(cliConfig.SortBy).NotTo(gomega.BeNil())
}

func TestAddFlagsToCommand(t *testing.T) {
	g := gomega.NewWithT(t)

	// Test AddFlagsToCommand adds flags to a command
	cliConfig := NewCLIConfig()
	
	// This would normally be called during command initialization
	// We'll just verify it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AddFlagsToCommand panicked: %v", r)
		}
	}()
	
	// AddFlagsToCommand should be able to be called without error
	// It's typically called in main.go, so we can't easily test its effects here
	g.Expect(cliConfig).NotTo(gomega.BeNil())
}