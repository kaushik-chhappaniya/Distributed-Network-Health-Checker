package compose

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	Nodes   int
	Output  string
	Network string
	Image   string
}

func Generate(cfg Config, log *log.Logger) {

	log.Printf("<INFO> Creating docker compose file '%s' with %d nodes.", cfg.Output, cfg.Nodes)

	fileDeleted, err := checkAlreadyExists(&cfg.Output)
	if err != nil {
		log.Fatalf("<ERROR> Error checking existing file: %v", err)
	}
	if fileDeleted {
		log.Printf("<INFO> Existing file '%s' deleted successfully.", cfg.Output)
	}

	var sb strings.Builder
	sb.WriteString("services:\n")
	container_name, _, _ := strings.Cut(cfg.Image, ":")
	for i := 0; i < cfg.Nodes; i++ {
		fmt.Fprintf(&sb, fmt.Sprintf(
			`  %s%d:
    image: %s
    container_name: %s%d
    command: tail -f /dev/null
    networks:
     - %s
`, container_name, i, cfg.Image, container_name, i, cfg.Network))
	}

	fmt.Fprintf(&sb, "\nnetworks:\n  %s:\n    name: %s\n    driver: bridge\n", cfg.Network, cfg.Network)

	err = os.WriteFile(cfg.Output, []byte(sb.String()), 0644)
	if err != nil {
		log.Fatalf("<ERROR> Error writing to file: %v", err)
	}

	log.Printf("<INFO> Docker compose file '%s' created successfully with %d nodes.", cfg.Output, cfg.Nodes)
}

func checkAlreadyExists(outputFile *string) (bool, error) {
	// Delete file if it exists
	if _, err := os.Stat(*outputFile); err == nil {
		err = os.Remove(*outputFile)
		if err != nil {
			return false, fmt.Errorf("Warning: could not delete existing file: %v", err)
		}
	}
	return true, nil
}
