# Template-Based Filename Generation

This document explains how to use the template-based filename generation system in goname.

## Overview

The `FileService` now supports configurable filename templates using Go's `text/template` syntax. This allows users to customize how movie and TV show files are renamed according to their preferences.

## Available Template Fields

The following fields are available for use in templates:

| Field | Description | Type | Example |
|-------|-------------|------|---------|
| `{{.Name}}` | Movie title or TV show name | string | "Breaking Bad" |
| `{{.Title}}` | Episode title (TV shows only) | string | "Pilot" |
| `{{.Year}}` | Release year | int | 2008 |
| `{{.Season}}` | Season number (TV shows only) | int | 1 |
| `{{.Episode}}` | Episode number (TV shows only) | int | 5 |
| `{{.Director}}` | Director name (movies only) | string | "Vince Gilligan" |
| `{{.Genre}}` | Genre | string | "Drama" |

**Note:** File extensions are automatically preserved and should not be included in templates.

## Predefined Templates

### TV Show Templates

- **Default**: `{{.Name}} - S{{.Season}}E{{.Episode}} - {{.Title}}`
  - Output: `Breaking Bad - S01E01 - Pilot.mkv` (extension preserved from original)

- **Simple**: `{{.Name}} S{{.Season}}E{{.Episode}}`
  - Output: `Breaking Bad S01E01.mkv` (extension preserved from original)

- **With Year**: `{{.Name}} ({{.Year}}) - S{{.Season}}E{{.Episode}} - {{.Title}}`
  - Output: `Breaking Bad (2008) - S01E01 - Pilot.mkv` (extension preserved from original)

### Movie Templates

- **Default**: `{{.Name}} ({{.Year}})`
  - Output: `The Shawshank Redemption (1994).mp4` (extension preserved from original)

- **Simple**: `{{.Name}}`
  - Output: `The Shawshank Redemption.mp4` (extension preserved from original)

- **With Genre**: `{{.Name}} ({{.Year}}) - {{.Genre}}`
  - Output: `The Shawshank Redemption (1994) - Drama.mp4` (extension preserved from original)

## Usage Examples

### Basic Usage

```go
// Create a new file service with default templates
fs := services.NewFileService()

// Generate filename using default template
filename := fs.GenerateTVShowFileName(show, episode, "original.mkv")
```

### Custom Templates

```go
// Create file service with custom templates
fs := services.NewFileServiceWithTemplates(
    "{{.Name}} {{.Season}}x{{.Episode}} {{.Title}}", // TV template
    "{{.Name}} [{{.Year}}]"                          // Movie template
)

// Or set templates on existing service
fs.SetTVShowTemplate("{{.Name}} - {{.Season}}x{{.Episode}}")
fs.SetMovieTemplate("{{.Name}} ({{.Year}})")
```

### Template Validation

```go
template := "{{.Name}} - S{{.Season}}E{{.Episode}}"
if err := fs.ValidateTemplate(template); err != nil {
    fmt.Printf("Invalid template: %v\n", err)
}
```

### Preview Template Output

```go
// Preview what a template would generate
output, err := fs.PreviewTemplateOutput("{{.Name}} S{{.Season}}E{{.Episode}}", false)
if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Preview: %s\n", output) // "Sample TV Show S01E05" (extension added automatically)
}
```

### Get Available Presets

```go
presets := fs.GetTemplatePresets()
for name, template := range presets {
    fmt.Printf("%s: %s\n", name, template)
}
```

## Template Formatting Functions

The template system automatically handles:

- **Filename sanitization**: Removes invalid characters for filenames
- **Number padding**: Season and episode numbers are padded with zeros (e.g., `S01E05`)
- **Error fallback**: If template parsing fails, falls back to sensible defaults

## Advanced Template Examples

### Using Conditional Logic

```go
// Template that shows year only if available
template := "{{.Name}}{{if .Year}} ({{.Year}}){{end}}"
```

### Multiple Field Formatting

```go
// Template with multiple optional fields
template := "{{.Name}}{{if .Year}} ({{.Year}}){{end}}{{if .Genre}} [{{.Genre}}]{{end}}"
```

### Custom Number Formatting

```go
// Template with custom season/episode format
template := "{{.Name}} Season {{.Season}} Episode {{.Episode}}{{if .Title}} - {{.Title}}{{end}}"
```

## Error Handling

The template system includes robust error handling:

1. **Template Parsing Errors**: Invalid Go template syntax will be caught during validation
2. **Template Execution Errors**: If a template references missing data, it falls back to default formats
3. **Filename Sanitization**: All generated filenames are automatically sanitized for the target filesystem

## Configuration Integration

Templates can be stored in configuration files and loaded at runtime:

```yaml
templates:
  tv_show: "{{.Name}} - S{{.Season}}E{{.Episode}} - {{.Title}}"
  movie: "{{.Name}} ({{.Year}})"
```

This allows users to customize filename formats without modifying code. Extensions are automatically preserved from the original files.
