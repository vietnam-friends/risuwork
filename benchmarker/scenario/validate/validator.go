package validate

import (
	"errors"
	"fmt"
	"net/http"
	"risuwork-benchmarker/scenario/model"
	"time"
)

type StatusCodeUnMatchError struct {
	Method   string
	URL      string
	Expected int
	Actual   int
}

func (e StatusCodeUnMatchError) Error() string {
	return fmt.Sprintf("%s %s ステータスコードが正しくありません: expected %d but got %d", e.Method, e.URL, e.Expected, e.Actual)
}

func StatusCode(r *http.Response, statusCode int) error {
	if r.StatusCode != statusCode {
		// ステータスコードが一致しなければ HTTP メソッド、URL パス、期待したステータスコード、実際のステータスコードを持つ
		// エラーを返す
		return StatusCodeUnMatchError{
			Method:   r.Request.Method,
			URL:      r.Request.URL.Path,
			Expected: statusCode,
			Actual:   r.StatusCode,
		}
	}
	return nil
}

func JobForCS(actual, expected model.Job) error {
	// CSユーザーにはIsActiveを返さないため、バリデーションしない
	if actual.ID != expected.ID {
		return fmt.Errorf("job id: expected(%d) != actual(%d)", expected.ID, actual.ID)
	}
	if actual.Title != expected.Title {
		return fmt.Errorf("job title: expected(%s) != actual(%s)", expected.Title, actual.Title)
	}
	if actual.Description != expected.Description {
		return fmt.Errorf("job description: expected(%s) != actual(%s)", expected.Description, actual.Description)
	}
	if actual.Salary != expected.Salary {
		return fmt.Errorf("job salary: expected(%d) != actual(%d)", expected.Salary, actual.Salary)
	}
	if actual.Tags != expected.Tags {
		return fmt.Errorf("job tags: expected(%s) != actual(%s)", expected.Tags, actual.Tags)
	}
	if actual.Company.ID != expected.Company.ID {
		return fmt.Errorf("job company id: expected(%d) != actual(%d)", expected.Company.ID, actual.Company.ID)
	}
	if actual.Company.Name != expected.Company.Name {
		return fmt.Errorf("job company name: expected(%s) != actual(%s)", expected.Company.Name, actual.Company.Name)
	}
	if actual.Company.Industry != expected.Company.Industry {
		return fmt.Errorf("job company industry: expected(%s) != actual(%s)", expected.Company.Industry, actual.Company.Industry)
	}

	return nil
}

func JobForCL(actual, expected model.Job) error {
	// CSユーザーにはIsActiveを返さないため、バリデーションしない
	if actual.ID != expected.ID {
		return fmt.Errorf("job id: expected(%d) != actual(%d)", expected.ID, actual.ID)
	}
	if actual.Title != expected.Title {
		return fmt.Errorf("job title: expected(%s) != actual(%s)", expected.Title, actual.Title)
	}
	if actual.Description != expected.Description {
		return fmt.Errorf("job description: expected(%s) != actual(%s)", expected.Description, actual.Description)
	}
	if actual.Salary != expected.Salary {
		return fmt.Errorf("job salary: expected(%d) != actual(%d)", expected.Salary, actual.Salary)
	}
	if actual.Tags != expected.Tags {
		return fmt.Errorf("job tags: expected(%s) != actual(%s)", expected.Tags, actual.Tags)
	}
	if actual.IsActive != expected.IsActive {
		return fmt.Errorf("job is_active: expected(%t) != actual(%t)", expected.IsActive, actual.IsActive)
	}
	return nil
}

// jobResultのソート順はupdated_atの降順
func JobResultOrder(applications []model.Job) error {
	if len(applications) < 1 {
		return nil
	}
	for i := 0; i < len(applications)-1; i++ {
		if applications[i].UpdatedAt.Before(applications[i+1].UpdatedAt) {
			return fmt.Errorf("%dth job's updated_at: %s is before %dth job's updated_at: %s", i, applications[i].UpdatedAt.Format(time.StampMicro), i+1, applications[i+1].UpdatedAt.Format(time.StampMicro))
		}
	}
	return nil
}

// jobResultのソート順はupdated_atの降順
func ApplicationResultOrder(applications []model.Application) error {
	if len(applications) < 1 {
		return nil
	}
	for i := 0; i < len(applications)-1; i++ {
		if applications[i].CreatedAt.Before(applications[i+1].CreatedAt) {
			return errors.New("求人の並び順が正しくありません")
		}
	}
	return nil
}
