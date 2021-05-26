package main

import (
	"context"
	"do-manager/manager"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
)

func main() {
	ctx := context.TODO()

	apiToken := flag.String("token", "", "Digitalocean API token")
	registryName := flag.String("registry", "", "Digitalocean container registry name")
	count := flag.Int("count", 3, "Minimum number of tags allowed in the repository")
	flag.Parse()

	if *apiToken == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *registryName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	totalSpaceUsed := new(float64)

	subscriptionMemoryChannel := make(chan float64)
	repositoryChannel := make(chan []manager.Repository)
	tagsChannel := make(chan [][]manager.RepositoryTag)
	errorChannel := make(chan error)

	registryManager := manager.Initialize(*apiToken, *registryName, *count)

	waitGroup := new(sync.WaitGroup)

	waitGroup.Add(2)

	go registryManager.GetAllocatedSubscriptionMemory(ctx, subscriptionMemoryChannel, errorChannel, waitGroup)
	go registryManager.GetRepositories(ctx, repositoryChannel, errorChannel, waitGroup)

	subscriptionMemoryAllocated, repositories := <-subscriptionMemoryChannel, <-repositoryChannel

	go registryManager.GetRepositoryTags(ctx, repositories, totalSpaceUsed, tagsChannel, errorChannel)

	tags := <-tagsChannel

	percentageSpaceUsed := math.Ceil((*totalSpaceUsed / subscriptionMemoryAllocated) * 100)

	fmt.Printf("You have used over %.0f percent of allocated memory for the month\n", percentageSpaceUsed)

	if percentageSpaceUsed > float64(80) {
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
