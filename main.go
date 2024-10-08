package main

import (
	"fmt"
	. "github.com/RobsonDevCode/BankApi/api"
	. "github.com/RobsonDevCode/BankApi/api/Setting"
	"log"
)

const (
	Development = iota
	UAT
	Production
)

// Api start up
func main() {
	fmt.Printf("Bank Api is now Starting up...")

	//Set Environment based on iota
	env := SetEnvironmentSettings(Development)

	err := SetConfig(env)

	if err != nil {
		log.Fatalf("Error: Faild Api Start up, %v", err)
	}

	store, err := NewAccountRepository()

	if err != nil {
		log.Fatalf("Error: Faild Api Start up, %v", err)
	}

	if err = store.Init(); err != nil {
		log.Fatalf("Error: Failed Api Start up, %v", err)
	}

	server := NewAPIServer(":3000", store)

	err = server.Start()

	if err != nil {
		log.Fatalf("Error: Failed Api Start up, %v", err)
	}
}
