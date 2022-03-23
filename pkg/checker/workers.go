package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bythepixel/urlchecker/pkg/client"
)

var maxErrors uint64 = 5

func XMLWorker(ctx context.Context, cancel context.CancelFunc, urlChan chan string, id int, messager Messager, wg *sync.WaitGroup, sleep time.Duration, errorCount *uint64) {
	defer wg.Done()
	for {
		select {
		case url, ok := <-urlChan:
			if !ok {
				return
			}

			status, _, err := client.Fetch(url)
			if err != nil {
				log.Printf("Error: %s\n", err.Error())
			}

			if status != 200 {
				log.Println(status)
				msg := fmt.Sprintf("Invalid HTTP Response Status %d", status)
				messager.SendMessage(status, url, msg)
				atomic.AddUint64(errorCount, 1)
			}

			if *errorCount > maxErrors {
				log.Printf("Aborting... error count [%d] is greater than max error count [%d]", *errorCount, maxErrors)
				cancel()
			}

			time.Sleep(sleep * time.Second)
		case <-ctx.Done():
			return
		}

	}

}
