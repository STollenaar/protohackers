package problem10

import (
	"bytes"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var (
	legalCharacters = regexp.MustCompile(`^[a-zA-Z0-9\.\-\_]+$`).MatchString
)

type File struct {
	name string

	data      [][]byte
	childeren map[string]*File
}

func newFiles() *File {
	return &File{
		name:      "/",
		childeren: map[string]*File{},
	}
}

func (f *File) getFile(fullPath string) *File {
	fullPath = filepath.Clean(fullPath)

	if fullPath == "/" {
		return f
	}

	fullPath = strings.TrimPrefix(fullPath, "/")

	paths := strings.Split(fullPath, "/")
	current := f

	for _, path := range paths {
		if _, ok := current.childeren[path]; !ok {
			current.childeren[path] = &File{name: path, childeren: make(map[string]*File)}
		}
		current = current.childeren[path]
	}
	return current
}

func (f *File) listDir(fullPath string) (files []*File) {
	dir := f.getFile(fullPath)

	if dir == nil {
		return nil
	}
	for _, file := range dir.childeren {
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].name < files[j].name
	})

	return files
}

func (f *File) putFile(fullPath string, data []byte) int {
	file := f.getFile(fullPath)

	if len(data) == 0 {
		if len(file.data) == 0 {
			file.data = append(file.data, []byte{})
		}
		return len(file.data)
	}
	last := len(file.data) - 1
	if last < 0 {
		last = 0
	}

	if len(file.data) != 0 && bytes.Equal(file.data[last], data) {
		return len(file.data)
	}
	file.data = append(file.data, data)
	return len(file.data)
}

func (f *File) isLeggalName(fullPath string) bool {
	if strings.Contains(fullPath, "//") {
		return false
	}
	fullPath = filepath.Clean(fullPath)
	fullPath = strings.TrimPrefix(fullPath, "/")

	parts := strings.Split(fullPath, "/")

	for _, part := range parts {
		if part == "" {
			continue
		}
		if !legalCharacters(part) {
			return false
		}
	}
	return true
}

// checks if s is ascii and printable, aka doesn't include tab, backspace, etc.
func (f *File) IsPrintable(data []byte) bool {
	return utf8.Valid(data)
}
