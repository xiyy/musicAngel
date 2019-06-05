package main

import (
	"database/sql"
	"log"
	"musicAngel/config"
	"musicAngel/database"
	"musicAngel/httpserver"
	"net"
	"os"
	"os/signal"
	"runtime"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}
func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	db, err := sql.Open("mysql", config.MYSQL_CONFIG)
	if err != nil {
		log.Fatal(err)
	}
	dbManager := &database.DbManager{Db: db}
	defer dbManager.Close()
	listener, err := net.Listen("tcp", "localhost:80")
	if err != nil {
		log.Fatal(err)
	}
	server := httpserver.NewHttpServer(dbManager)
	go server.Serve(listener)

	<-signals
	server.Stop()

}
