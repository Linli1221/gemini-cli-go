$response = Invoke-WebRequest -Uri 'http://localhost:8080/v1/chat/completions' -Method POST -Headers @{'Content-Type'='application/json'; 'Authorization'='Bearer sk-1234567890abcdef1234567890abcdef'} -InFile 'test_request.json'
Write-Host "Status Code: $($response.StatusCode)"
Write-Host "Content:"
$response.Content | ConvertFrom-Json | ConvertTo-Json -Depth 10