package main

import (
	"fmt"

	"github.com/Nitesh-04/realtime-racing/config"
)

func main() {
	fmt.Println("init")
	config.ConnectDB()
}