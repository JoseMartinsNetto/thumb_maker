package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	inputDir := os.Args[1]
	outputDir := filepath.Join(inputDir, "thumbs")

	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Erro ao criar pasta de saída: %v", err)
	}

	err = filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mp4" && ext != ".mov" && ext != ".mkv" {
			return nil
		}

		videoName := strings.TrimSuffix(filepath.Base(path), ext)
		safeName := sanitizeFilename(videoName)
		finalThumb := filepath.Join(outputDir, safeName+"_thumb.jpg")

		// 1. Obter duração do vídeo
		duration := getVideoDuration(path)

		// 2. Gerar imagem de thumbs
		cmdThumb := exec.Command("ffmpeg", "-i", path,
			"-vf", "select=not(mod(n\\,100)),scale=320:-1,tile=5x1",
			"-frames:v", "1", finalThumb)

		cmdThumb.Stdout = os.Stdout
		cmdThumb.Stderr = os.Stderr
		fmt.Println("Gerando thumbs para:", path)
		if err := cmdThumb.Run(); err != nil {
			log.Printf("Erro ao gerar thumbs de %s: %v\n", path, err)
			return nil
		}

		// 3. Adicionar texto com ImageMagick
		text := fmt.Sprintf("Duração: %s", duration)

		cmdText := exec.Command("convert", finalThumb,
			"-gravity", "NorthWest",
			"-pointsize", "24",
			"-fill", "white",
			"-undercolor", "black",
			"-annotate", "+10+10", text,
			finalThumb)

		cmdText.Stdout = os.Stdout
		cmdText.Stderr = os.Stderr
		fmt.Println("Adicionando texto com ImageMagick...")
		if err := cmdText.Run(); err != nil {
			log.Printf("Erro ao adicionar texto em %s: %v\n", finalThumb, err)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Finalizado.")
}

// getVideoDuration retorna a duração no formato HH:MM:SS
func getVideoDuration(path string) string {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "00:00:00"
	}

	var dur float64
	fmt.Sscanf(out.String(), "%f", &dur)
	h := int(dur) / 3600
	m := (int(dur) % 3600) / 60
	s := int(dur) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// sanitizeFilename remove caracteres inválidos de nomes de arquivos
func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	return re.ReplaceAllString(name, "_")
}
