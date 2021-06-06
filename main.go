package main

import (
	"context"
	"do-manager/manager"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
)

func main() {
	ctx := context.TODO()
	digitalOceanToken, tokenFound := os.LookupEnv("DIGITALOCEANTOKEN")
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

	totalSpaceUsed := new(float64)

	subscriptionMemoryChannel := make(chan float64)
	repositoryChannel := make(chan []manager.Repository)
	tagsChannel := make(chan [][]manager.RepositoryTag)
	errorChannel := make(chan error)

	registryManager := manager.Initialize(digitalOceanToken, registry, maxImage)

	waitGroup := new(sync.WaitGroup)

	waitGroup.Add(2)

	go registryManager.GetAllocatedSubscriptionMemory(ctx, subscriptionMemoryChannel, errorChannel, waitGroup)
	go registryManager.GetRepositories(ctx, repositoryChannel, errorChannel, waitGroup)

	subscriptionMemoryAllocated, repositories := <-subscriptionMemoryChannel, <-repositoryChannel

	go registryManager.GetRepositoryTags(ctx, repositories, totalSpaceUsed, tagsChannel, errorChannel)

	tags := <-tagsChannel

	percentageSpaceUsed := math.Ceil((*totalSpaceUsed / subscriptionMemoryAllocated) * 100)

	fmt.Printf("You have used over %.0f percent of allocated memory for the month\n", percentageSpaceUsed)

	if percentageSpaceUsed > float64(percentage) {
		deletedTags := registryManager.DeleteExtraTags(ctx, repositories, tags)

		if deletedTags > 1 {
			status, err := registryManager.StartGarbageCollection(ctx)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Your current garbage collection status is %s\n", status)
		}
	}

	if len(errorChannel) > 0 {
		for err := range errorChannel {
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	waitGroup.Wait()
}
