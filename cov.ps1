go test ./internal/automation/... -coverprofile=cov_automation.out -v 2>&1 | Out-Null
go tool cover -func=cov_automation.out

go test ./internal/browser/... -coverprofile=cov_browser.out -v 2>&1 | Out-Null
go tool cover -func=cov_browser.out

go test ./internal/scheduler/... -coverprofile=cov_scheduler.out -v 2>&1 | Out-Null
go tool cover -func=cov_scheduler.out

go test ./internal/workflow/... -coverprofile=cov_workflow.out -v 2>&1 | Out-Null
go tool cover -func=cov_workflow.out

go test ./internal/social/... -coverprofile=cov_social.out -v 2>&1 | Out-Null
go tool cover -func=cov_social.out

go test . -coverprofile=cov_main.out -v 2>&1 | Out-Null
go tool cover -func=cov_main.out
