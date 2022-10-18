package manager

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
)

func RunContainerManager(ctx context.Context, apiToken string, registryName string, count int, percentageThreshold int) {
	totalSpaceUsed := new(float64)

	subscriptionMemoryChannel := make(chan float64)
	repositoryChannel := make(chan []Repository)
	tagsChannel := make(chan [][]RepositoryTag)
	errorChannel := make(chan error)

	registryManager := Initialize(apiToken, registryName, count)

	waitGroup := new(sync.WaitGroup)

	waitGroup.Add(2)

	go registryManager.GetAllocatedSubscriptionMemory(ctx, subscriptionMemoryChannel, errorChannel, waitGroup)
	go registryManager.GetRepositories(ctx, repositoryChannel, errorChannel, waitGroup)

	subscriptionMemoryAllocated, repositories := <-subscriptionMemoryChannel, <-repositoryChannel

	waitGroup.Add(1)

	go registryManager.GetRepositoryTags(ctx, repositories, totalSpaceUsed, tagsChannel, errorChannel, waitGroup)

	tags := <-tagsChannel

	percentageSpaceUsed := math.Ceil((*totalSpaceUsed / subscriptionMemoryAllocated) * 100)

	fmt.Printf("You have used over %.0f percent of allocated memory for the month\n", percentageSpaceUsed)

	if percentageSpaceUsed > float64(percentageThreshold) {
		deletedTags := registryManager.DeleteExtraTags(ctx, repositories, tags)
		fmt.Printf("%d tags were deleted\n", deletedTags)
	}

	activeGarbageCollection := registryManager.GetActiveGarbageCollection(ctx)

	if activeGarbageCollection == "Active" {
		fmt.Printf("Garbage collection is active")
	} else {
		status, err := registryManager.StartGarbageCollection(ctx)
		if err == nil {
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
