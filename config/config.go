package config

import "github.com/fortnoxab/fnxlogrus"

type Config struct {
	Log   fnxlogrus.Config
	Token string
	Port  string `default:"8080"`
}
