# table

Static table rendering for CLI output.

## Overview

The `table` package provides simple, non-interactive table rendering for command-line applications. It's designed for tools that need to display tabular data and exit (like `kubectl get`, `ls -l`, `docker ps`, etc.).

This is separate from the interactive Bubble Tea components in the parent `tui` package.

## Installation

```bash
go get github.com/SCKelemen/tui/table
```

## Quick Start

```go
package main

import (
    "github.com/SCKelemen/tui/table"
)

func main() {
    t := table.New("Name", "Status", "Age")
    t.AddRow("service-a", "Running", "2d")
    t.AddRow("service-b", "Stopped", "5h")
    t.AddRow("service-c", "Starting", "30s")
    t.Print()
}
```

Output:

```
┌───────────┬──────────┬──────┐
│ Name      │ Status   │ Age  │
├───────────┼──────────┼──────┤
│ service-a │ Running  │ 2d   │
├───────────┼──────────┼──────┤
│ service-b │ Stopped  │ 5h   │
├───────────┼──────────┼──────┤
│ service-c │ Starting │ 30s  │
└───────────┴──────────┴──────┘
```

## Features

- **Unicode box drawing** - Beautiful rounded borders by default
- **Multiple border styles** - Rounded, double-line, or ASCII for compatibility
- **Automatic column sizing** - Columns auto-expand to fit content
- **Bold headers** - Optional bold formatting for headers
- **Simple or detailed** - Choose between full borders or compact mode

## Border Styles

### Rounded (default)
```go
t := table.New("A", "B")
t.AddRow("1", "2")
t.Print()
```

```
┌───┬───┐
│ A │ B │
├───┼───┤
│ 1 │ 2 │
└───┴───┘
```

### Double-line
```go
t := table.New("A", "B")
t.SetBorderStyle(table.BorderStyleDouble)
t.AddRow("1", "2")
t.Print()
```

```
╔═══╦═══╗
║ A ║ B ║
╠═══╬═══╣
║ 1 ║ 2 ║
╚═══╩═══╝
```

### ASCII (maximum compatibility)
```go
t := table.New("A", "B")
t.SetBorderStyle(table.BorderStyleASCII)
t.AddRow("1", "2")
t.Print()
```

```
+---+---+
| A | B |
+---+---+
| 1 | 2 |
+---+---+
```

## Simple Mode

For more compact output without row separators:

```go
t := table.New("Name", "Status")
t.AddRow("service-a", "Running")
t.AddRow("service-b", "Stopped")
t.PrintSimple()
```

```
┌───────────┬─────────┐
│ Name      │ Status  │
├───────────┼─────────┤
│ service-a │ Running │
│ service-b │ Stopped │
└───────────┴─────────┘
```

## API Reference

### Creating Tables

```go
// Create a table with headers
t := table.New("Column1", "Column2", "Column3")

// Set border style
t.SetBorderStyle(table.BorderStyleRounded)  // default
t.SetBorderStyle(table.BorderStyleDouble)
t.SetBorderStyle(table.BorderStyleASCII)

// Control header formatting
t.SetHeaderBold(true)  // default
t.SetHeaderBold(false)
```

### Adding Data

```go
// Add a single row
t.AddRow("value1", "value2", "value3")

// Add multiple rows at once
rows := [][]string{
    {"a", "b", "c"},
    {"d", "e", "f"},
}
t.AddRows(rows)

// Clear all rows (keeps headers)
t.Clear()
```

### Rendering

```go
// Get rendered string
output := t.Render()        // with row separators
output := t.RenderSimple()  // without row separators

// Print to stdout
t.Print()        // with row separators
t.PrintSimple()  // without row separators

// Or use as a Stringer
fmt.Println(t)  // calls Render()
```

## Use Cases

Perfect for:

- **kubectl-style output** - `myapp get resources`
- **ls-style listings** - `myapp list --format table`
- **Status displays** - `myapp status --all`
- **Comparison tables** - `myapp compare service-a service-b`
- **Any CLI tool** that needs formatted table output

## Comparison with Interactive Components

| Feature | `table` package | `tui` interactive components |
|---------|----------------|------------------------------|
| Use case | Static CLI output | Interactive TUI apps |
| Dependencies | None (stdlib only) | Bubble Tea |
| Keyboard nav | No | Yes |
| Mouse support | No | Yes |
| Updates | No | Yes (real-time) |
| Complexity | Very simple | Full framework |

## Design Philosophy

The `table` package follows these principles:

1. **Zero dependencies** - Only uses Go stdlib
2. **Print and exit** - For non-interactive CLI tools
3. **Beautiful by default** - Unicode borders, proper spacing
4. **Fallback support** - ASCII mode for restricted terminals
5. **Simple API** - Minimal methods, intuitive usage

## Performance

The table renderer is optimized for CLI output:

- **Memory efficient** - Single string builder allocation
- **Fast rendering** - Benchmarked at ~50µs for 100 rows
- **Minimal allocations** - Reuses buffers where possible

## Future Enhancements

Potential additions:

- Column alignment (left, right, center)
- Color support for cells
- Column width limits with truncation
- Footer rows
- Multi-line cell support
- Custom cell padding
- Row highlighting

## Contributing

Contributions welcome! Please ensure:

- All tests pass: `go test ./table`
- Code is formatted: `go fmt ./table`
- Examples work: `go run ./table/example`

## License

MIT
