# 🎵 Termify

> A sleek, feature-rich Spotify TUI (Terminal User Interface) client written in Go

![Termify Screenshot](https://github.com/dietzy1/termify/raw/main/assets/screenshot.png)

## ✨ Features

- **Beautiful Terminal Interface**: Enjoy a modern, responsive UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Spotify Integration**: Full access to your Spotify library and playback controls
- **Playlist Management**: Browse and play your playlists directly from the terminal
- **Playback Controls**: Play, pause, skip tracks, adjust volume, and more
- **Keyboard Shortcuts**: Efficient navigation with intuitive keybindings
- **Authentication**: Secure OAuth2 PKCE flow for Spotify API authentication
- **Cross-Platform**: Works on macOS, Linux, and Windows

## 🚀 Installation

### Prerequisites

- Go 1.16 or higher
- A Spotify account (Premium required for full playback functionality)

### From Source

```bash
# Clone the repository
git clone https://github.com/dietzy1/termify.git
cd termify

# Build the application
go build -o termify

# Run Termify
./termify
```

### Using Go Install

```bash
go install github.com/dietzy1/termify@latest
```

## 🔑 Authentication

On first run, Termify will:

1. Open a browser window for Spotify authentication
2. Ask you to log in to your Spotify account
3. Request necessary permissions
4. Redirect back to the application

Your credentials are securely stored for future sessions.

## 🎮 Usage

### Navigation

- **Tab**: Cycle through sections
- **Arrow Keys**: Navigate within sections
- **Enter**: Select items
- **q**: Quit the application

### Playback Controls

- **Space**: Play/Pause
- **n**: Next track
- **p**: Previous track
- **s**: Toggle shuffle
- **r**: Toggle repeat mode
- **+/-**: Adjust volume

## 📦 Project Structure

```
termify/
├── internal/           # Internal packages
│   ├── authentication/ # Spotify authentication
│   ├── config/         # Application configuration
│   └── tui/            # Terminal user interface
│       ├── tui.go              # Main entry model
│       ├── application.go      # Main application model
│       ├── playbackControls.go # Playback control UI
│       ├── spotifyState.go     # Spotify API integration
│       └── ...                 # Other UI components
└── main.go             # Application entry point
```

## 🛠️ Development

### Requirements

- Go 1.16+
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
- [Spotify Web API Go](https://github.com/zmb3/spotify) for Spotify integration

### Building from Source

```bash
# Clone the repository
git clone https://github.com/dietzy1/termify.git
cd termify

# Install dependencies
go mod download

# Run in development mode
go run main.go
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.
