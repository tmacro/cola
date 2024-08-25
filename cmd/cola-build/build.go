package main

// import "os"

// func createWorkDir() (func() error, string) {
// 	dir, err := os.MkdirTemp("", "cola-build-*")
// 	if err != nil {
// 		panic(err)
// 	}

// 	return func() error {
// 		return os.RemoveAll(dir)
// 	}, dir
// }

// func buildExtension() {
// 	cleanup, workDir := createWorkDir()
// 	defer cleanup()

// 	// Build the extension in the work directory.
// }
