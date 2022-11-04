package loader

import (
	"os"
	"path/filepath"
)

type WalkFun func(dir, name string, isDir bool) error

func WalkDepth(name string, maxDepth int, fn WalkFun) error {
	var (
		dep = 0
		err error
	)
	dirs := []string{name}
	for len(dirs) > 0 && dep != maxDepth {
		dirs, err = WalkDirs(dirs, fn)
		if err != nil {
			return err
		}
		dep++
	}
	return nil
}

func WalkDirs(dirs []string, fn WalkFun) ([]string, error) {
	var subDirs []string
	for _, d := range dirs {
		ff, err := os.ReadDir(d)
		if err != nil {
			return nil, err
		}
		for _, f := range ff {
			err = fn(d, f.Name(), f.IsDir())
			if err != nil {
				return nil, err
			}
			if f.IsDir() {
				subDirs = append(subDirs, filepath.Join(d, f.Name()))
			}
		}
	}
	return subDirs, nil
}
