$IP = "100.102.191.66"
$TIME = 60
$LOG_FILE = "matrix_results.txt"


function Run-Matrix {
    Write-Output "=========================================="
Write-Output "  PHASE A: INGESTION CEILING (PROTOBUF)   "
Write-Output "=========================================="
$workers = @(100, 250, 500, 750, 1000)
foreach ($w in $workers) {
    Write-Output "`n---> Running Ingest with $w Workers"
    go run scripts/benchmark.go -ip $IP -mode ingest -payload proto -time $TIME -workers $w -qos 1
    
    # Let the OS clear sockets and the broker GC to breathe between tests
    Start-Sleep -Seconds 30
}

Write-Output "`n=========================================="
Write-Output "  PHASE B: FANOUT QUEUE LIMITS (PROTOBUF) "
Write-Output "=========================================="
$subs = @(100, 500, 1000, 1500, 2000)
foreach ($s in $subs) {
    Write-Output "`n---> Running Fanout with 50 Pubs and $s Subs"
    go run scripts/benchmark.go -ip $IP -mode fanout -payload proto -time $TIME -pubs 50 -workers $s -qos 1
    
    Start-Sleep -Seconds 30
}

Write-Output "`n=========================================="
Write-Output "  PHASE C: THUNDERING HERD (CONNECTION SPIKE) "
Write-Output "=========================================="
Write-Output "`n---> Hitting broker with 5,000 simultaneous connections..."
go run scripts/benchmark.go -ip $IP -mode ingest -payload proto -time 15 -workers 5000 -qos 1

    Write-Output "`n[MATRIX COMPLETE] Results have been saved to $LOG_FILE. Please copy the contents of that file and paste them here!"
}

Write-Host "Running Benchmark Matrix... (Saving cleanly to $LOG_FILE)"
Run-Matrix *>&1 | Out-File -FilePath $LOG_FILE -Encoding utf8
Write-Host "Done! Please check $LOG_FILE"
