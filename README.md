# msicrafter

“Retro-style MSI editing and transformation wizardry — right in your terminal.”

## Key Capabilities

| Capability          | Details                                                                 |
|---------------------|-------------------------------------------------------------------------|
| 📄 Table Explorer   | View tables, schema, and records (ANSI-bordered, colored terminal)       |
| ✍️ Table Editor      | Add/edit/delete records; validation included                             |
| 🧠 MSI Validation    | Built-in schema validator and required-field check                       |
| 🔁 Transform Support | Create `.mst` transform files from before/after states                  |
| 🔍 Patch Comparison  | Compare two MSI files for table-level differences                       |
| 📦 Export & Zip     | Backup original MSI, export tables as CSV/JSON, compress changes         |
| 🧯 Error Handling    | All actions wrapped with recoverable `try/catch`-like handlers/logging   |
| 💾 Safe Save         | Confirm changes with prompt; optionally skip/abort per table             |
| 🎨 Retro Output      | Colorful ASCII UI, pseudo-modal prompts, animated “Working…” displays    |

## Folder Structure

```
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
```

## Key Libraries

- `github.com/go-ole/go-ole` – COM automation
- `github.com/charmbracelet/lipgloss` + `bubbletea` – retro-style terminal UI
- `github.com/dsnet/compress` – fast zipping
- `github.com/urfave/cli/v2` – CLI structure
- `encoding/csv`, `encoding/json` – for exports
- `log`, `errors`, and custom recoverable wrappers

## Example Usage

```bash
# View tables
msicrafter tables ./MyApp.msi

# Query contents
msicrafter query ./MyApp.msi "SELECT * FROM Property"

# Edit
msicrafter edit ./MyApp.msi --table Property --set ProductVersion=9.9.9

# Create transform (diff-based)
msicrafter transform --original original.msi --modified edited.msi --output patch.mst

# Export and zip
msicrafter export ./MyApp.msi --format json --zip

# Compare two MSI files
msicrafter diff ./v1.msi ./v2.msi
```

## Resilience Strategy

| Component      | Resilience Method                             |
|----------------|-----------------------------------------------|
| MSI Ops        | Wrapped in `safeExecute("opName", func() {})` |
| Log            | Writes structured logs to `.msicrafter.log`   |
| Panic Recover  | Full `recover()` with retro splash            |
| Dry Run Mode   | `--dry-run` available before committing        |


## TIPS AND TRICKS


## FEEDBACK 



```
## Additional Steps

### Test

Run these commands:
```
go mod tidy
go build -o msicrafter.exe
```

Then execute:
```
./msicrafter tables "C:\Path\To\Sample.msi"
```

### Next Milestones

- Add query with arbitrary SQL
- Build edit and validation logic
- Snapshot & diff → transform
- Zip export before save
- Structured logging + error recovery
- Fun retro progress/status UI
```