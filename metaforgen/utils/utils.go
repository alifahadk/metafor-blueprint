package utils

import (
	"fmt"
	"metaforgen/config"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English)

func ToTitle(name string) string {
	return titleCaser.String(name)
}
func ExtractDependencies(cfg config.SystemConfig) map[string][]string {
	depGraph := make(map[string][]string)

	for _, server := range cfg.Servers {
		serviceName := fmt.Sprintf("svc%s", server.Name)
		seen := map[string]struct{}{}

		for _, apiCfg := range server.APIs {
			for _, ds := range apiCfg.DownstreamServices {
				targetService := fmt.Sprintf("svc%s", ds.Target)
				if targetService == "" || targetService == serviceName {
					continue
				}
				// Avoid duplicate dependencies
				if _, exists := seen[targetService]; !exists {
					depGraph[serviceName] = append(depGraph[serviceName], targetService)
					seen[targetService] = struct{}{}
				}
			}
		}
	}

	return depGraph
}
