package main

import "embed"

//go:embed files/*
var file embed.FS

func getEmbeddedFile(name string) (string, error) {
	content, err := file.ReadFile("files/" + name)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func mustGetEmbeddedFile(name string) string {
	content, err := getEmbeddedFile(name)
	if err != nil {
		panic(err)
	}
	return content
}
