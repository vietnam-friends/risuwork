package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/isucon/isucandar/agent"
)

// POST /initialize
func PostInitialize(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.POST("/api/initialize", nil)
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}

// POST /finalize
func PostFinalize(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.POST("/api/finalize", nil)
	if err != nil {
		return nil, err
	}
	return ag.Do(ctx, req)
}

// POST /cs/signup
func PostCSSignup(ctx context.Context, ag *agent.Agent, email, password, name string) (*http.Response, error) {
	type SignupRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	json, err := json.Marshal(SignupRequest{
		Email:    email,
		Password: password,
		Name:     name,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.POST("/api/cs/signup", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

// POST /cs/login
func PostCSLogin(ctx context.Context, ag *agent.Agent, email, password string) (*http.Response, error) {
	type SignupRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json, err := json.Marshal(SignupRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.POST("/api/cs/login", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

type JobSearchQuery struct {
	Keyword    string
	MinSalary  int
	MaxSalary  int
	Tag        string
	IndustryID string
	Page       *int
}

// GET /cs/job_search
func GetCSJobSearch(ctx context.Context, ag *agent.Agent, query JobSearchQuery) (*http.Response, error) {
	target := fmt.Sprintf("/api/cs/job_search?keyword=%s&min_salary=%d&max_salary=%d&tag=%s&industry_id=%s", query.Keyword, query.MinSalary, query.MaxSalary, query.Tag, query.IndustryID)
	if query.Page != nil {
		target = fmt.Sprintf("%s&page=%d", target, *query.Page)
	}

	req, err := ag.GET(target)
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}

// POST /cs/logout
func PostCSLogout(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.POST("/api/cs/logout", nil)
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}

// POST /cs/application
func PostCSApplication(ctx context.Context, ag *agent.Agent, jobID int) (*http.Response, error) {
	type ApplicationRequest struct {
		JobID int `json:"job_id"`
	}

	json, err := json.Marshal(ApplicationRequest{
		JobID: jobID,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.POST("/api/cs/application", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

// GET /cs/applications
func GetCSApplications(ctx context.Context, ag *agent.Agent, page *int) (*http.Response, error) {
	target := "/api/cs/applications"
	if page != nil {
		target = fmt.Sprintf("%s?page=%d", target, *page)
	}
	req, err := ag.GET(target)
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}

// POST /cl/company
func PostCLCompany(ctx context.Context, ag *agent.Agent, name, industryID string) (*http.Response, error) {
	type CompanyRequest struct {
		Name       string `json:"name"`
		IndustryID string `json:"industry_id"`
	}

	json, err := json.Marshal(CompanyRequest{
		Name:       name,
		IndustryID: industryID,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.POST("/api/cl/company", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

// POST /cl/signup
func PostCLSignup(ctx context.Context, ag *agent.Agent, email, password, name string, companyID int) (*http.Response, error) {
	type SignupRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Name      string `json:"name"`
		CompanyID int    `json:"company_id"`
	}

	json, err := json.Marshal(SignupRequest{
		Email:     email,
		Password:  password,
		Name:      name,
		CompanyID: companyID,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.POST("/api/cl/signup", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

// POST /cl/login
func PostCLLogin(ctx context.Context, ag *agent.Agent, email, password string) (*http.Response, error) {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	json, err := json.Marshal(LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.POST("/api/cl/login", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

// POST /cl/logout
func PostCLLogout(ctx context.Context, ag *agent.Agent) (*http.Response, error) {
	req, err := ag.POST("/api/cl/logout", nil)
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}

// POST /cl/job
func PostCLJob(ctx context.Context, ag *agent.Agent, title string, description string, salary int, tag string) (*http.Response, error) {
	type JobRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Salary      int    `json:"salary"`
		Tags        string `json:"tags"`
	}

	json, err := json.Marshal(JobRequest{
		Title:       title,
		Description: description,
		Salary:      salary,
		Tags:        tag,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.POST("/api/cl/job", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

// PATCH /cl/job/{job_id}
func PatchCLJob(ctx context.Context, ag *agent.Agent, jobID int, title *string, description *string, salary *int, tag *string, isActive *bool) (*http.Response, error) {
	type UpdateJobRequest struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Salary      *int    `json:"salary"`
		Tags        *string `json:"tags"`
		IsActive    *bool   `json:"is_active"`
	}

	json, err := json.Marshal(UpdateJobRequest{
		Title:       title,
		Description: description,
		Salary:      salary,
		Tags:        tag,
		IsActive:    isActive,
	})
	if err != nil {
		return nil, err
	}

	req, err := ag.PATCH(fmt.Sprintf("/api/cl/job/%d", jobID), bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return ag.Do(ctx, req)
}

// POST /cl/job/{job_id}/archive
func POSTCLJobArchive(ctx context.Context, ag *agent.Agent, jobID int) (*http.Response, error) {
	req, err := ag.POST(fmt.Sprintf("/api/cl/job/%d/archive", jobID), nil)
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}

// GET /cl/job/{job_id}
func GetCLJob(ctx context.Context, ag *agent.Agent, jobID int) (*http.Response, error) {
	req, err := ag.GET(fmt.Sprintf("/api/cl/job/%d", jobID))
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}

// GET /cl/jobs
func GetCLJobs(ctx context.Context, ag *agent.Agent, page *int) (*http.Response, error) {
	target := "/api/cl/jobs"
	if page != nil {
		target = fmt.Sprintf("%s?page=%d", target, *page)
	}
	req, err := ag.GET(target)
	if err != nil {
		return nil, err
	}

	return ag.Do(ctx, req)
}
