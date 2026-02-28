package domain

import (
	"testing"
)

// createCloneTestCase creates a test case for Clone validation.
func createCloneTestCase(
	name string,
	startLine, endLine, startPos, endPos int,
	wantErr bool,
) struct {
	name    string
	clone   Clone
	wantErr bool
} {
	return struct {
		name    string
		clone   Clone
		wantErr bool
	}{
		name: name,
		clone: Clone{
			StartLine: MustNewLineNumber(uint16(startLine)),
			EndLine:   MustNewLineNumber(uint16(endLine)),
			StartPos:  NewBytePosition(uint32(startPos)),
			EndPos:    NewBytePosition(uint32(endPos)),
			Status:    FileProcessingStateCompleted,
		},
		wantErr: wantErr,
	}
}

// TestClone_IsValid tests Clone.IsValid method.
func TestClone_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		clone   Clone
		wantErr bool
	}{
		createCloneTestCase("valid clone", 10, 20, 100, 200, false),
		{
			name: "end line before start line",
			clone: Clone{
				StartLine: MustNewLineNumber(20),
				EndLine:   MustNewLineNumber(10),
				Status:    FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		createCloneTestCase("end pos before start pos", 10, 20, 200, 100, true),
		{
			name: "invalid status",
			clone: Clone{
				StartLine: MustNewLineNumber(10),
				EndLine:   MustNewLineNumber(20),
				Status:    FileProcessingState("invalid"),
			},
			wantErr: true,
		},
		{
			name: "zero positions are valid",
			clone: Clone{
				StartLine: MustNewLineNumber(10),
				EndLine:   MustNewLineNumber(20),
				StartPos:  0,
				EndPos:    0,
				Status:    FileProcessingStateCompleted,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.clone.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Clone.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCloneGroup_IsValid tests CloneGroup.IsValid method.
func TestCloneGroup_IsValid(t *testing.T) {
	validClone := Clone{
		StartLine: MustNewLineNumber(10),
		EndLine:   MustNewLineNumber(20),
		Status:    FileProcessingStateCompleted,
	}

	tests := []struct {
		name       string
		cloneGroup CloneGroup
		wantErr    bool
	}{
		{
			name: "valid clone group",
			cloneGroup: CloneGroup{
				ID:       "group-1",
				Clones:   []Clone{validClone},
				Hash:     "abc123",
				Size:     10,
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: false,
		},
		{
			name: "empty clones",
			cloneGroup: CloneGroup{
				ID:       "group-2",
				Clones:   []Clone{},
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "nil clones",
			cloneGroup: CloneGroup{
				ID:       "group-3",
				Clones:   nil,
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			cloneGroup: CloneGroup{
				ID:       "group-4",
				Clones:   []Clone{validClone},
				Severity: CloneSeverity("invalid"),
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			cloneGroup: CloneGroup{
				ID:       "group-5",
				Clones:   []Clone{validClone},
				Severity: CloneSeverityMedium,
				Status:   FileProcessingState("invalid"),
			},
			wantErr: true,
		},
		{
			name: "invalid clone in group",
			cloneGroup: CloneGroup{
				ID: "group-6",
				Clones: []Clone{
					{
						StartLine: MustNewLineNumber(20),
						EndLine:   MustNewLineNumber(10),
						Status:    FileProcessingStateCompleted,
					},
				},
				Severity: CloneSeverityMedium,
				Status:   FileProcessingStateCompleted,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cloneGroup.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("CloneGroup.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
