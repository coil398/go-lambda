package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	u := uuid.New()
	uuid := u.String()
	fmt.Println(uuid)
}
