package entities

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"teiwrappergolang/internal/domain/errors"
)

type ValidationConfig struct {
	MaxInputLength    int
	MaxBatchSize      int
	MaxSentencesCount int
	AllowEmptyStrings bool
}

func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		MaxInputLength:    8192,
		MaxBatchSize:      32,
		MaxSentencesCount: 100,
		AllowEmptyStrings: false,
	}
}

type Validator struct {
	config *ValidationConfig
}

func NewValidator(config *ValidationConfig) *Validator {
	if config == nil {
		config = DefaultValidationConfig()
	}
	return &Validator{config: config}
}

func (v *Validator) ValidateText(text string, fieldName string) *errors.ValidationError {
	if !v.config.AllowEmptyStrings && strings.TrimSpace(text) == "" {
		return errors.NewValidationError(fieldName, "cannot be empty", text)
	}

	if !utf8.ValidString(text) {
		return errors.NewValidationError(fieldName, "must be valid UTF-8", text)
	}

	if utf8.RuneCountInString(text) > v.config.MaxInputLength {
		return errors.NewValidationError(fieldName,
			"exceeds maximum length", map[string]any{
				"length":     utf8.RuneCountInString(text),
				"max_length": v.config.MaxInputLength,
			})
	}

	return nil
}

func (v *Validator) ValidateTexts(texts []string, fieldName string) *errors.MultiValidationError {
	validationErr := &errors.MultiValidationError{}

	if len(texts) == 0 {
		validationErr.Add(fieldName, "cannot be empty", len(texts))
		return validationErr
	}

	if len(texts) > v.config.MaxBatchSize {
		validationErr.Add(fieldName, "exceeds maximum batch size", map[string]interface{}{
			"size":     len(texts),
			"max_size": v.config.MaxBatchSize,
		})
	}

	for i, text := range texts {
		if err := v.ValidateText(text, fieldName+"["+string(rune(i))+"]"); err != nil {
			validationErr.Add(err.Field, err.Message, err.Value)
		}
	}

	if validationErr.HasErrors() {
		return validationErr
	}
	return nil
}

func (v *Validator) ValidatePromptName(promptName *string) *errors.ValidationError {
	if promptName == nil {
		return nil
	}

	if strings.TrimSpace(*promptName) == "" {
		return errors.NewValidationError("prompt_name", "cannot be empty", *promptName)
	}

	validPromptName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPromptName.MatchString(*promptName) {
		return errors.NewValidationError("prompt_name",
			"must contain only letters, numbers, underscores, and hyphens", *promptName)
	}

	return nil
}

func (v *Validator) ValidateEncodingFormat(format EncodingFormat) *errors.ValidationError {
	if format == "" {
		return nil
	}

	switch format {
	case EncodingFloat, EncodingBase64:
		return nil
	default:
		return errors.NewValidationError("encoding_format",
			"must be 'float' or 'base64'", string(format))
	}
}

func (v *Validator) ValidateTruncationDirection(direction TruncationDirection) *errors.ValidationError {
	if direction == "" {
		return nil
	}

	switch direction {
	case TruncationLeft, TruncationRight:
		return nil
	default:
		return errors.NewValidationError("truncation_direction",
			"must be 'Left' or 'Right'", string(direction))
	}
}

func (v *Validator) ValidateEmbedRequest(req *EmbedRequest) error {
	if err := v.ValidateTexts(req.Inputs.Data, "inputs"); err != nil {
		return err
	}

	if err := v.ValidatePromptName(req.PromptName); err != nil {
		return err
	}

	if err := v.ValidateTruncationDirection(req.TruncationDirection); err != nil {
		return err
	}

	return nil
}

func (v *Validator) ValidateSimilarityRequest(req *SimilarityRequest) error {
	if err := v.ValidateText(req.Inputs.SourceSentence, "source_sentence"); err != nil {
		return err
	}

	if len(req.Inputs.Sentences) > v.config.MaxSentencesCount {
		return errors.NewValidationError("sentences", "exceeds maximum sentences count",
			map[string]any{
				"count":     len(req.Inputs.Sentences),
				"max_count": v.config.MaxSentencesCount,
			})
	}

	if err := v.ValidateTexts(req.Inputs.Sentences, "sentences"); err != nil {
		return err
	}

	if req.Parameters != nil {
		if err := v.ValidatePromptName(req.Parameters.PromptName); err != nil {
			return err
		}

		if err := v.ValidateTruncationDirection(req.Parameters.TruncationDirection); err != nil {
			return err
		}
	}

	return nil
}
