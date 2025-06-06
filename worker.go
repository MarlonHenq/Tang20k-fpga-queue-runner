package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Submission struct {
	Key      string `json:"key"`
	Code     string `json:"code"`
	Exercise string `json:"exercise"`
}

var ctx = context.Background()

func main() {
	godotenv.Load()

	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	for {
		result, err := client.BRPop(ctx, 0*time.Second, "verilog_jobs").Result()
		if err != nil || len(result) < 2 {
			log.Println("Erro ao consumir da fila:", err)
			continue
		}

		var sub Submission
		if err := json.Unmarshal([]byte(result[1]), &sub); err != nil {
			log.Println("Erro ao decodificar job:", err)
			continue
		}

		fmt.Println("Processando exercício:", sub.Exercise)
		processJob(sub)
	}
}

func processJob(sub Submission) {
	// Criar diretório temporário
	dir := fmt.Sprintf("tmp/%d", time.Now().UnixNano())
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Printf("Erro ao criar diretório: %s\n", err)
		return
	}

	topModule := "top" // Altere se você espera outro nome
	mainFile := filepath.Join(dir, "main.v")
	synthFile := filepath.Join(dir, "synth.json")
	pnrFile := filepath.Join(dir, "pnr.json")
	outputBit := filepath.Join(dir, "out.fs")

	// Escreve o código Verilog no arquivo
	err = os.WriteFile(mainFile, []byte(sub.Code), 0644)
	if err != nil {
		log.Printf("Erro ao escrever Verilog: %s\n", err)
		return
	}

	// Comando 1: Yosys (sintetiza Verilog em JSON)
	err = runCommand(fmt.Sprintf(
		"yosys -p \"read_verilog main.v; synth_gowin -top %s; write_json %s\"",
		topModule, synthFile),
		dir,
	)
	if err != nil {
		log.Printf("Erro no Yosys: %s\n", err)
		return
	}

	// Comando 2: NextPNR (place & route)
	err = runCommand(fmt.Sprintf(
		"nextpnr-himbaechel --json %s --write %s --device GW1NR-LV9QN88PC6/I5 -vopt freq=27 --vopt enable-globals --vopt enable-auto-longwires --vopt family=GW1N-9C",
		synthFile, pnrFile),
		dir,
	)
	if err != nil {
		log.Printf("Erro no nextpnr-himbaechel: %s\n", err)
		return
	}

	// Comando 3: gowin_pack (gera bitstream)
	err = runCommand(fmt.Sprintf(
		"gowin_pack -d GW1N-9C -o %s %s",
		outputBit, pnrFile),
		dir,
	)
	if err != nil {
		log.Printf("Erro no gowin_pack: %s\n", err)
		return
	}

	// Comando 4: enviar para FPGA com openFPGALoader
	err = runCommand(fmt.Sprintf(
		"openFPGALoader -b tangPrimer20K %s",
		outputBit),
		dir,
	)
	if err != nil {
		log.Printf("Erro no openFPGALoader: %s\n", err)
		return
	}

	log.Printf("✅ Submissão \"%s\" concluída com sucesso!", sub.Exercise)
}

func runCommand(cmdStr string, dir string) error {
	log.Printf("Executando: %s", cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Erro:\n%s", string(out))
	}
	return err
}
