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
	if len(os.Args) < 2 {
		fmt.Println("Deve especificar a pasta para a varredura!")
		return
	}

	inputDir := os.Args[1]
	outputDir := filepath.Join(inputDir, "thumbs")

	os.MkdirAll(outputDir, 0755)

	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".mp4" && ext != ".mov" && ext != ".mkv" {
			return nil
		}

		videoName := strings.TrimSuffix(filepath.Base(path), ext)
		safeName := sanitizeFilename(videoName)
		tmpDir := filepath.Join(os.TempDir(), "thumbs_"+safeName)
		os.MkdirAll(tmpDir, 0755)

		durationSec := getDurationSeconds(path)
		if durationSec <= 0 {
			log.Printf("Duração inválida para %s", path)
			return nil
		}

		// Posições: 1/6, 2/6, ..., 5/6 da duração
		var thumbFiles []string
		for i := 1; i <= 5; i++ {
			sec := float64(i) * durationSec / 6
			ts := fmt.Sprintf("%02d:%02d:%02d", int(sec)/3600, (int(sec)%3600)/60, int(sec)%60)
			outFile := filepath.Join(tmpDir, fmt.Sprintf("thumb_%d.jpg", i))
			cmd := exec.Command("ffmpeg", "-ss", ts, "-i", path, "-frames:v", "1", "-q:v", "2", outFile)
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.Run()
			thumbFiles = append(thumbFiles, outFile)
		}

		// Compor imagem final com os 5 thumbs
		outputFile := filepath.Join(outputDir, safeName+"_thumb.jpg")
		args := append([]string{"+append"}, thumbFiles...)
		args = append(args, outputFile)
		err = exec.Command("convert", args...).Run()
		if err != nil {
			log.Printf("Erro ao compor thumbs de %s: %v", videoName, err)
			return nil
		}

		// Adicionar anotação da duração com convert
		durationStr := formatDuration(durationSec)
		annotatedCmd := exec.Command("convert", outputFile,
			"-gravity", "NorthWest",
			"-pointsize", "72",
			"-fill", "white",
			"-undercolor", "black",
			"-annotate", "+10+10", "Duração: "+durationStr,
			outputFile)

		annotatedCmd.Stdout = nil
		annotatedCmd.Stderr = nil
		err = annotatedCmd.Run()
		if err != nil {
			log.Printf("Erro ao adicionar texto na imagem: %v", err)
		}

		fmt.Println("Thumbs geradas:", outputFile)

		return nil
	})
}

// getDurationSeconds usa ffprobe para extrair a duração em segundos
func getDurationSeconds(path string) float64 {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return 0
	}

	var dur float64
	fmt.Sscanf(out.String(), "%f", &dur)
	return dur
}

func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	return re.ReplaceAllString(name, "_")
}

func formatDuration(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
