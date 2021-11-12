package main

import (
	"context"
	"do-manager/manager"
	"flag"
	"os"
)

func main() {
	ctx := context.TODO()

	apiToken := flag.String("token", "", "Digitalocean API token")
	registryName := flag.String("registry", "", "Digitalocean container registry name")
	count := flag.Int("count", 3, "Minimum number of tags allowed in the repository")
	percentageThreshold := flag.Int("percentage", 50, "Maximum percentage threshold allowed before cleaning can occur")

	flag.Parse()

	if *apiToken == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *registryName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	manager.RunContainerManager(ctx, *apiToken, *registryName, *count, *percentageThreshold)
}
