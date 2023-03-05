package main

import (
	"DDNS/pkg/cloud"
	"DDNS/pkg/common"
	"DDNS/pkg/deamon"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

var configPath *string = flag.String("configPath", "./config.json", "Config json file")

func main() {
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	cloud.Load()

	config := common.Config{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	cloudProvider, err := common.Manager.GetCloudProvider(config.Cloud, config.Version)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = cloudProvider.Init(config.Extra)
	if err != nil {
		fmt.Println(err)
	}

	/*
		if err := cloudProvider.Update(); err != nil {
			fmt.Println(err)
		}
	*/
	fmt.Println("test ui")

	deamon.Show()

}
