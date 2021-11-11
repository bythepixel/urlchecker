package checker

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/bythepixel/urlchecker/pkg/client"
)

func XMLWorker(ctx context.Context, urlChan chan string, id int, messager Messager, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case url, ok := <-urlChan:
			if !ok {
				return
			}
			// log.Printf("Worker [%d] %s...\n", id, url)
			status, _, err := client.Fetch(url)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
			}

			if status != 200 {
				log.Println(status)
				msg := fmt.Sprintf("Invalid HTTP Response Status %d", status)
				messager.SendMessage(status, url, msg)
				continue
			}

			// log.Printf("Worker [%d] %s good!\n", id, url)
		case <-ctx.Done():
			return
		}

	}

}
