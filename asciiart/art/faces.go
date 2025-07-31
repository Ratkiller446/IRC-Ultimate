package art

import (
	"fmt"
	"math/rand"
	"time"
)

// KawaiiFaces contains various kawaii faces and expressions using ASCII characters
var KawaiiFaces = []string{
	"(^_^)",
	"(>^_^)>",
	"(^o^)/",
	"(^_~)",
	"(^_^)",
	"(^o^)b",
	"(^_^)v",
	"(^_^)d",
	"(^_^)y",
}

// GetRandomFace returns a random kawaii face
func GetRandomFace() string {
	rand.Seed(time.Now().UnixNano())
	return KawaiiFaces[rand.Intn(len(KawaiiFaces))]
}

// GetFaceWithMessage creates a kawaii face with a message
func GetFaceWithMessage(face, msg string) string {
	return fmt.Sprintf("%s %s", face, msg)
}
