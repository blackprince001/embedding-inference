package entities

import (
	"strings"

	"github.com/blackprince001/embedding-inference/internal/domain/errors"
)

type SimilarityInput struct {
	SourceSentence string   `json:"source_sentence" validate:"required"`
	Sentences      []string `json:"sentences" validate:"required,min=1"`
}

func (s *SimilarityInput) Validate() error {
	validationErr := &errors.MultiValidationError{}

	if strings.TrimSpace(s.SourceSentence) == "" {
		validationErr.Add("source_sentence", "source sentence cannot be empty", s.SourceSentence)
	}

	if len(s.Sentences) == 0 {
		validationErr.Add("sentences", "sentences array cannot be empty", len(s.Sentences))
	} else {
		for idx, sentence := range s.Sentences {
			if strings.TrimSpace(sentence) == "" {
				validationErr.Add("sentences",
					"sentence at index "+string(rune(idx))+" cannot be empty", sentence)
			}
		}
	}

	if validationErr.HasErrors() {
		return validationErr
	}
	return nil
}

type SimilarityParameters struct {
	PromptName          *string             `json:"prompt_name,omitempty"`
	Truncate            *bool               `json:"truncate,omitempty"`
	TruncationDirection TruncationDirection `json:"truncation_direction,omitempty"`
}

func (p *SimilarityParameters) SetDefaults() {
	if p.Truncate == nil {
		p.Truncate = BoolPtr(false)
	}
	if p.TruncationDirection == "" {
		p.TruncationDirection = TruncationRight
	}
}

type SimilarityRequest struct {
	Inputs     SimilarityInput       `json:"inputs" validate:"required"`
	Parameters *SimilarityParameters `json:"parameters,omitempty"`
}

func (r *SimilarityRequest) Validate() error {
	return r.Inputs.Validate()
}

func (r *SimilarityRequest) SetDefaults() {
	if r.Parameters == nil {
		r.Parameters = &SimilarityParameters{}
	}
	r.Parameters.SetDefaults()
}

type SimilarityResponse struct {
	Similarities []float32 `json:"similarities"`
}
