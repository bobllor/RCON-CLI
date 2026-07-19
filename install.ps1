<#
    .DESCRIPTION
    Install gorcon onto a Windows device.

    .AUTHOR
    bobllor
#>

$file_name="gorcon.windows.amd64.zip"

$baseUrl = "https://github.com/bobllor/rcon-cli/releases/latest"
$url="$base_url/download/$file_name"


# speeds up webrequest download, just windows things
$ProgressPreference = 'SilentlyContinue'

$tempPath = "$env:temp\gorcontmp"

if(!(Test-Path "$tempPath")){
    mkdir "$tempPath"
}

# will just let this fail if an error occurs, better for
# user logging
Invoke-WebRequest -uri "$url" -usebasicparsing -o "$tempPath\$file_name"

$appPath = "$env:localappdata\gorcon"

if(!(Test-Path "$appPath")){
    mkdir "$appPath"
}

Expand-Archive "$tempPath\$file_name" "$appPath"

if(Test-Path "$tempPath"){
    rm -Force "$tempPath"
}