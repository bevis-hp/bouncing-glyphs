Output demo/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 650
Set Framerate 30
Set WindowBar Colorful
Set Padding 10

# Show glyph intake from piped stdin
Type "openssl rand -hex 32 | glyphfall"
Enter
Sleep 17s

# Quit
Type "q"
Sleep 500ms
