package poki_handler

import "github.com/go-rod/rod"

var (
	browser *rod.Browser
)

func Initialize() {
	browser = rod.New().MustConnect()
}
