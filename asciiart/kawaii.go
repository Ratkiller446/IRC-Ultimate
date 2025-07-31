package asciiart

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Art represents a piece of ASCII art with its name and content
type Art struct {
	Name    string
	Content string
}

// Collection holds multiple pieces of ASCII art
type Collection struct {
	Name  string
	Arts  []Art
	theme string
}

// NewCollection creates a new collection of ASCII art
func NewCollection(name string) *Collection {
	return &Collection{
		Name: name,
		Arts: make([]Art, 0),
	}
}

// AddArt adds a new piece of art to the collection
func (c *Collection) AddArt(name, content string) {
	c.Arts = append(c.Arts, Art{
		Name:    name,
		Content: content,
	})
}

// GetRandom returns a random piece of art from the collection
func (c *Collection) GetRandom() (Art, error) {
	if len(c.Arts) == 0 {
		return Art{}, fmt.Errorf("no art available in collection")
	}
	rand.Seed(time.Now().UnixNano())
	return c.Arts[rand.Intn(len(c.Arts))], nil
}

// GetByName returns a specific piece of art by name
func (c *Collection) GetByName(name string) (Art, error) {
	for _, art := range c.Arts {
		if strings.EqualFold(art.Name, name) {
			return art, nil
		}
	}
	return Art{}, fmt.Errorf("art not found: %s", name)
}

// FormatForDisplay formats the art with proper padding for screen display
func (a Art) FormatForDisplay(width int) string {
	lines := strings.Split(a.Content, "\n")
	var result strings.Builder

	for _, line := range lines {
		padding := (width - len(line)) / 2
		if padding > 0 {
			result.WriteString(strings.Repeat(" ", padding))
		}
		result.WriteString(line)
		result.WriteString("\n")
	}

	return result.String()
}
