# VHS (https://github.com/charmbracelet/vhs) script for Termify demo
# Set the output file
Output termify-demo.gif

# Set the terminal dimensions and styling
Set FontSize 16
Set Width 1920
Set Height 1080
Set TypingSpeed 100ms
Set Framerate 30
Set Padding 0
Set PlaybackSpeed 1
Set LoopOffset 5

# Setup: Start with a clean terminal and launch application before showing
Hide
Type "go build -o termify .. && clear"
Enter
Type "./termify"
Enter
Sleep 3s
Show


# Navigate through the interface (adjust based on your actual navigation commands)
Down 3
Sleep 1s
Tab
Sleep 1s
Enter
Down 2
Sleep 1s
Type "q"
Sleep 1s
Down 2
Sleep 1s
Type "q"
Sleep 1s

# Show search functionality (adjust based on your actual commands)
Type "/"
Type "hello darkness"
Sleep 1s
Tab 2
Sleep 1s

# Select a track from the search results
Down 2
Sleep 1s
Enter
Sleep 2s
Down 2
Sleep 1s

# Open queue view
Type "v"
Sleep 2s

# Show help menu
Type "?"
Sleep 2s
Escape
Sleep 1s


# Properly end the recording
Hide
Ctrl+C
Sleep 2s

# Cleanup (only if you want this to happen after recording)
Type "rm termify"
Enter