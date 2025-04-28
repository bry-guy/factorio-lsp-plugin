package main

import (
	"log" // Import the log package
	"os"
	"path/filepath"

	"github.com/bry-guy/factorio-lsp-plugin/pkg/api"       // Corrected import path
	"github.com/bry-guy/factorio-lsp-plugin/pkg/generator" // Corrected import path
	"github.com/spf13/cobra"                               // Using Cobra for better CLI
)

var (
	runtimeURL   string
	prototypeURL string
	outputDir    string
)

var rootCmd = &cobra.Command{
	Use:   "factorio-api-gen",
	Short: "factorio-api-gen generates LuaLS definitions from Factorio API JSON",
	Long:  `A tool to download the Factorio Runtime and Prototype API JSON files and generate Lua Language Server definition files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Configure logging
		log.SetOutput(os.Stdout)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

		log.Println("Starting Factorio API Generator...")
		log.Printf("Runtime API URL: %s", runtimeURL)
		log.Printf("Prototype API URL: %s", prototypeURL)
		log.Printf("Output Directory: %s", outputDir)

		// 1. Download and Parse Runtime API JSON
		runtimeAPI := &api.API{}
		log.Println("Initiating runtime API download and parsing...")
		err := api.DownloadAndParseAPI(runtimeURL, runtimeAPI)
		if err != nil {
			log.Fatalf("Fatal error downloading/parsing runtime API from %s: %v", runtimeURL, err)
		}
		log.Println("Runtime API download and parsing complete.")

		// 2. Download and Parse Prototype API JSON
		prototypeAPI := &api.API{}
		log.Println("Initiating prototype API download and parsing...")
		err = api.DownloadAndParseAPI(prototypeURL, prototypeAPI)
		if err != nil {
			log.Fatalf("Fatal error downloading/parsing prototype API from %s: %v", prototypeURL, err)
		}
		log.Println("Prototype API download and parsing complete.")

		// 3. Generate Lua Definitions
		log.Println("Initiating Lua definition generation...")
		gen := generator.NewGenerator()
		definitions, err := gen.GenerateDefinitions(runtimeAPI, prototypeAPI)
		if err != nil {
			log.Fatalf("Fatal error generating Lua definitions: %v", err)
		}
		log.Println("Lua definition generation complete.")

		// 4. Write Definitions to Files
		log.Printf("Ensuring output directory exists: %s", outputDir)
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			log.Fatalf("Fatal error creating output directory %s: %v", outputDir, err)
		}
		log.Println("Output directory is ready.")

		log.Println("Writing generated definitions to files...")
		for filename, content := range definitions {
			outputPath := filepath.Join(outputDir, filename)
			log.Printf("Writing file: %s", outputPath)
			err := os.WriteFile(outputPath, []byte(content), 0644)
			if err != nil {
				log.Fatalf("Fatal error writing definition file %s: %v", outputPath, err)
			}
			log.Printf("Successfully wrote %s", outputPath)
		}

		log.Println("\nFactorio Lua definitions generated successfully.")
		log.Printf("Generated files are located in: %s", outputDir)
		log.Println("\nTo use these definitions with lua-language-server, configure your editor's settings to add this directory to the Lua.workspace.library setting.")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&runtimeURL, "runtime-url", "https://lua-api.factorio.com/latest/runtime-api.json", "URL for the Factorio Runtime API JSON")
	rootCmd.PersistentFlags().StringVar(&prototypeURL, "prototype-url", "https://lua-api.factorio.com/latest/prototype-api.json", "URL for the Factorio Prototype API JSON")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output", "./output/factorio", "Output directory for generated Lua definitions")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra handles errors by printing to Stderr, but we can log here too if needed
		// log.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
