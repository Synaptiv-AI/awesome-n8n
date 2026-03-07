package condense

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

type Cache struct {
	dir string
}

func NewCache(dir string) *Cache {
	os.MkdirAll(dir, 0755)
	return &Cache{dir: dir}
}

func (c *Cache) Get(content string) (string, bool) {
	path := c.pathFor(content)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func (c *Cache) Put(content, condensed string) {
	path := c.pathFor(content)
	os.WriteFile(path, []byte(condensed), 0644)
}

func (c *Cache) pathFor(content string) string {
	hash := sha256.Sum256([]byte(content))
	return filepath.Join(c.dir, fmt.Sprintf("%x.txt", hash))
}
