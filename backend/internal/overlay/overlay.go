package overlay

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/grysha11/camagru-backend/internal/imaging"
)

type Entry struct {
	ID    string
	Bytes []byte
	Image image.Image
}

type Store struct {
	entries map[string]Entry
	order   []string
}

func Load(dir string) (*Store, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("overlay: reading %q: %w", dir, err)
	}

	s := &Store{entries: make(map[string]Entry)}
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".png" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("overlay: reading %q: %w", f.Name(), err)
		}

		img, err := imaging.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("overlay: decoding %q: %w", f.Name(), err)
		}

		s.entries[f.Name()] = Entry{ID: f.Name(), Bytes: data, Image: img}
		s.order = append(s.order, f.Name())
	}

	return s, nil
}

func (s *Store) Get(id string) (Entry, bool) {
	e, ok := s.entries[id]
	return e, ok
}

func (s *Store) List() []Entry {
	out := make([]Entry, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.entries[id])
	}
	return out
}
