package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Template represents a reusable prompt template
type Template struct {
	Name        string    `yaml:"name" json:"name"`
	Prompt      string    `yaml:"prompt" json:"prompt"`
	Description string    `yaml:"description,omitempty" json:"description,omitempty"`
	Tags        []string  `yaml:"tags,omitempty" json:"tags,omitempty"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
}

// Manager handles template storage and retrieval
type Manager struct {
	templatesPath string
	templates     map[string]*Template
}

// TemplateCollection represents the YAML file structure
type TemplateCollection struct {
	Templates []Template `yaml:"templates"`
}

// NewManager creates a new template manager
func NewManager(configDir string) (*Manager, error) {
	templatesPath := filepath.Join(configDir, "templates.yaml")

	m := &Manager{
		templatesPath: templatesPath,
		templates:     make(map[string]*Template),
	}

	// Load existing templates if file exists
	if _, err := os.Stat(templatesPath); err == nil {
		if err := m.load(); err != nil {
			return nil, fmt.Errorf("failed to load templates: %w", err)
		}
	}

	return m, nil
}

// Save saves a template
func (m *Manager) Save(template *Template) error {
	if template.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if err := ValidateTemplate(template.Prompt); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	now := time.Now()
	if template.CreatedAt.IsZero() {
		template.CreatedAt = now
	}
	template.UpdatedAt = now

	m.templates[template.Name] = template

	return m.save()
}

// Get retrieves a template by name
func (m *Manager) Get(name string) (*Template, error) {
	template, exists := m.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return template, nil
}

// List returns all templates
func (m *Manager) List() []*Template {
	templates := make([]*Template, 0, len(m.templates))
	for _, t := range m.templates {
		templates = append(templates, t)
	}
	return templates
}

// Delete removes a template
func (m *Manager) Delete(name string) error {
	if _, exists := m.templates[name]; !exists {
		return fmt.Errorf("template not found: %s", name)
	}

	delete(m.templates, name)
	return m.save()
}

// Export exports all templates to a file
func (m *Manager) Export(path string) error {
	collection := TemplateCollection{
		Templates: make([]Template, 0, len(m.templates)),
	}

	for _, t := range m.templates {
		collection.Templates = append(collection.Templates, *t)
	}

	data, err := yaml.Marshal(collection)
	if err != nil {
		return fmt.Errorf("failed to marshal templates: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

// Import imports templates from a file
func (m *Manager) Import(path string) error {
	data, err := os.ReadFile(path) // #nosec G304 -- path is user-provided CLI argument
	if err != nil {
		return fmt.Errorf("failed to read templates file: %w", err)
	}

	var collection TemplateCollection
	if err := yaml.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to parse templates file: %w", err)
	}

	// Validate and import each template
	for i, template := range collection.Templates {
		if err := ValidateTemplate(template.Prompt); err != nil {
			return fmt.Errorf("invalid template at index %d (%s): %w", i, template.Name, err)
		}
		m.templates[template.Name] = &collection.Templates[i]
	}

	return m.save()
}

// load loads templates from the YAML file
func (m *Manager) load() error {
	data, err := os.ReadFile(m.templatesPath)
	if err != nil {
		return err
	}

	var collection TemplateCollection
	if err := yaml.Unmarshal(data, &collection); err != nil {
		return err
	}

	for i := range collection.Templates {
		t := &collection.Templates[i]
		m.templates[t.Name] = t
	}

	return nil
}

// save saves templates to the YAML file
func (m *Manager) save() error {
	collection := TemplateCollection{
		Templates: make([]Template, 0, len(m.templates)),
	}

	for _, t := range m.templates {
		collection.Templates = append(collection.Templates, *t)
	}

	data, err := yaml.Marshal(collection)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(m.templatesPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(m.templatesPath, data, 0600)
}

// Variables returns the variables required by a template
func (t *Template) Variables() []string {
	return ExtractVariables(t.Prompt)
}

// Render renders the template with provided variables
func (t *Template) Render(variables map[string]string) (string, error) {
	return SubstituteVariables(t.Prompt, variables)
}
