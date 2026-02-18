package main

import "time"

var data = map[string]any{
	"app.name":    "selfshop.ms-catalog",
	"app.runmode": "prod",

	"log.min_level": "info",
	"log.format":    "auto",

	"entry.http.write_timeout":   20 * time.Second,
	"entry.http.idle_timeout":    90 * time.Second,
	"entry.http.read_timeout":    10 * time.Second,
	"entry.http.request_timeout": 15 * time.Second,
}
