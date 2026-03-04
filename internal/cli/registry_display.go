package cli

import (
	"fmt"

	"github.com/mohammedsamin/mcpup/internal/registry"
)

func formatRegistryOption(t registry.Template) string {
	return fmt.Sprintf("%-22s %-12s %s", t.Name, t.Category, truncateMenuText(t.Description, 72))
}

func formatSetupRegistryOption(t registry.Template) string {
	return fmt.Sprintf("%-22s %-12s %s", t.Name, t.Category, truncateMenuText(t.Description, 68))
}

func truncateMenuText(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}
