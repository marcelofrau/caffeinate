package icon

import (
	"embed"
	"log"
)

//go:embed app_icon.ico
var iconFS embed.FS

// PNG returns the application icon as bytes.
func PNG() []byte {
	data, err := iconFS.ReadFile("app_icon.ico")
	if err != nil {
		log.Fatalf("failed to load icon: %v", err)
	}
	return data
}
