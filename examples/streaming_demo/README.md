# Streaming ToolBlock Demo

This demo showcases real-time streaming output with animated spinners for multiple ToolBlocks.

## Features Demonstrated

- **ActivityBar** with animated spinner and progress updates
- **Multiple streaming ToolBlocks** running concurrently
- **Status indicators**: Running (⠋ spinner), Complete (✓), Error (✗)
- **Real-time output** streaming line-by-line
- **Stage progression**: Tests → Build → Error Demo
- **Restart capability**: Press 'r' to restart the demonstration

## Running the Demo

```bash
go run main.go
```

## What You'll See

### Stage 1: Running Tests
```
✳ Running tests… (esc to interrupt · 1s · ✓ Tests passed) ⠋

⏺ Bash(go test -v) ⠋
  ⎿  === RUN   TestApplicationCreation
     --- PASS: TestApplicationCreation (0.00s)
     === RUN   TestComponentAddition
     … +12 lines (ctrl+o to expand)
```

The ActivityBar shows "Running tests..." with an animated spinner (✳ ⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏).
The test block streams output line-by-line with its own spinner.

### Stage 2: Building Project
```
✳ Building project… (✓ Build complete) ⠋

⏺ Bash(go test -v) ✓     ← Tests complete (green)
  ⎿  [output collapsed]

⏺ Bash(go build .) ⠋     ← Build streaming (cyan)
  ⎿  Building project...
     Compiling main.go
     … +3 lines (ctrl+o to expand)
```

### Stage 3: Error Demo
```
✳ Running failing tests… ⠋

⏺ Bash(go test -v) ✓
⏺ Bash(go build .) ✓
⏺ Bash(go test ./broken) ✗    ← Error state (red)
  ⎿  === RUN   TestInvalidInput
         main_test.go:42: Expected nil, got error
     --- FAIL: TestInvalidInput (0.02s)
     … +2 lines
```

## Keyboard Controls

- **r** - Restart demonstration from beginning
- **q** - Quit
- **Tab** - Focus next component
- **Ctrl+O** - Expand/collapse focused ToolBlock

## Implementation Details

### Spinner Animation
Each component has its own animated spinner that updates at 100ms intervals:
- ActivityBar: ✳ with braille spinner (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)
- ToolBlock: Icon (⏺) with spinner while streaming

### Status Colors
- **Cyan (⏺ ⠋)** - Running/Streaming
- **Green (⏺ ✓)** - Complete/Success
- **Red (⏺ ✗)** - Error/Failed
- **Yellow (⏺ ⚠)** - Warning (not shown in this demo)

### Real-time Streaming
Output is added line-by-line using:
```go
block.AppendLine("single line")
block.AppendLines([]string{"line 1", "line 2"})
```

### Status Management
```go
// Start streaming
block := tui.NewToolBlock("Bash", "cmd", []string{}, tui.WithStreaming())

// Add output as it arrives
block.AppendLine("output line")

// Mark complete
block.SetStatus(tui.StatusComplete)  // Green ✓
block.SetStatus(tui.StatusError)     // Red ✗
```

## Code Structure

- **Model**: Manages application state, stages, and line-by-line streaming
- **Init**: Starts ActivityBar and streaming cycle
- **Update**: Handles streaming messages, stage transitions, and restarts
- **View**: Renders all components with current state

## Testing

The demo includes a restart feature that resets all components to their initial state:
1. Press 'r' at any time
2. All blocks reset to empty streaming state
3. Stage counter resets to 0
4. ActivityBar restarts with "Running tests..."
5. Streaming begins again from the start
