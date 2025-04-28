# Factorio LuaLS API Generator

This project provides a Golang tool to generate Lua Language Server (`lua-language-server`) definition files from the official Factorio Runtime and Prototype API JSON documentation.

These generated definitions can be used by `lua-language-server` to provide accurate autocompletion, hover documentation, and diagnostics when modding Factorio.

## Getting Started

### Prerequisites

* Go (version 1.18 or higher recommended)
* Git

### Building the Tool

1.  Clone this repository:
    ```bash
    git clone <repository_url>
    cd factorio-lua-ls-api-gen
    ```
2.  Build the Go application:
    ```bash
    go build -o factorio-api-gen
    ```
    This will create an executable file named `factorio-api-gen` (or `factorio-api-gen.exe` on Windows) in the current directory.

### Running the Generator

Run the compiled tool to download the latest API JSON files and generate the Lua definitions:

```bash
./factorio-api-gen
```

By default, this will:

* Download from `https://lua-api.factorio.com/latest/runtime-api.json` and `https://lua-api.factorio.com/latest/prototype-api.json`.
* Generate the `.lua` definition files in the `./output/factorio` directory.

You can customize the URLs and output directory using command-line flags:

```bash
./factorio-api-gen --runtime-url <custom_runtime_url> --prototype-url <custom_prototype_url> --output <custom_output_directory>
```

### Using the Generated Definitions with `lua-language-server`

1.  Ensure you have `lua-language-server` installed and configured for your editor (e.g., VS Code extension, Neovim LSP setup).
2.  Configure your `lua-language-server` settings to include the generated output directory in its library path.

    For **Visual Studio Code**, open your settings (`settings.json`) and add or modify the `Lua.workspace.library` setting:

    ```json
    {
        "Lua.workspace.library": [
            "/path/to/your/factorio-lua-ls-api-gen/output/factorio"
            // Add other Lua libraries you use here
        ]
    }
    ```
    Replace `/path/to/your/factorio-lua-ls-api-gen` with the actual path to the directory where you cloned this repository.

    For **other editors**, consult your `lua-language-server` or LSP client documentation on how to configure the workspace library.

3.  Restart your editor or the `lua-language-server` to load the new definitions.

## Repository Structure

```
factorio-lua-ls-api-gen/
├── go.mod               # Go module file
├── go.sum               # Go dependency checksums
├── main.go              # Main application entry point
├── pkg/                 # Internal packages
│   ├── api/             # Handles API data structures and loading
│   │   ├── types.go     # Go structs for JSON unmarshalling
│   │   └── loader.go    # Functions for downloading and parsing JSON
│   └── generator/       # Handles generating LuaLS definitions
│       └── generator.go # Logic for converting API data to LuaLS annotations
└── README.md            # This file
└── .gitignore           # Specifies intentionally untracked files
└── LICENSE              # Project license
```

## Contributing

If you find issues with the generated definitions or the tool, feel free to open an issue or submit a pull request.

## License

TODO
