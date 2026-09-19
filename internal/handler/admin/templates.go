package admin

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"presensee/internal/model"
)

//go:embed templates/*.html
var templateFS embed.FS

// Engine handles template parsing, layout wrapping, and rendering.
type Engine struct {
	templates map[string]*template.Template
}

var defaultEngine *Engine

func init() {
	defaultEngine = NewEngine()
}

// NewEngine initializes and parses all admin templates with the AdminLTE layout.
func NewEngine() *Engine {
	eng := &Engine{
		templates: make(map[string]*template.Template),
	}

	funcMap := template.FuncMap{
		"containsUint": func(slice []uint, val uint) bool {
			for _, item := range slice {
				if item == val {
					return true
				}
			}
			return false
		},
		"derefUserType": func(t any) string {
			if t == nil {
				return "-"
			}
			if str, ok := t.(*model.UserType); ok && str != nil {
				return string(*str)
			}
			if str, ok := t.(model.UserType); ok {
				return string(str)
			}
			return fmt.Sprint(t)
		},
		"derefUint": func(u *uint) uint {
			if u == nil {
				return 0
			}
			return *u
		},
		"hasPrefix": strings.HasPrefix,
	}

	layoutContent, err := templateFS.ReadFile("templates/layout.html")
	if err != nil {
		panic(fmt.Sprintf("failed to read layout.html: %v", err))
	}

	// List of standalone pages that do NOT use the main sidebar layout
	standalone := map[string]bool{
		"login.html":       true,
		"kartu_siswa.html": true,
	}

	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		panic(fmt.Sprintf("failed to read templates dir: %v", err))
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") || entry.Name() == "layout.html" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".html")
		content, err := templateFS.ReadFile(filepath.Join("templates", entry.Name()))
		if err != nil {
			panic(fmt.Sprintf("failed to read template %s: %v", entry.Name(), err))
		}

		var tmpl *template.Template
		if standalone[entry.Name()] {
			tmpl = template.Must(template.New(entry.Name()).Funcs(funcMap).Parse(string(content)))
		} else {
			// Parse layout then override content block
			tmpl = template.Must(template.New("layout.html").Funcs(funcMap).Parse(string(layoutContent)))
			tmpl = template.Must(tmpl.Parse(string(content)))
		}

		eng.templates[name] = tmpl
	}

	return eng
}

// Render renders an HTML template by name to the ResponseWriter.
func (e *Engine) Render(w http.ResponseWriter, name string, data any) {
	tmpl, ok := e.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("Template '%s' not found", name), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
	}
}

func renderTemplate(w http.ResponseWriter, name string, data any) {
	defaultEngine.Render(w, name, data)
}
