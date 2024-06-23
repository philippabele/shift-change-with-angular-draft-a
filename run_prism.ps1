# Get the directory where the script is located
$scriptPath = $PSScriptRoot

# Construct the path to the YAML file
$yamlFilePath = Join-Path -Path $scriptPath -ChildPath "backend.yaml"

# Check if Prism is installed
$prismInstalled = Get-Command prism -ErrorAction SilentlyContinue
if (-not $prismInstalled) {
    Write-Output "Prism is not installed, installing..."
    npm install -g @stoplight/prism-cli
} else {
    Write-Output "Prism is already installed."
}

# Run the Prism command with the correct path to the YAML file
prism mock -h 0.0.0.0 -p 4010 -d true --json-schema-faker-fillProperties true $yamlFilePath