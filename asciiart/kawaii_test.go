package asciiart_test

import (
	"fmt"
	"testing"

	"irc-client/asciiart"
	"irc-client/asciiart/art"
)

func TestKawaiiArt(t *testing.T) {
	// Create a new collection
	kawaiiCollection := asciiart.NewCollection("Kawaii Collection")

	// Add some cat art
	for i, cat := range art.CatFaces {
		kawaiiCollection.AddArt(fmt.Sprintf("cat_%d", i), cat)
	}

	// Add some kawaii faces
	for i, face := range art.KawaiiFaces {
		kawaiiCollection.AddArt(fmt.Sprintf("face_%d", i), face)
	}

	// Get terminal width
	width := asciiart.TerminalWidth()

	t.Logf("Terminal width: %d\n", width)

	// Display a welcome message
	t.Log("\n" + asciiart.CreateBanner("Welcome to Kawaii IRC!", "box", width))

	// Get and display a random piece of art
	randomArt, err := kawaiiCollection.GetRandom()
	if err != nil {
		t.Fatalf("Error getting random art: %v", err)
	}

	t.Log("\nRandom Kawaii Art:")
	t.Log(randomArt.Content)

	// Display a cat with a message
	t.Log("\nCat with message:")
	// Just use one of the cat faces directly
	t.Log(asciiart.CenterText(art.CatFaces[0], width))
}
