package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bry-guy/factorio-lsp-plugin/pkg/api"
	"github.com/bry-guy/factorio-lsp-plugin/pkg/generator"
	"github.com/spf13/cobra"
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
		fmt.Println("Downloading and parsing Factorio API JSON...")

		// 1. Download and Parse JSON
		runtimeAPI := &api.API{}
		err := api.DownloadAndParseAPI(runtimeURL, runtimeAPI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading/parsing runtime API: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully parsed runtime API from %s\n", runtimeURL)

		prototypeAPI := &api.API{}
		err = api.DownloadAndParseAPI(prototypeURL, prototypeAPI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading/parsing prototype API: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully parsed prototype API from %s\n", prototypeURL)

		// 2. Generate Lua Definitions
		gen := generator.NewGenerator()
		definitions, err := gen.GenerateDefinitions(runtimeAPI, prototypeAPI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating Lua definitions: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Successfully generated Lua definitions in memory.")

		// 3. Write Definitions to Files
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory %s: %v\n", outputDir, err)
			os.Exit(1)
		}
		fmt.Printf("Ensured output directory %s exists.\n", outputDir)

		for filename, content := range definitions {
			outputPath := filepath.Join(outputDir, filename)
			err := os.WriteFile(outputPath, []byte(content), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing definition file %s: %v\n", outputPath, err)
				os.Exit(1)
			}
			fmt.Printf("Generated %s\n", outputPath)
		}

		fmt.Println("\nFactorio Lua definitions generated successfully.")
		fmt.Printf("Generated files are located in: %s\n", outputDir)
		fmt.Println("\nTo use these definitions with lua-language-server, configure your editor's settings to add this directory to the Lua.workspace.library setting.")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&runtimeURL, "runtime-url", "https://lua-api.factorio.com/latest/runtime-api.json", "URL for the Factorio Runtime API JSON")
	rootCmd.PersistentFlags().StringVar(&prototypeURL, "prototype-url", "https://lua-api.factorio.com/latest/prototype-api.json", "URL for the Factorio Prototype API JSON")
	rootCmd.PersistentFlags().StringVar(&outputDir, "output", "./output/factorio", "Output directory for generated Lua definitions")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
