package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

var (
	config = flag.String("config", "./config.json", "path to config file")
)

func main10() {
	flag.Parse()
	client := liteclient.NewConnectionPool()

	content, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	config := liteclient.GlobalConfig{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	err = client.AddConnectionsFromConfig(context.Background(), &config)
	if err != nil {
		log.Fatalln("connection err: ", err.Error())
		return
	}

	// initialize ton API lite connection wrapper
	api := ton.NewAPIClient(client)

	master, err := api.GetMasterchainInfo(context.Background())
	if err != nil {
		log.Fatalln("get masterchain info err: ", err.Error())
		return
	}

	log.Println(master)
}
