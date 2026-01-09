# Modal Demo

This demo showcases the Modal component for displaying overlay dialogs with different interaction patterns.

## Features Demonstrated

- **Three modal types**:
  - **Alert**: Information message with OK button
  - **Confirm**: Yes/No question dialogs
  - **Input**: Text input with OK/Cancel buttons
- **Centered overlay** with backdrop
- **Keyboard navigation** between buttons (Tab/Shift+Tab)
- **Text wrapping** for long messages
- **Focus management** with visual feedback
- **Callback support** for user actions
- **Command palette integration**

## Running the Demo

```bash
go run main.go
```

## Keyboard Controls

### Modal Controls
- **Tab / →** - Next button
- **Shift+Tab / ←** - Previous button
- **Enter** - Confirm selected button
- **Esc** - Cancel/close modal

### Demo Controls
- **1** - Show Alert modal
- **2** - Show Confirm modal
- **3** - Show Input modal
- **Ctrl+K** - Open command palette
- **q** - Quit

## Modal Types

### 1. Alert Modal

Information display with a single OK button:

```go
modal.ShowAlert(
    "Information",
    "This is an important message.",
    func() tea.Cmd {
        // Handle OK action
        return nil
    },
)
```

**Use cases**:
- Success notifications
- Error messages
- Information display
- Completion notifications

### 2. Confirm Modal

Question with Yes/No buttons:

```go
modal.ShowConfirm(
    "Confirmation",
    "Are you sure you want to proceed?",
    func() tea.Cmd {
        // Handle Yes action
        return nil
    },
    func() tea.Cmd {
        // Handle No action
        return nil
    },
)
```

**Use cases**:
- Delete confirmations
- Destructive action warnings
- Permission requests
- Save/discard prompts

### 3. Input Modal

Text input with OK/Cancel buttons:

```go
modal.ShowInput(
    "User Input",
    "Please enter your name:",
    "John Doe", // placeholder
    func(value string) tea.Cmd {
        // Handle OK with input value
        return nil
    },
    func() tea.Cmd {
        // Handle Cancel
        return nil
    },
)
```

**Use cases**:
- Name/label input
- Search queries
- Configuration values
- Quick text entry

## Implementation Details

### Modal Structure

```go
type Modal struct {
    visible    bool
    focused    bool
    modalType  ModalType
    title      string
    message    string
    buttons    []ModalButton
    selected   int
    textInput  textinput.Model
    hasInput   bool
}

type ModalButton struct {
    Label  string
    Action func(string) tea.Cmd
}
```

### Button Navigation

Users can navigate between buttons using Tab or arrow keys:

```go
case tea.KeyTab, tea.KeyRight:
    if m.selected < len(m.buttons)-1 {
        m.selected++
    } else {
        m.selected = 0 // Wrap around
    }
```

### Text Wrapping

Long messages automatically wrap to fit the modal width:

```go
func wrapText(text string, width int) []string {
    words := strings.Fields(text)
    // Build lines that fit within width
    // ...
}
```

### Centered Overlay

Modals are centered on screen with backdrop:

```go
modalWidth := min(60, m.width-8)
modalHeight := min(contentHeight+4, m.height-6)
startX := (m.width - modalWidth) / 2
startY := max(3, (m.height-modalHeight)/3)
```

## Usage in Your Application

### Basic Usage

```go
// Create modal
modal := tui.NewModal()
app.AddComponent(modal)

// Show alert
modal.ShowAlert("Title", "Message", func() tea.Cmd {
    return nil
})

// Show confirmation
modal.ShowConfirm("Title", "Message",
    func() tea.Cmd { /* Yes */ return nil },
    func() tea.Cmd { /* No */ return nil })

// Show input
modal.ShowInput("Title", "Message", "placeholder",
    func(value string) tea.Cmd { /* OK */ return nil },
    func() tea.Cmd { /* Cancel */ return nil })
```

### Custom Buttons

```go
modal := tui.NewModal(
    tui.WithModalTitle("Custom Dialog"),
    tui.WithModalMessage("Choose an option:"),
    tui.WithModalButtons([]tui.ModalButton{
        {
            Label: "Option 1",
            Action: func(s string) tea.Cmd {
                // Handle option 1
                return nil
            },
        },
        {
            Label: "Option 2",
            Action: func(s string) tea.Cmd {
                // Handle option 2
                return nil
            },
        },
        {
            Label: "Cancel",
            Action: func(s string) tea.Cmd {
                // Handle cancel
                return nil
            },
        },
    }),
)
```

### Checking Visibility

```go
if modal.IsVisible() {
    // Modal is currently displayed
}
```

### Manual Control

```go
// Show modal
modal.SetTitle("Dynamic Title")
modal.SetMessage("Dynamic message")
modal.Show()

// Hide modal
modal.Hide()
```

## Integration with Command Palette

The demo shows how modals work alongside other components like the command palette:

```go
case tea.KeyMsg:
    switch msg.String() {
    case "1":
        if !m.modal.IsVisible() && !m.commandPalette.IsVisible() {
            // Only show modal if nothing else is visible
            modal.ShowAlert(...)
        }
    }
```

## Visual Design

### Modal Structure

```
╭─── Title ────────────────────────────────────────────╮
│                                                       │
│  This is the message text that can wrap across       │
│  multiple lines to fit within the modal width.       │
│                                                       │
│  [Optional text input field]                         │
│                                                       │
│              [ Button 1 ]  [ Button 2 ]              │
│                                                       │
└─ Tab: navigate · Enter: confirm · Esc: cancel ───────┘
```

### Focus States

- **Selected button**: Inverted colors `\033[7m[ Label ]\033[0m`
- **Normal button**: Dimmed brackets `\033[2m[ \033[0mLabel\033[2m ]\033[0m`
- **Input focus**: Cursor blinks in text field

## Future Enhancements

Potential improvements for the Modal component:

- Custom modal sizes
- Scrollable content for long messages
- Multi-line text input (textarea)
- Dropdown/select inputs
- Form validation
- Modal stacking (multiple modals)
- Animations (fade in/out)
- Custom color schemes
- Progress/loading modals
- Rich content (lists, tables)
