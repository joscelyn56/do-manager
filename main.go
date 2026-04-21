package main

import (
	"context"
	"do-manager/manager"
	"log"
	"os"
	"strconv"
	"time"
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

	waitPeriodStr, _ := os.LookupEnv("WAIT_PERIOD")
	cleanupEnabledStr, _ := os.LookupEnv("CLEANUP_ENABLED")

	// Convert max image count to int
	maxImage, err := strconv.Atoi(maxImageCount)
	if err != nil {
		log.Fatal("INVALID MAX IMAGE COUNT PROVIDED. MUST BE A NUMBER")
	}

	percentage, err := strconv.Atoi(percentageThreshold)
	if err != nil {
		log.Fatal("INVALID PERCENTAGE THRESHOLD PROVIDED. MUST BE A NUMBER")
	}

	waitPeriod := 10 // default 10 minutes
	if waitPeriodStr != "" {
		wp, err := strconv.Atoi(waitPeriodStr)
		if err != nil {
			log.Fatal("INVALID WAIT PERIOD PROVIDED. MUST BE A NUMBER (minutes)")
		}
		waitPeriod = wp
	}

	cleanupEnabled := true // default: enabled
	if cleanupEnabledStr != "" {
		ce, err := strconv.ParseBool(cleanupEnabledStr)
		if err != nil {
			log.Fatal("INVALID CLEANUP_ENABLED PROVIDED. MUST BE A BOOLEAN (true/false)")
		}
		cleanupEnabled = ce
	}

	manager.RunContainerManager(ctx, digitalOceanToken, registry, maxImage, percentage, time.Duration(waitPeriod)*time.Minute, cleanupEnabled)
}
