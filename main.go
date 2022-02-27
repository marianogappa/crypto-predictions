package main

import (
	"log"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/smrunner"
	"github.com/marianogappa/predictions/statestorage"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	postgresDBStorage, err := statestorage.NewPostgresDBStateStorage()
	if err != nil {
		log.Fatal(err)
	}
	market := market.NewMarket()

	s := newServer(postgresDBStorage, market)
	runner := smrunner.NewSMRunner(postgresDBStorage, market)
	go func() {
		for {
			result := runner.Run(int(time.Now().Unix()))
			if len(result.Errors) > 0 {
				log.Println("State Machine runner finished with errors:")
				for i, err := range result.Errors {
					log.Printf("%v) %v", i+1, err.Error())
				}
				log.Println()
			}
			time.Sleep(10 * time.Second)
		}
	}()

	s.mustBlockinglyServeAll(1234)
}
