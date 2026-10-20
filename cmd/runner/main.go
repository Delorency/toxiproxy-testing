package main

import (
	"flag"
	"log"
	"path/filepath"

	"pr4/internal/config"
	"pr4/internal/runner"
)

func main() {
	scenarioName := flag.String("scenario", "stepwise_degradation", "имя сценария без .json")
	flag.Parse()

	cfg := config.MustLoad()

	scenarioPath := filepath.Join(cfg.Runner.ScenariosDir, *scenarioName+".json")
	log.Printf("loading scenario: %s", scenarioPath)

	r := runner.New(cfg)
	outPath, err := r.Run(scenarioPath)
	if err != nil {
		log.Fatalf("runner error: %v", err)
	}

	log.Printf("results saved to: %s", outPath)
}
