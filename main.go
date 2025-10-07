package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/scan", scanHandler)
	http.HandleFunc("/badge", badgeHandler)

	fmt.Println("🚀 Go Dev Dashboard running on port " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

// Page principale
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Go Dev Dashboard</title>
			<style>
				body { font-family: Arial; background: #f4f4f4; margin: 40px; color: #333; }
				h1 { color: #007bff; }
				.card { background: #fff; padding: 20px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); margin-bottom: 20px; }
				a { color: #007bff; text-decoration: none; }
			</style>
		</head>
		<body>
			<h1>👋 Go Dev Dashboard</h1>
			<div class="card">
				<p>Bienvenue dans votre tableau de bord freelance Go.</p>
				<ul>
					<li><a href="/health">Health Check</a></li>
					<li><a href="/metrics">View Metrics</a></li>
					<li><a href="/scan">Dependency Scan</a></li>
					<li><a href="/badge">Generate Badge</a></li>
				</ul>
			</div>
			<p>🕓 Uptime: <b>` + time.Since(startTime).String() + `</b></p>
		</body>
		</html>
	`))
}

// Endpoint de santé
func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok", "message": "Service healthy"}
	json.NewEncoder(w).Encode(resp)
}

// Endpoint de métriques système
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	mem := &runtime.MemStats{}
	runtime.ReadMemStats(mem)

	data := map[string]interface{}{
		"go_version": runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
		"alloc_mb":   mem.Alloc / 1024 / 1024,
		"uptime":     time.Since(startTime).String(),
	}
	json.NewEncoder(w).Encode(data)
}

// Scan des dépendances
func scanHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("go", "list", "-m", "-u", "all")
	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "Erreur d'analyse des dépendances: "+err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(out)
}

// Génération d’un badge dynamique
func badgeHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "ok"
	}

	var c color.RGBA
	switch status {
	case "ok":
		c = color.RGBA{0, 200, 0, 255} // vert
	case "warn":
		c = color.RGBA{255, 165, 0, 255} // orange
	default:
		c = color.RGBA{200, 0, 0, 255} // rouge
	}

	img := image.NewRGBA(image.Rect(0, 0, 120, 40))
	draw.Draw(img, img.Bounds(), &image.Uniform{c}, image.Point{}, draw.Src)

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}
