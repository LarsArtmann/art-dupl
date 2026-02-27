package domain

// Repository represents source code repository.
type Repository struct {
	Path        string       `json:"path"`
	Name        string       `json:"name"`
	Files       []SourceFile `json:"files"`
	Language    string       `json:"language"`
	Framework   string       `json:"framework"`
	Size        uint64       `json:"size"`
	LastIndexed string       `json:"lastIndexed"`
}

func (r Repository) IsValid() error {
	if r.Path == "" {
		return ErrRepositoryPathEmpty
	}

	if r.Name == "" {
		return ErrRepositoryNameEmpty
	}

	if r.Language == "" {
		return ErrRepositoryLanguageEmpty
	}

	return nil
}

// SourceFile represents a source code file.
type SourceFile struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Size         uint64   `json:"size"`
	Language     string   `json:"language"`
	Content      string   `json:"content,omitempty"`
	Hash         string   `json:"hash"`
	Dependencies []string `json:"dependencies,omitempty"`
}

func (sf SourceFile) IsValid() error {
	return validateFields(
		validationRule{sf.Path != "", "source file path cannot be empty"},
		validationRule{sf.Name != "", "source file name cannot be empty"},
		validationRule{sf.Size > 0, "source file size cannot be zero"},
		validationRule{sf.Hash != "", "source file hash cannot be empty"},
	)
}
