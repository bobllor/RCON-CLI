<#
    .DESCRIPTION
    Uninstall gorcon and remove the PATH from the device.
#>

$appPath = "$env:localappdata"
$gorconPath = "$appPath\Programs\gorcon"

$hasUninstalled = $false

if(test-path "$gorconPath"){
    echo "Removing gorcon..."
    rm -force -recurse "$gorconPath"

    $hasUninstalled = $true
    echo "Removed gorcon"
}

$paths = [System.Environment]::GetEnvironmentVariable("PATH", "User")

if($paths.split(";") -contains "$gorconPath"){
    echo "Removing PATH variable..."
    $newPaths = $paths.replace("$gorconPath;", "")

    [System.Environment]::SetEnvironmentVariable("PATH", "$newPaths", "User")
    echo "PATH variable removed"
}

if($hasUninstalled){
    echo "gorcon has been uninstalled"
}else{
    echo "gorcon installation not found"
}