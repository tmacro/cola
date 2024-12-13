package ignition

import (
	"fmt"
	"os"
	"strconv"

	ignitionTypes "github.com/coreos/ignition/v2/config/v3_4/types"
	"github.com/tmacro/cola/pkg/config"
)

func parseOctal(s string) (int, error) {
	val, err := strconv.ParseInt(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return int(val), nil
}

func generateFiles(cfg *config.ApplianceConfig, g *generator) error {
	for _, file := range cfg.Files {
		mode, err := parseOctal(file.Mode)
		if err != nil {
			return fmt.Errorf("failed to parse file mode %s: %v", file.Mode, err)
		}

		f := ignitionTypes.File{
			Node: ignitionTypes.Node{
				Path: file.Path,
			},
			FileEmbedded1: ignitionTypes.FileEmbedded1{
				Mode: toPtr(mode),
			},
		}

		if file.Owner != "" {
			f.Node.User = ignitionTypes.NodeUser{
				Name: toPtr(file.Owner),
			}
		}

		if file.Group != "" {
			f.Node.Group = ignitionTypes.NodeGroup{
				Name: toPtr(file.Group),
			}
		}

		if file.Overwrite {
			f.Overwrite = toPtr(file.Overwrite)
		}

		if file.Inline != "" {
			f.FileEmbedded1.Contents.Source = toPtr(toDataUrl(file.Inline))
		} else if file.SourcePath != "" {
			content, err := os.ReadFile(file.SourcePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", file.SourcePath, err)
			}

			f.FileEmbedded1.Contents.Source = toPtr(toDataUrl(string(content)))
		} else if file.URL != "" {
			f.FileEmbedded1.Contents.Source = toPtr(file.URL)
		}

		g.Files = append(g.Files, f)
	}

	return nil
}

func validateFiles(g *generator) error {
	seenPaths := make(map[string]struct{})
	for _, file := range g.Files {
		if _, ok := seenPaths[file.Path]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateFile, file.Path)
		}
		seenPaths[file.Path] = struct{}{}
	}

	return nil
}

func generateDirectories(cfg *config.ApplianceConfig, g *generator) error {
	for _, dir := range cfg.Directories {
		mode, err := parseOctal(dir.Mode)
		if err != nil {
			return fmt.Errorf("failed to parse file mode %s: %v", dir.Mode, err)
		}

		d := ignitionTypes.Directory{
			Node: ignitionTypes.Node{
				Path: dir.Path,
			},
			DirectoryEmbedded1: ignitionTypes.DirectoryEmbedded1{
				Mode: toPtr(mode),
			},
		}

		if dir.Owner != "" {
			d.Node.User = ignitionTypes.NodeUser{
				Name: toPtr(dir.Owner),
			}
		}

		if dir.Group != "" {
			d.Node.Group = ignitionTypes.NodeGroup{
				Name: toPtr(dir.Group),
			}
		}

		g.Directories = append(g.Directories, d)
	}

	return nil
}

func validateDirectories(g *generator) error {
	seenPaths := make(map[string]struct{})
	for _, dir := range g.Directories {
		if _, ok := seenPaths[dir.Path]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateDirectory, dir.Path)
		}
		seenPaths[dir.Path] = struct{}{}
	}

	return nil
}

func generateSymlinks(cfg *config.ApplianceConfig, g *generator) error {
	for _, link := range cfg.Symlinks {
		s := ignitionTypes.Link{
			Node: ignitionTypes.Node{
				Path: link.Path,
			},
			LinkEmbedded1: ignitionTypes.LinkEmbedded1{
				Hard:   toPtr(false),
				Target: toPtr(link.Target),
			},
		}

		if link.Owner != "" {
			s.Node.User = ignitionTypes.NodeUser{
				Name: toPtr(link.Owner),
			}

		}

		if link.Group != "" {
			s.Node.Group = ignitionTypes.NodeGroup{
				Name: toPtr(link.Group),
			}
		}

		if link.Overwrite {
			s.Overwrite = toPtr(link.Overwrite)
		}

		g.Links = append(g.Links, s)
	}

	return nil
}

func validateSymlinks(g *generator) error {
	seenPaths := make(map[string]struct{})
	for _, link := range g.Links {
		if _, ok := seenPaths[link.Path]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateSymlink, link.Path)
		}
		seenPaths[link.Path] = struct{}{}
	}

	return nil
}
