# Serves a copy of the real stable components.xml from http://localhost:<Port>/
# in which only the "core-manager" component points at a local core-manager.zip.
# This lets the OLD C# manager install the new Go manager exactly the way a real
# release would ("push it as a component through the old manager").
#
#   powershell -ExecutionPolicy Bypass -File rehearsal.ps1 -Zip C:\fptest\bin\core-manager.zip
#
# Then point Flashpoint at it (legacy fpm):  Manager\fpm.exe source http://localhost:8642/components.xml
# Stop with Ctrl+C. Only changes files in -Work. Windows PowerShell 5.1 is enough.
param(
    [Parameter(Mandatory = $true)][string]$Zip,
    [int]$Port = 8642,
    [string]$Work = "C:\fptest\rehearsal",
    [string]$Source = "https://nexus-dev.flashpointarchive.org/repository/stable/components.xml"
)
$ErrorActionPreference = "Stop"

Add-Type -TypeDefinition @"
public static class Crc32 {
    public static string Of(byte[] b) {
        uint[] t = new uint[256];
        for (uint i = 0; i < 256; i++) { uint c = i; for (int k = 0; k < 8; k++) c = (c & 1) != 0 ? 0xEDB88320u ^ (c >> 1) : c >> 1; t[i] = c; }
        uint crc = 0xFFFFFFFFu;
        foreach (byte x in b) crc = t[(crc ^ x) & 0xFF] ^ (crc >> 8);
        return (crc ^ 0xFFFFFFFFu).ToString("x8");
    }
}
"@

New-Item -ItemType Directory -Force -Path $Work | Out-Null
Copy-Item -Force $Zip (Join-Path $Work "core-manager.zip")
$bytes = [IO.File]::ReadAllBytes((Join-Path $Work "core-manager.zip"))
$hash = [Crc32]::Of($bytes)
$installSize = 0
Add-Type -AssemblyName System.IO.Compression.FileSystem
$z = [IO.Compression.ZipFile]::OpenRead((Join-Path $Work "core-manager.zip"))
foreach ($e in $z.Entries) { $installSize += $e.Length }
$z.Dispose()

# Edit the raw text (no XML re-serialisation: the C# manager crashes on
# anything but category/component elements, and keeps working with new values).
$xml = (New-Object Net.WebClient).DownloadString($Source)
$xml = $xml -replace '<list url="[^"]*"', "<list url=`"http://localhost:$Port/`""
$pattern = '(<component id="manager"[^>]*?)download-size="\d+" install-size="\d+"([^>]*?)hash="[0-9a-f]{8}"'
if ($xml -notmatch $pattern) { throw "core-manager entry not found in $Source" }
$xml = $xml -replace $pattern, "`${1}download-size=`"$($bytes.Length)`" install-size=`"$installSize`"`${2}hash=`"$hash`""
[IO.File]::WriteAllText((Join-Path $Work "components.xml"), $xml, (New-Object Text.UTF8Encoding $false))
Write-Host "core-manager.zip: hash $hash, download-size $($bytes.Length), install-size $installSize"

$l = New-Object Net.HttpListener
$l.Prefixes.Add("http://localhost:$Port/")
$l.Start()
Write-Host "Serving $Work on http://localhost:$Port/ (Ctrl+C to stop)"
try {
    while ($l.IsListening) {
        $ctx = $l.GetContext()
        $name = [IO.Path]::GetFileName($ctx.Request.Url.AbsolutePath)  # ignores ?h=...
        $file = Join-Path $Work $name
        Write-Host "$($ctx.Request.HttpMethod) $($ctx.Request.Url.PathAndQuery)"
        if ($name -and (Test-Path -PathType Leaf $file)) {
            $data = [IO.File]::ReadAllBytes($file)
            $ctx.Response.ContentLength64 = $data.Length
            if ($ctx.Request.HttpMethod -ne "HEAD") { $ctx.Response.OutputStream.Write($data, 0, $data.Length) }
        } else {
            $ctx.Response.StatusCode = 404
        }
        $ctx.Response.Close()
    }
} finally {
    $l.Stop()
}
