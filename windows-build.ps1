New-Item -ItemType Directory -Force dist | Out-Null

$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; Output = ".\\dist\\pingtool-windows-amd64.exe" },
    @{ GOOS = "windows"; GOARCH = "arm64"; Output = ".\\dist\\pingtool-windows-arm64.exe" },
    @{ GOOS = "linux";   GOARCH = "amd64"; Output = ".\\dist\\pingtool-linux-amd64" },
    @{ GOOS = "linux";   GOARCH = "arm64"; Output = ".\\dist\\pingtool-linux-arm64" },
    @{ GOOS = "darwin";  GOARCH = "amd64"; Output = ".\\dist\\pingtool-darwin-amd64" },
    @{ GOOS = "darwin";  GOARCH = "arm64"; Output = ".\\dist\\pingtool-darwin-arm64" }
)

foreach ($target in $targets) {
    $env:CGO_ENABLED = "0"
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    go build -o $target.Output .
}

Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue