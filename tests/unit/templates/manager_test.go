package templates_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jasongoecke/go-veo3/pkg/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()

	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, manager)

	// Should create empty manager if no templates file exists
	list := manager.List()
	assert.Empty(t, list)
}

func TestManager_SaveAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	template := &templates.Template{
		Name:        "test-template",
		Prompt:      "Generate a {{style}} image of {{subject}}",
		Description: "Test template",
		Tags:        []string{"test", "example"},
	}

	// Save template
	err = manager.Save(template)
	require.NoError(t, err)
	assert.False(t, template.CreatedAt.IsZero())
	assert.False(t, template.UpdatedAt.IsZero())

	// Get template
	retrieved, err := manager.Get("test-template")
	require.NoError(t, err)
	assert.Equal(t, template.Name, retrieved.Name)
	assert.Equal(t, template.Prompt, retrieved.Prompt)
	assert.Equal(t, template.Description, retrieved.Description)
	assert.Equal(t, template.Tags, retrieved.Tags)
}

func TestManager_SaveEmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	template := &templates.Template{
		Name:   "",
		Prompt: "Test prompt",
	}

	err = manager.Save(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestManager_SaveInvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	template := &templates.Template{
		Name:   "invalid",
		Prompt: "{{unclosed variable",
	}

	err = manager.Save(template)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid template")
}

func TestManager_GetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	_, err = manager.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_List(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Save multiple templates
	templates := []*templates.Template{
		{Name: "template1", Prompt: "Prompt 1"},
		{Name: "template2", Prompt: "Prompt 2"},
		{Name: "template3", Prompt: "Prompt 3"},
	}

	for _, tmpl := range templates {
		err := manager.Save(tmpl)
		require.NoError(t, err)
	}

	// List should return all templates
	list := manager.List()
	assert.Len(t, list, 3)

	// Check all templates are present
	names := make(map[string]bool)
	for _, tmpl := range list {
		names[tmpl.Name] = true
	}
	assert.True(t, names["template1"])
	assert.True(t, names["template2"])
	assert.True(t, names["template3"])
}

func TestManager_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Save template
	template := &templates.Template{
		Name:   "delete-me",
		Prompt: "Test prompt",
	}
	err = manager.Save(template)
	require.NoError(t, err)

	// Delete template
	err = manager.Delete("delete-me")
	require.NoError(t, err)

	// Verify deletion
	_, err = manager.Get("delete-me")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_DeleteNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	err = manager.Delete("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_Export(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Save templates
	template1 := &templates.Template{
		Name:        "template1",
		Prompt:      "Prompt 1",
		Description: "Description 1",
		Tags:        []string{"tag1"},
	}
	template2 := &templates.Template{
		Name:   "template2",
		Prompt: "Prompt 2",
		Tags:   []string{"tag2", "tag3"},
	}

	err = manager.Save(template1)
	require.NoError(t, err)
	err = manager.Save(template2)
	require.NoError(t, err)

	// Export
	exportPath := filepath.Join(tmpDir, "export.yaml")
	err = manager.Export(exportPath)
	require.NoError(t, err)

	// Verify file exists and contains data
	data, err := os.ReadFile(exportPath) // #nosec G304 -- test file in temp directory
	require.NoError(t, err)
	assert.Contains(t, string(data), "template1")
	assert.Contains(t, string(data), "template2")
	assert.Contains(t, string(data), "Prompt 1")
	assert.Contains(t, string(data), "Prompt 2")
}

