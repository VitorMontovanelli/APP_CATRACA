package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// O Wails/WebView2 não expõe o handler de permissão de mídia por padrão,
	// o que faz com que o getUserMedia (câmera) seja negado silenciosamente.
	// Este argumento faz o Chromium/WebView2 autorizar automaticamente os
	// pedidos de câmera e microfone, sem precisar de prompt. Deve ser definido
	// antes da criação do ambiente WebView2 (início do wails.Run).
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--use-fake-ui-for-media-stream")

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "CatracaVMD",
		Width:  800,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			app.onShutdown()
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
