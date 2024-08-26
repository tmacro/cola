package files

import "embed"

//go:embed files/*
var _files embed.FS

func MustGetEmbeddedFile(name string) string {
	content, err := _files.ReadFile("files/" + name)
	if err != nil {
		panic(err)
	}
	return string(content)
}