func TestManager_Import(t *testing.T) {
	tmpDir := t.TempDir()
	manager1, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Save templates in first manager
	template1 := &templates.Template{
		Name:   "imported1",
		Prompt: "Imported prompt 1",
		Tags:   []string{"import"},
	}
	template2 := &templates.Template{
		Name:   "imported2",
		Prompt: "Imported prompt 2",
	}

	err = manager1.Save(template1)
	require.NoError(t, err)
	err = manager1.Save(template2)
	require.NoError(t, err)

	// Export from first manager
	exportPath := filepath.Join(tmpDir, "import-test.yaml")
	err = manager1.Export(exportPath)
	require.NoError(t, err)

	// Create second manager in different directory
	tmpDir2 := t.TempDir()
	manager2, err := templates.NewManager(tmpDir2)
	require.NoError(t, err)

	// Import into second manager
	err = manager2.Import(exportPath)
	require.NoError(t, err)

	// Verify templates were imported
	list := manager2.List()
	assert.Len(t, list, 2)

	retrieved1, err := manager2.Get("imported1")
	require.NoError(t, err)
	assert.Equal(t, "Imported prompt 1", retrieved1.Prompt)

	retrieved2, err := manager2.Get("imported2")
	require.NoError(t, err)
	assert.Equal(t, "Imported prompt 2", retrieved2.Prompt)
}

func TestManager_ImportInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Try to import non-existent file
	err = manager.Import(filepath.Join(tmpDir, "nonexistent.yaml"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}

func TestManager_ImportInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Create invalid YAML file
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	err = os.WriteFile(invalidPath, []byte("not: valid: yaml:"), 0600)
	require.NoError(t, err)

	err = manager.Import(invalidPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestManager_ImportInvalidTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Create YAML with invalid template
	invalidPath := filepath.Join(tmpDir, "invalid-template.yaml")
	content := `templates:
  - name: invalid
    prompt: "{{unclosed variable"
`
	err = os.WriteFile(invalidPath, []byte(content), 0600)
	require.NoError(t, err)

	err = manager.Import(invalidPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid template")
}

func TestManager_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manager and save template
	manager1, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	template := &templates.Template{
		Name:   "persistent",
		Prompt: "Persistent prompt",
	}
	err = manager1.Save(template)
	require.NoError(t, err)

	// Create new manager in same directory
	manager2, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Should load existing template
	retrieved, err := manager2.Get("persistent")
	require.NoError(t, err)
	assert.Equal(t, "Persistent prompt", retrieved.Prompt)
}

func TestManager_Update(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := templates.NewManager(tmpDir)
	require.NoError(t, err)

	// Save initial template
	template := &templates.Template{
		Name:   "updatable",
		Prompt: "Original prompt",
	}
	err = manager.Save(template)
	require.NoError(t, err)

	createdAt := template.CreatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	// Update template
	updated := &templates.Template{
		Name:   "updatable",
		Prompt: "Updated prompt",
	}
	err = manager.Save(updated)
	require.NoError(t, err)

	// Verify update
	retrieved, err := manager.Get("updatable")
	require.NoError(t, err)
	assert.Equal(t, "Updated prompt", retrieved.Prompt)
	assert.True(t, retrieved.UpdatedAt.After(createdAt))
}

func TestTemplate_Variables(t *testing.T) {
	template := &templates.Template{
		Name:   "test",
		Prompt: "A {{style}} image of {{subject}} in {{location}}",
	}

	vars := template.Variables()
	assert.Len(t, vars, 3)
	assert.Contains(t, vars, "style")
	assert.Contains(t, vars, "subject")
	assert.Contains(t, vars, "location")
}

func TestTemplate_Render(t *testing.T) {
	template := &templates.Template{
		Name:   "test",
		Prompt: "A {{style}} image of {{subject}}",
	}

	result, err := template.Render(map[string]string{
		"style":   "cinematic",
		"subject": "mountain landscape",
	})
	require.NoError(t, err)
	assert.Equal(t, "A cinematic image of mountain landscape", result)
}

func TestTemplate_RenderMissingVariable(t *testing.T) {
	template := &templates.Template{
		Name:   "test",
		Prompt: "A {{style}} image of {{subject}}",
	}

	_, err := template.Render(map[string]string{
		"style": "cinematic",
		// missing "subject"
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required variable")
}
