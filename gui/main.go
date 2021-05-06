package main

import (
  _ "embed"
  "github.com/wailsapp/wails"
)

//go:embed frontend/build/main.js
var js string

//go:embed frontend/build/main.css
var css string

func main() {

  app := wails.CreateApp(&wails.AppConfig{
    Width:  1024,
    Height: 768,
    Title:  "tip",
    JS:     js,
    CSS:    css,
  })
  app.Bind(&Controller{})
  app.Run()
}
