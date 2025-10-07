package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
)

func main() {
	// Utiliser le port depuis les variables d'environnement ou 8090 par défaut
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/blue", blueHandler)
	
	println("🚀 Server starting on port " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Go App</title>
		</head>
		<body>
			<h1>🚀 Hello from Go Application!</h1>
			<p>Your application is running successfully!</p>
			<ul>
				<li><a href="/blue">Blue Image</a></li>
				<li><a href="/health">Health Check</a></li>
			</ul>
		</body>
		</html>
	`))
}

func blueHandler(w http.ResponseWriter, r *http.Request) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 255, 255}}, image.ZP, draw.Src)
	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}
