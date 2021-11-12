package main

import (
	"context"
	"do-manager/manager"
	"log"
	"os"
	"strconv"
)

func main() {
	ctx := context.TODO()
	digitalOceanToken, tokenFound := os.LookupEnv("DIGITALOCEAN_TOKEN")
	registry, registryFound := os.LookupEnv("REGISTRY")
	maxImageCount, maxImageCountFound := os.LookupEnv("MAX_IMAGE_COUNT")
	percentageThreshold, percentageThresholdFound := os.LookupEnv("PERCENTAGE_THRESHOLD")

	if !tokenFound {
		log.Fatal("DIGITALOCEAN TOKEN NOT SET")
	}

	if !registryFound {
		log.Fatal("REGISTRY NAME NOT SET")
	}

	if !maxImageCountFound {
		log.Fatal("MAXIMUM IMAGE COUNT NOT SET")
	}

	if !percentageThresholdFound {
		log.Fatal("PERCENTAGE THRESHOLD NOT SET")
	}

	// Convert max image count to int
	maxImage, err := strconv.Atoi(maxImageCount)
	if err != nil {
		log.Fatal("INVALID MAX IMAGE COUNT PROVIDED. MUST BE A NUMBER")
	}

	percentage, err := strconv.Atoi(percentageThreshold)
	if err != nil {
		log.Fatal("INVALID PERCENTAGE THRESHOLD PROVIDED. MUST BE A NUMBER")
	}

	manager.RunContainerManager(ctx, digitalOceanToken, registry, maxImage, percentage)
}
