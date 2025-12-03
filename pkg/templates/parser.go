package templates

import (
	"fmt"
	"regexp"
	"strings"
)

// ExtractVariables extracts variable names from a template string
// Variables are in the format {{variable_name}}
func ExtractVariables(template string) []string {
	re := regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	// Use map to deduplicate
	vars := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			vars[match[1]] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(vars))
	for v := range vars {
		result = append(result, v)
	}

	return result
}

// SubstituteVariables replaces variables in a template with provided values
func SubstituteVariables(template string, variables map[string]string) (string, error) {
	// Check if all required variables are provided
	requiredVars := ExtractVariables(template)
	for _, v := range requiredVars {
		if _, ok := variables[v]; !ok {
			return "", fmt.Errorf("missing required variable: %s", v)
		}
	}

	// Perform substitution
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// ValidateTemplate validates a template string
func ValidateTemplate(template string) error {
	if template == "" {
		return fmt.Errorf("template cannot be empty")
	}

	// Check for balanced braces
	openCount := strings.Count(template, "{{")
	closeCount := strings.Count(template, "}}")

	if openCount != closeCount {
		return fmt.Errorf("unbalanced braces in template")
	}

	// Check for nested braces
	if strings.Contains(template, "{{{") || strings.Contains(template, "}}}") {
		return fmt.Errorf("nested braces not allowed in template")
	}

	// Extract and validate variable names
	re := regexp.MustCompile(`\{\{([^}]*)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			if varName == "" {
				return fmt.Errorf("empty variable name in template")
			}

			// Check if variable name contains only valid characters
			validName := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
			if !validName.MatchString(varName) {
				return fmt.Errorf("invalid variable name: %s (use only letters, numbers, and underscores)", match[1])
			}
		}
	}

	return nil
}
