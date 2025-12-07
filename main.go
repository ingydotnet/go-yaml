package main

import (
	"fmt"
	"go.yaml.in/yaml/v4"
)

func main() {
	var data map[string]string
	err := yaml.Unmarshal([]byte("key: value"), &data)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully unmarshaled:", data)
}
