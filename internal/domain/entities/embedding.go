package entities

import (
	"encoding/json"
	"fmt"
	"strings"

	"teiwrappergolang/internal/domain/errors"
)

type TruncationDirection string

const (
	TruncationLeft  TruncationDirection = "Left"
	TruncationRight TruncationDirection = "Right"
)

type EncodingFormat string

const (
	EncodingFloat  EncodingFormat = "float"
	EncodingBase64 EncodingFormat = "base64"
)

type InputType any

type Input struct {
	Data []string `json:"data"`
}

func (i *Input) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		i.Data = arr
		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		i.Data = []string{str}
		return nil
	}

	return fmt.Errorf("input must be a string or array of strings")
}

func (i Input) MarshalJSON() ([]byte, error) {
	if len(i.Data) == 1 {
		return json.Marshal(i.Data[0])
	}
	return json.Marshal(i.Data)
}

func (i *Input) Validate() *errors.MultiValidationError {
	validationErr := &errors.MultiValidationError{}

	if len(i.Data) == 0 {
		validationErr.Add("input", "input cannot be empty", nil)
		return validationErr
	}

	for idx, text := range i.Data {
		if strings.TrimSpace(text) == "" {
			validationErr.Add("input", fmt.Sprintf("input[%d] cannot be empty", idx), text)
		}
	}

	if validationErr.HasErrors() {
		return validationErr
	}
	return nil
}

type EmbedRequest struct {
	Inputs              Input               `json:"inputs" validate:"required"`
	Normalize           *bool               `json:"normalize,omitempty"`
	PromptName          *string             `json:"prompt_name,omitempty"`
	Truncate            *bool               `json:"truncate,omitempty"`
	TruncationDirection TruncationDirection `json:"truncation_direction,omitempty"`
}

func (r *EmbedRequest) Validate() error {
	if validationErr := r.Inputs.Validate(); validationErr != nil {
		return validationErr
	}
	return nil
}

type EmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

type EmbedAllRequest struct {
	Inputs              Input               `json:"inputs" validate:"required"`
	PromptName          *string             `json:"prompt_name,omitempty"`
	Truncate            *bool               `json:"truncate,omitempty"`
	TruncationDirection TruncationDirection `json:"truncation_direction,omitempty"`
}

func (r *EmbedAllRequest) Validate() error {
	if validationErr := r.Inputs.Validate(); validationErr != nil {
		return validationErr
	}
	return nil
}

type EmbedAllResponse struct {
	Embeddings [][][]float32 `json:"embeddings"`
}

type SparseValue struct {
	Index int     `json:"index"`
	Value float32 `json:"value"`
}

type EmbedSparseRequest struct {
	Inputs              Input               `json:"inputs" validate:"required"`
	PromptName          *string             `json:"prompt_name,omitempty"`
	Truncate            *bool               `json:"truncate,omitempty"`
	TruncationDirection TruncationDirection `json:"truncation_direction,omitempty"`
}

func (r *EmbedSparseRequest) Validate() error {
	if validationErr := r.Inputs.Validate(); validationErr != nil {
		return validationErr
	}
	return nil
}

type EmbedSparseResponse struct {
	Embeddings [][]SparseValue `json:"embeddings"`
}

func BoolPtr(b bool) *bool {
	return &b
}

func StringPtr(s string) *string {
	return &s
}

func (r *EmbedRequest) SetDefaults() {
	if r.Normalize == nil {
		r.Normalize = BoolPtr(true)
	}
	if r.Truncate == nil {
		r.Truncate = BoolPtr(false)
	}
	if r.TruncationDirection == "" {
		r.TruncationDirection = TruncationRight
	}
}

func (r *EmbedAllRequest) SetDefaults() {
	if r.Truncate == nil {
		r.Truncate = BoolPtr(false)
	}
	if r.TruncationDirection == "" {
		r.TruncationDirection = TruncationRight
	}
}

func (r *EmbedSparseRequest) SetDefaults() {
	if r.Truncate == nil {
		r.Truncate = BoolPtr(false)
	}
	if r.TruncationDirection == "" {
		r.TruncationDirection = TruncationRight
	}
}
