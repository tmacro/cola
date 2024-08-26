package templates

import "embed"

//go:embed templates/*.tpl
var _templates embed.FS

func mustGetEmbeddedFile(name string) string {
	content, err := _templates.ReadFile("templates/" + name)
	if err != nil {
		panic(err)
	}
	return string(content)
}
