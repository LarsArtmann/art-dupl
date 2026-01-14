package types_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/types"
)

var _ = Describe("Type Safety: Result[T]", func() {
	Context("When creating successful results", func() {
		It("should create valid ok results", func() {
			result := types.Ok("test value")

			Expect(result.IsOk()).To(BeTrue())
			Expect(result.IsErr()).To(BeFalse())
			value, err := result.Unwrap()
			Expect(value).To(Equal("test value"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should support generic types", func() {
			result := types.Ok(42)
			Expect(result.IsOk()).To(BeTrue())
			Expect(result.Value).To(Equal(42))
		})
	})

	Context("When creating error results", func() {
		It("should create valid error results", func() {
			testErr := errors.New("test error")
			result := types.Err[string](testErr)

			Expect(result.IsOk()).To(BeFalse())
			Expect(result.IsErr()).To(BeTrue())
			value, err := result.Unwrap()
			Expect(value).To(Equal(""))
			Expect(err).To(Equal(testErr))
		})

		It("should create formatted error results", func() {
			result := types.Errf[string]("validation failed: %s", "missing field")

			Expect(result.IsOk()).To(BeFalse())
			Expect(result.Error).To(HaveOccurred())
			Expect(result.Error.Error()).To(ContainSubstring("validation failed"))
		})
	})

	Context("When mapping results", func() {
		It("should map success values", func() {
			result := types.Map(types.Ok(5), func(x int) string {
				return fmt.Sprintf("value-%d", x)
			})

			Expect(result.IsOk()).To(BeTrue())
			Expect(result.Value).To(Equal("value-5"))
		})

		It("should propagate errors through map", func() {
			err := errors.New("original error")
			result := types.Map(types.Err[int](err), func(x int) string {
				return "should not reach here"
			})

			Expect(result.IsErr()).To(BeTrue())
			Expect(result.Error).To(Equal(err))
		})
	})

	Context("When providing defaults", func() {
		It("should return value for successful results", func() {
			result := types.Ok("actual")
			Expect(result.Or("default")).To(Equal("actual"))
		})

		It("should return default for error results", func() {
			result := types.Err[string](errors.New("error"))
			Expect(result.Or("default")).To(Equal("default"))
		})
	})
})

var _ = Describe("Type Safety: Option[T]", func() {
	Context("When creating options", func() {
		It("should create some option", func() {
			opt := types.Some("value")

			Expect(opt.IsSome()).To(BeTrue())
			Expect(opt.IsNone()).To(BeFalse())
			Expect(opt.Unwrap()).To(Equal("value"))
		})

		It("should create none option", func() {
			opt := types.None[string]()

			Expect(opt.IsSome()).To(BeFalse())
			Expect(opt.IsNone()).To(BeTrue())
			Expect(opt.Unwrap()).To(Equal(""))
		})
	})

	Context("When providing defaults", func() {
		It("should return value for some option", func() {
			opt := types.Some("actual")
			Expect(opt.Or("default")).To(Equal("actual"))
		})

		It("should return default for none option", func() {
			opt := types.None[string]()
			Expect(opt.Or("default")).To(Equal("default"))
		})
	})

	Context("When converting to results", func() {
		It("should create successful result from some", func() {
			opt := types.Some("value")
			result := opt.ToResult("no value provided")

			Expect(result.IsOk()).To(BeTrue())
			Expect(result.Value).To(Equal("value"))
		})

		It("should create error result from none", func() {
			opt := types.None[string]()
			result := opt.ToResult("no value provided")

			Expect(result.IsErr()).To(BeTrue())
			Expect(result.Error.Error()).To(ContainSubstring("no value provided"))
		})
	})

	Context("When filtering options", func() {
		It("should keep some if predicate passes", func() {
			opt := types.Some(42)
			filtered := opt.Filter(func(x int) bool { return x > 20 })

			Expect(filtered.IsSome()).To(BeTrue())
			Expect(filtered.Unwrap()).To(Equal(42))
		})

		It("should convert to none if predicate fails", func() {
			opt := types.Some(10)
			filtered := opt.Filter(func(x int) bool { return x > 20 })

			Expect(filtered.IsNone()).To(BeTrue())
		})

		It("should keep none unchanged", func() {
			opt := types.None[int]()
			filtered := opt.Filter(func(x int) bool { return x > 20 })

			Expect(filtered.IsNone()).To(BeTrue())
		})
	})
})

// NOTE: Enum tests below are commented out because enums (DetectionState, AnalysisMode, FileProcessingState)
// have been moved to domain package during enum consolidation phase.
// Tests for these enums should be added to domain/clone_test.go or domain package tests.
/*
var _ = Describe("Type Safety: Enums", func() {
	Context("DetectionState enum", func() {
		It("should validate detection states", func() {
			validStates := []types.DetectionState{
				types.DetectionStateUnknown,
				types.DetectionStatePending,
				types.DetectionStateRunning,
				types.DetectionStateCompleted,
				types.DetectionStateFailed,
			}

			for _, state := range validStates {
				Expect(state.IsValid()).To(BeTrue(), fmt.Sprintf("State %s should be valid", state))
			}
		})

		It("should reject invalid detection states", func() {
			invalidState := types.DetectionState("invalid")
			Expect(invalidState.IsValid()).To(BeFalse())
		})

		It("should marshal and unmarshal JSON correctly", func() {
			original := types.DetectionStateCompleted
			data, err := json.Marshal(original)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"completed"`))

			var unmarshaled types.DetectionState
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).ToNot(HaveOccurred())
			Expect(unmarshaled).To(Equal(original))
		})

		It("should reject invalid JSON during unmarshal", func() {
			invalidJSON := []byte(`"invalid-state"`)
			var state types.DetectionState
			err := json.Unmarshal(invalidJSON, &state)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid detection state"))
		})
	})

	Context("AnalysisMode enum", func() {
		It("should validate analysis modes", func() {
			validModes := []types.AnalysisMode{
				types.AnalysisModeFull,
				types.AnalysisModeIncremental,
				types.AnalysisModeQuick,
				types.AnalysisModeDeep,
			}

			for _, mode := range validModes {
				Expect(mode.IsValid()).To(BeTrue(), fmt.Sprintf("Mode %s should be valid", mode))
			}
		})

		It("should reject invalid analysis modes", func() {
			invalidMode := types.AnalysisMode("invalid")
			Expect(invalidMode.IsValid()).To(BeFalse())
		})

		It("should marshal and unmarshal JSON correctly", func() {
			original := types.AnalysisModeFull
			data, err := json.Marshal(original)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"full"`))

			var unmarshaled types.AnalysisMode
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).ToNot(HaveOccurred())
			Expect(unmarshaled).To(Equal(original))
		})
	})

	Context("FileProcessingState enum", func() {
		It("should validate file processing states", func() {
			validStates := []types.FileProcessingState{
				types.FileProcessingStateQueued,
				types.FileProcessingStateReading,
				types.FileProcessingStateProcessing,
				types.FileProcessingStateCompleted,
				types.FileProcessingStateError,
			}

			for _, state := range validStates {
				Expect(state.IsValid()).To(BeTrue(), fmt.Sprintf("State %s should be valid", state))
			}
		})

		It("should reject invalid file processing states", func() {
			invalidState := types.FileProcessingState("invalid")
			Expect(invalidState.IsValid()).To(BeFalse())
		})

		It("should marshal and unmarshal JSON correctly", func() {
			original := types.FileProcessingStateCompleted
			data, err := json.Marshal(original)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`"completed"`))

			var unmarshaled types.FileProcessingState
			err = json.Unmarshal(data, &unmarshaled)
			Expect(err).ToNot(HaveOccurred())
			Expect(unmarshaled).To(Equal(original))
		})
	})
})
*/
