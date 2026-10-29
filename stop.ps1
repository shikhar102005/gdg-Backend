# =====================================================================
#   Campus Sensor Monitoring & Alerts - PowerShell Stop Script
# =====================================================================

Write-Host "Stopping Campus Sensor services on ports 8081 and 5173..." -ForegroundColor Yellow

$ports = @(8081, 5173)
$stopped = 0

foreach ($port in $ports) {
    $conns = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($conns) {
        $pidsToKill = $conns.OwningProcess | Select-Object -Unique
        foreach ($p in $pidsToKill) {
            Write-Host "Terminating process $p listening on port $port..." -ForegroundColor Cyan
            Stop-Process -Id $p -Force -ErrorAction SilentlyContinue
            $stopped++
        }
    }
}

if ($stopped -gt 0) {
    Write-Host "Services successfully stopped." -ForegroundColor Green
} else {
    Write-Host "No active processes found on ports 8081 or 5173." -ForegroundColor Gray
}
