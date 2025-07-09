try {
    $response = Invoke-WebRequest -Uri 'http://localhost:8080/v1/chat/completions' -Method POST -Headers @{'Content-Type'='application/json'; 'Authorization'='Bearer sk-1234567890abcdef1234567890abcdef'} -InFile 'test_request.json'
    Write-Host "Status Code: $($response.StatusCode)"
    Write-Host "Response Content:"
    Write-Host $response.Content
} catch {
    Write-Host "Error: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Error Response Body: $responseBody"
    }
}