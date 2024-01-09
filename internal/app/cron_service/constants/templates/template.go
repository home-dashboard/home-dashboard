package templates

import (
	_ "embed"
	"fmt"
	"github.com/siaikin/home-dashboard/internal/pkg/comfy_log"
	"io"
	"text/template"
)

var logger = comfy_log.New("[cron_service templates]")

//go:embed .gitattributes.template
var GitAttributesTemplate string

//go:embed .gitignore.template
var GitIgnoreTemplate string

//go:embed package.json.template
var PackageJsonTemplate string

//go:embed README.md.template
var ReadmeMdTemplate string

var Templates = template.New("templates")

func init() {
	if _, err := Templates.New(".gitattributes").Parse(GitAttributesTemplate); err != nil {
		logger.Fatal("parse GitAttributesTemplate failed: %s\n", err.Error())
	}
	if _, err := Templates.New(".gitignore").Parse(GitIgnoreTemplate); err != nil {
		logger.Fatal("parse GitIgnoreTemplate failed: %s\n", err.Error())
	}
	if _, err := Templates.New("package.json").Parse(PackageJsonTemplate); err != nil {
		logger.Fatal("parse PackageJsonTemplate failed: %s\n", err.Error())
	}
	if _, err := Templates.New("README.md").Parse(ReadmeMdTemplate); err != nil {
		logger.Fatal("parse ReadmeMdTemplate failed: %s\n", err.Error())
	}
}

func ExecuteTemplate(templateName string, data any, w io.Writer) error {
	switch templateName {
	case ".gitattributes":
		return Templates.ExecuteTemplate(w, ".gitattributes", data)
	case ".gitignore":
		return Templates.ExecuteTemplate(w, ".gitignore", data)
	case "package.json":
		return Templates.ExecuteTemplate(w, "package.json", data)
	case "README.md":
		return Templates.ExecuteTemplate(w, "README.md", data)
	default:
		return fmt.Errorf("unknown template name: %s", templateName)
	}
}

type PackageJSONData struct {
	Name            string
	Version         string
	Description     string
	Main            string
	Scripts         map[string]string
	Dependencies    map[string]string
	DevDependencies map[string]string
}
