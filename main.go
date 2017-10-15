package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	initdb()
	defer closedb()
	var rport = flag.Int("rport", 5000, "registry port you want to start,it must not be used")
	var wport = flag.Int("wport", 8088, "web port you want to listen, it must not be used")
	go webTask(*wport)
	if client, container, err := startRegistry(*rport, *wport); err != nil {
		log.Fatalf("err hanppen %v", err)
	} else {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, os.Kill)
		go func() {
			for sig := range c {
				removeRegister(client, container)
				log.Fatalf("meet sig %+v", sig)
			}
		}()
		log.Printf("registry done!")
		select {}
	}
}
