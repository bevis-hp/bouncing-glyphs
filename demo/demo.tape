Output demo/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 650
Set Framerate 30
Set WindowBar Colorful
Set Padding 10

# Build the binary silently so the recording starts cleanly
Hide
Type "go build -o /tmp/bouncing-glyphs-demo . && clear"
Enter
Sleep 6s

# Start the animation
Show
Type "/tmp/bouncing-glyphs-demo -count 12"
Enter
Sleep 5s

# Spawn a few more glyphs interactively
Space
Sleep 2500ms
Space
Sleep 2500ms
Space
Sleep 3s

# Quit
Type "q"
Sleep 500ms

# Cleanup
Hide
Type "rm /tmp/bouncing-glyphs-demo"
Enter
