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
	// var prediction types.Prediction
	// err := json.Unmarshal([]byte(`
	// {
	// 	"version": "1.0.0",
	// 	"createdAt": "2022-02-09T17:22:00Z",
	// 	"authorHandle": "MarianoGappa",
	// 	"post": "https://twitter.com/kjwebgjkweb",
	// 	"define": {
	// 		"a": {
	// 			"variable": "BINANCE:BTC-USDT",
	// 			"operator": ">=",
	// 			"operand": 67601,
	// 			"from_ts": "2022-02-09T17:22:00Z",
	// 			"to_ts": "2022-04-26T00:00:00Z"
	// 		}
	// 	},
	// 	"predict": {
	// 		"predict": ""
	// 	}
	// }
	// `), &prediction)

	// if err != nil {
	// 	log.Fatal(err)
	// }
	// bs, err := json.Marshal(&prediction)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(bs))
	// os.Exit(0)

	boltDBStorage, err := statestorage.NewBoltDBStateStorage()
	if err != nil {
		log.Fatal(err)
	}
	market := market.NewMarket()

	s := newServer(boltDBStorage, market)
	runner := smrunner.NewSMRunner(boltDBStorage, market)
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
