package main

import (
	"DDNS/pkg/cloud"
	"DDNS/pkg/deamon"
	"flag"
	"fmt"
	"os"
)

type SubFlag struct {
	*flag.FlagSet
	comment string
}

//var configPath *string = flag.String("configPath", "./config.json", "Config json file")

func main() {
	stop := &SubFlag{
		FlagSet: flag.NewFlagSet("stop", flag.ExitOnError),
		comment: "stop ddns deamon process",
	}
	start := &SubFlag{FlagSet: flag.NewFlagSet("start", flag.ExitOnError), comment: "start ddns"}
	configPath := start.String("configPath", "./config.json", "Config json file")

	subCommands := map[string]*SubFlag{
		start.Name(): start,
		stop.Name():  stop,
	}

	/*
		usage := func() {
			fmt.Println("Usage: ddns COMMAND")
			for _, v := range subCommands {
				fmt.Println(v.Name(), v.comment)
				v.PrintDefaults()
				fmt.Println()
			}
		}
	*/

	if len(os.Args) < 2 {
		//usage()
		os.Args = append(os.Args, "start")
	}
	subFlag, ok := subCommands[os.Args[1]]
	if !ok {
		subFlag = subCommands["start"]
		_ = subFlag.Parse(os.Args[1:])
	} else {
		_ = subFlag.Parse(os.Args[2:])
	}

	//flag.Parse()

	cloud.Load()

	switch subFlag.Name() {
	case "start":
		data, err := os.ReadFile(*configPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		deamon.Run(data)
	case "stop":
		deamon.Stop()
	}

}
