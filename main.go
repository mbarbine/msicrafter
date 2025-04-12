msicrafter/
├── main.go
├── core/
│   ├── msi_reader.go        # Table listing, query, schema reading
│   ├── msi_editor.go        # Editing records, validations
│   ├── msi_transform.go     # Create transform from snapshot
│   ├── msi_diff.go          # Patch comparison between MSIs
│   ├── msi_export.go        # Table exporter (JSON, CSV) and ZIP
│   └── error_handler.go     # Wrapper functions for recovery/logging
├── retro/
│   ├── screen.go            # Retro ANSI layout and screen drawing
│   ├── colors.go            # Terminal color and effect helpers
├── cli/
│   ├── commands.go          # Entry CLI logic
├── assets/
│   ├── splash.txt           # ASCII art splash screen
├── go.mod
