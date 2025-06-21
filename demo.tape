# Galick Demo
Output galick-demo.gif

Set Shell zsh
Set FontSize 18
Set Width 1200
Set Height 600
Set Padding 20

# Initialize a new config
Type "galick init"
Enter
Sleep 2s

# Show the generated config
Type "cat loadtest.yaml | head -n 20"
Enter
Sleep 3s

# Run a load test
Type "galick run simple"
Enter
Sleep 5s

# Show reports
Type "cat output/dev/simple/report.md"
Enter
Sleep 5s

# Generate a different report format
Type "galick report simple --format json"
Enter
Sleep 2s

# Show the JSON report
Type "cat output/dev/simple/report.json | jq"
Enter
Sleep 3s
