package action

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"risuwork-benchmarker/scenario/api"
	"risuwork-benchmarker/scenario/fixture"
	"risuwork-benchmarker/scenario/model"
	"strings"

	"github.com/isucon/isucandar/agent"
)

func CreateCSUserWithAuth(ctx context.Context, ag *agent.Agent) (*model.CSUser, error) {
	// dummy user 作成
	cs := fixture.GenerateCSUser()

	// dummy user を登録
	resp, err := api.PostCSSignup(ctx, ag, cs.Email, cs.Password, cs.Name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return cs, nil
}

func CreateCompany(ctx context.Context, ag *agent.Agent) (*model.Company, error) {
	// dummy company 作成
	c := fixture.GenerateCompany()

	// dummy company を登録
	resp, err := api.PostCLCompany(ctx, ag, c.Name, c.IndustryID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create company (status:%s), (body: %s)", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(c); err != nil {
		return nil, err
	}
	return c, nil
}

func CreateCLUserWithAuth(ctx context.Context, ag *agent.Agent) (*model.CLUser, error) {
	// dummy company を作成・登録
	c, err := CreateCompany(ctx, ag)
	if err != nil {
		return nil, err
	}

	// 登録したcompanyで dummy user 作成
	cl := fixture.GenerateCLUser()
	cl.CompanyID = c.ID
	cl.Company = *c

	// duumy user を登録
	resp, err := api.PostCLSignup(ctx, ag, cl.Email, cl.Password, cl.Name, cl.CompanyID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return cl, nil
}

func CreateCLUserForSpecifiedComapnyWithAuth(ctx context.Context, ag *agent.Agent, company *model.Company) (*model.CLUser, error) {

	// 登録したcompanyで dummy user 作成
	cl := fixture.GenerateCLUser()
	cl.CompanyID = company.ID
	cl.Company = *company

	// duumy user を登録
	resp, err := api.PostCLSignup(ctx, ag, cl.Email, cl.Password, cl.Name, cl.CompanyID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return cl, nil
}

func CreateJobWithNewUser(ctx context.Context, ag *agent.Agent) (*model.CLUser, *model.Job, error) {
	// dummy user 作成・登録
	cl, err := CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		return nil, nil, err
	}

	// dummy job 作成
	job := fixture.GenerateJob()

	// dummy user で dummy job を登録
	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("failed to create job (status:%s), (body: %s)", resp.Status, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
		return nil, nil, err
	}
	return cl, job, nil
}

// agはCLユーザーでログインしている前提
func CreateJob(ctx context.Context, ag *agent.Agent) (*model.Job, error) {
	// dummy job 作成
	job := fixture.GenerateJob()

	// ログインユーザーで dummy job を登録
	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
		return nil, err
	}
	return job, nil
}

// agはCLユーザーでログインしている前提
func CreateJobByTitle(ctx context.Context, ag *agent.Agent, title string) (*model.Job, error) {
	// dummy job 作成
	job := fixture.GenerateJob()

	job.Title = title

	// ログインユーザーで dummy job を登録
	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
		return nil, err
	}
	return job, nil
}

// agはCLユーザーでログインしている前提
func CreateJobByDescription(ctx context.Context, ag *agent.Agent, description string) (*model.Job, error) {
	// dummy job 作成
	job := fixture.GenerateJob()

	job.Description = description

	// ログインユーザーで dummy job を登録
	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
		return nil, err
	}
	return job, nil
}

// agはCLユーザーでログインしている前提
func CreateJobBySalary(ctx context.Context, ag *agent.Agent, salary int) (*model.Job, error) {
	// dummy job 作成
	job := fixture.GenerateJob()

	job.Salary = salary

	// ログインユーザーで dummy job を登録
	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
		return nil, err
	}
	return job, nil
}

// agはCLユーザーでログインしている前提
func Create4JobsByTag(ctx context.Context, ag *agent.Agent, tag string) ([]*model.Job, error) {
	// dummy job 作成
	job1 := fixture.GenerateJob()
	job2 := fixture.GenerateJob()
	job3 := fixture.GenerateJob()
	job4 := fixture.GenerateJob()

	job1.Tags = tag
	job2.Tags = tag + "," + fixture.GenerateRandomTagsN(1)
	job3.Tags = fixture.GenerateRandomTagsN(1) + "," + tag
	tags := strings.Split(fixture.GenerateRandomTagsN(2), ",")
	job4.Tags = tags[0] + "," + tag + "," + tags[1]

	jobs := []*model.Job{job1, job2, job3, job4}

	// ログインユーザーで dummy job を登録
	for _, job := range jobs {
		err := func(job *model.Job) error {
			resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
				return err
			}
			return nil
		}(job)
		if err != nil {
			return nil, err
		}
	}
	return jobs, nil
}

func CreateArchivedJobWithNewUser(ctx context.Context, ag *agent.Agent) (*model.Job, error) {
	job := fixture.GenerateJob()
	_, err := CreateCLUserWithAuth(ctx, ag)
	if err != nil {
		return nil, err
	}

	resp, err := api.PostCLJob(ctx, ag, job.Title, job.Description, job.Salary, job.Tags)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(job); err != nil {
		return nil, err
	}

	resp, err = api.POSTCLJobArchive(ctx, ag, job.ID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return job, nil
}
