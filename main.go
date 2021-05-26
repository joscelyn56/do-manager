package main

import (
	"context"
	"do-manager/manager"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"
)

func main() {
	ctx := context.TODO()
	digitalOceanToken, tokenFound := os.LookupEnv("DIGITALOCEANTOKEN")
	registry, registryFound := os.LookupEnv("REGISTRY")

	if !tokenFound {
		log.Fatal("DIGITALOCEAN TOKEN NOT SET")
	}

	if !registryFound {
		log.Fatal("REGISTRY NAME NOT SET")
	}

	totalSpaceUsed := new(float64)
	deletedTags := new(int)

	subscriptionMemoryChannel := make(chan float64)
	repositoryChannel := make(chan []manager.Repository)
	tagsChannel := make(chan [][]manager.RepositoryTag)
	errorChannel := make(chan error)

	fmt.Println(time.Now())

	registryManager := manager.Initialize(digitalOceanToken, registry, 2)

	waitGroup := new(sync.WaitGroup)

	waitGroup.Add(2)

	go registryManager.GetAllocatedSubscriptionMemory(ctx, subscriptionMemoryChannel, errorChannel, waitGroup)
	go registryManager.GetRepositories(ctx, repositoryChannel, errorChannel, waitGroup)

	if len(errorChannel) > 0 {
		for err := range errorChannel {
			if err != nil {
				fmt.Print(err)
			}
		}
		os.Exit(1)
	}

	subscriptionMemoryAllocated, repositories := <-subscriptionMemoryChannel, <-repositoryChannel

	go registryManager.GetRepositoryTags(ctx, repositories, totalSpaceUsed, tagsChannel)

	tags := <-tagsChannel

	percentageSpaceUsed := math.Ceil((*totalSpaceUsed / subscriptionMemoryAllocated) * 100)

	fmt.Printf("You have used over %.0f percent of allocated memory for the month\n", percentageSpaceUsed)

	if percentageSpaceUsed > float64(80) {
		waitGroup.Add(1)
		go registryManager.DeleteExtraTags(ctx, repositories, tags, deletedTags, waitGroup)
	}

	if *deletedTags > 1 {
		status := registryManager.StartGarbageCollection(ctx)
		fmt.Printf("Your current garbage collection status is %s\n", status)
	}

	waitGroup.Wait()

	fmt.Println(time.Now())
}
