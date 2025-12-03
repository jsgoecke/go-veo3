package templates_test

import (
	"testing"

	"github.com/jasongoecke/go-veo3/pkg/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractVariables(t *testing.T) {
	tests := []struct {
		name     string
		template string
		want     []string
	}{
		{
			name:     "no variables",
			template: "A simple prompt with no variables",
			want:     []string{},
		},
		{
			name:     "single variable",
			template: "A {{subject}} in the park",
			want:     []string{"subject"},
		},
		{
			name:     "multiple variables",
			template: "A {{subject}} {{action}} in the {{location}}",
			want:     []string{"subject", "action", "location"},
		},
		{
			name:     "duplicate variables",
			template: "A {{color}} car and a {{color}} house",
			want:     []string{"color"},
		},
		{
			name:     "variables with underscores",
			template: "A {{main_subject}} with {{background_color}}",
			want:     []string{"main_subject", "background_color"},
		},
		{
			name:     "variables with numbers",
			template: "{{subject1}} and {{subject2}}",
			want:     []string{"subject1", "subject2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := templates.ExtractVariables(tt.template)
			assert.ElementsMatch(t, tt.want, vars)
		})
	}
}

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		variables map[string]string
		want      string
		wantErr   bool
	}{
		{
			name:      "no variables",
			template:  "A simple prompt",
			variables: map[string]string{},
			want:      "A simple prompt",
			wantErr:   false,
		},
		{
			name:     "single variable substitution",
			template: "A {{subject}} in the park",
			variables: map[string]string{
				"subject": "dog",
			},
			want:    "A dog in the park",
			wantErr: false,
		},
		{
			name:     "multiple variable substitution",
			template: "A {{subject}} {{action}} in the {{location}}",
			variables: map[string]string{
				"subject":  "cat",
				"action":   "playing",
				"location": "garden",
			},
			want:    "A cat playing in the garden",
			wantErr: false,
		},
		{
			name:     "missing variable",
			template: "A {{subject}} in the {{location}}",
			variables: map[string]string{
				"subject": "bird",
			},
			want:    "",
			wantErr: true,
		},
		{
			name:     "extra variables ignored",
			template: "A {{subject}} in the park",
			variables: map[string]string{
				"subject":  "dog",
				"location": "beach", // Extra, should be ignored
			},
			want:    "A dog in the park",
			wantErr: false,
		},
		{
			name:     "duplicate variables",
			template: "A {{color}} car and a {{color}} house",
			variables: map[string]string{
				"color": "red",
			},
			want:    "A red car and a red house",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := templates.SubstituteVariables(tt.template, tt.variables)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "valid template",
			template: "A {{subject}} in the {{location}}",
			wantErr:  false,
		},
		{
			name:     "empty template",
			template: "",
			wantErr:  true,
		},
		{
			name:     "template with unclosed variable",
			template: "A {{subject in the park",
			wantErr:  true,
		},
		{
			name:     "template with unopened variable",
			template: "A subject}} in the park",
			wantErr:  true,
		},
		{
			name:     "template with nested braces",
			template: "A {{sub{{ject}}}} in the park",
			wantErr:  true,
		},
		{
			name:     "template with empty variable name",
			template: "A {{}} in the park",
			wantErr:  true,
		},
		{
			name:     "template with only spaces in variable",
			template: "A {{   }} in the park",
			wantErr:  true,
		},
		{
			name:     "valid template with no variables",
			template: "A simple prompt with no variables",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := templates.ValidateTemplate(tt.template)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
