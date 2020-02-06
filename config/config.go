package main

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Age        int
	Cats       []string
	Pi         float64
	Perfection []int
	DOB        time.Time
}

func main() {
	configExample := `Age = 198
Cats = [ "Cauchy", "Plato" ]
Pi = 3.14
Perfection = [ 6, 28, 496, 8128 ]
DOB = 1987-07-05T05:45:00Z`

	var conf Config
	if _, err := toml.Decode(configExample, &conf); err != nil {
		// handle error
	}

	fmt.Printf("Pi: %f\n", conf.Pi)
}
