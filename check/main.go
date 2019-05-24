package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gnasr/concourse-slack-alert-resource/concourse"
)

func main() {
	err := json.NewEncoder(os.Stdout).Encode(concourse.CheckResponse{})
	if err != nil {
		log.Fatalln(err)
	}
}
