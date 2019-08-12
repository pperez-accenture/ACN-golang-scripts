package main

import (
	"fmt"
	"os/exec"
)

func main(){
	if err := exec.Command("cmd", "/C", "shutdown", "/f", "/r", "/t", "0").Run(); err != nil {
		fmt.Println("Failed to initiate shutdown:", err)
	}
}