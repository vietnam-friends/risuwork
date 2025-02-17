package model

import "time"

const UserTypeCS = "CS"
const UserTypeCL = "CL"

type CSUser struct {
	ID       int    `json:"id" fake:"skip"`
	Email    string `json:"email" fake:"{email}"`
	Password string `json:"password" fake:"{password}"`
	Name     string `json:"name" fake:"{name}"`
	UserType string `json:"user_type" fake:"CS"`
}

type Company struct {
	ID         int    `json:"id"`
	Name       string `json:"name" fake:"{company}"`
	IndustryID string `json:"industry_id" fake:"{industry_id}"`
	Industry   string `json:"industry" fake:"skip"`
}

type CLUser struct {
	ID        int     `json:"id" fake:"skip"`
	Email     string  `json:"email" fake:"{email}"`
	Password  string  `json:"password" fake:"{password}"`
	Name      string  `json:"name" fake:"{name}"`
	UserType  string  `json:"user_type" fake:"CL"`
	CompanyID int     `json:"company_id" fake:"skip"`
	Company   Company `json:"-" fake:"skip"`
}

type Job struct {
	ID           int       `json:"id" fake:"skip"`
	Title        string    `json:"title" fake:"{jobtitle}"`
	Description  string    `json:"description" fake:"{sentence:20}"`
	Salary       int       `json:"salary" fake:"{number:1000,1000000}"`
	Tags         string    `json:"tags" fake:"{tags}"`
	IsActive     bool      `json:"is_active" fake:"true"`
	CreateUserID int       `json:"create_user_id" fake:"skip"`
	Company      Company   `json:"company" fake:"skip"`
	CreatedAt    time.Time `json:"created_at" fake:"skip"`
	UpdatedAt    time.Time `json:"updated_at" fake:"skip"`
}

type JobSearchResponse struct {
	Jobs        []Job `json:"jobs"`
	Page        int   `json:"page"`
	HasNextPage bool  `json:"has_next_page"`
}

type Application struct {
	ID        int       `json:"id"`
	JobID     int       `json:"job_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Job       Job
}

type JobWithApplication struct {
	Job
	Applications []Application `json:"applications"`
}

type ApplicationsResponse struct {
	Applications []Application `json:"applications"`
	Page         int           `json:"page"`
	HasNextPage  bool          `json:"has_next_page"`
}

type JobsResponse struct {
	Jobs        []Job `json:"jobs"`
	Page        int   `json:"page"`
	HasNextPage bool  `json:"has_next_page"`
}
