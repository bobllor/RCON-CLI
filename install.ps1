<#
    .DESCRIPTION
    Install gorcon onto a Windows device.

    .AUTHOR
    bobllor
#>

$fileName="gorcon.windows.amd64.zip"

$baseUrl = "https://github.com/bobllor/rcon-cli/releases/latest"
$url="$baseUrl/download/$fileName"

echo "Starting gorcon installation..."
echo ""

$tempPath = "$env:temp\gorcontmp"

if(!(Test-Path "$tempPath")){
    mkdir "$tempPath" | out-null
}

# will just let this fail if an error occurs, better for
# user logging
echo "Downloading $fileName..."
try{
    Invoke-WebRequest -uri "$url" -usebasicparsing -o "$tempPath\$fileName"
}catch{
    Write-Error $_
    exit 1
}

$appPath = "$env:localappdata\Programs\gorcon"

if(!(Test-Path "$appPath")){
    mkdir "$appPath" | out-null 
}

echo "Extracting $fileName..."
Expand-Archive "$tempPath\$fileName" "$appPath" -Force
echo "Extracted files to $appPath"

$userEnvPaths = [System.Environment]::GetEnvironmentVariable("PATH", "User")

if(!($userEnvPaths.split(";") -contains "$appPath")){
    echo "Configuring PATH..."
    $userEnvPaths += "$appPath;"

    echo "$userEnvPaths"
    [System.Environment]::SetEnvironmentVariable("PATH", "$userEnvPaths", "User")

    echo "PATH configured ($appPath)"
}else{
    echo "PATH already configured, skipping"
}

if(Test-Path "$tempPath"){
    rm -Force -Recurse "$tempPath"
}

echo ""
echo "Successfully installed!"
echo 'Open a new terminal and run "gorcon" to get started'