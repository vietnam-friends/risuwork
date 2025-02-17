package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/aws/aws-xray-sdk-go/awsplugins/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

var db *sql.DB

const (
	// MYSQL ERROR Code
	// ref: https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html
	MYSQL_ER_DUP_ENTRY     = 1062
	ER_NO_REFERENCED_ROW_2 = 1452
)

const (
	JOB_SEARCH_PAGE_SIZE       = 50
	APPLICATION_LIST_PAGE_SIZE = 20
	JOB_LIST_PAGE_SIZE         = 50
)

var (
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")
	dbUser = os.Getenv("DB_USER")
	dbPass = os.Getenv("DB_PASS")
)

func xraymw() echo.MiddlewareFunc {
	return echo.WrapMiddleware(func(h http.Handler) http.Handler {
		return xray.Handler(xray.NewFixedSegmentNamer("app"), h)
	})
}

func init() {
	ecs.Init()
	_ = xray.Configure(xray.Config{
		ServiceVersion: "1.0.0",
	})
}

func main() {
	// initialize db client
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "3306"
	}
	if dbName == "" {
		dbName = "risuwork"
	}
	if dbUser == "" {
		dbUser = "isucon"
	}
	if dbPass == "" {
		dbPass = "isucon"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	var err error
	db, err = xray.SQLContext("mysql", dsn)
	if err != nil {
		log.Fatal("Error connecting to the database:", err)
	}
	defer db.Close()

	// Echoのインスタンスを作成
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)

	// Middleware
	e.Use(xraymw())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyDump(func(e echo.Context, req, res []byte) {
		e.Logger().Debugj(log.JSON{"req": string(req), "res": string(res)})
	}))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	// Handler
	e.POST("/api/initialize", initializeHandler) // ベンチマーカー向けAPI
	e.POST("/api/finalize", finalizeHandler)     // ベンチマーカー向けAPI

	e.POST("/api/cs/signup", csSignupHandler)
	e.POST("/api/cs/login", csLoginHandler)
	e.POST("/api/cs/logout", csLogoutHandler)

	e.GET("/api/cs/job_search", searchJobHandler)
	e.POST("/api/cs/application", applyJobHandler)
	e.GET("/api/cs/applications", listApplicationHandler)

	e.POST("/api/cl/company", createCompanyHandler)
	e.POST("/api/cl/signup", clSignupHandler)
	e.POST("/api/cl/login", clLoginHandler)
	e.POST("/api/cl/logout", clLogoutHandler)

	e.POST("/api/cl/job", createJobHandler)
	e.PATCH("/api/cl/job/:jobid", updateJobHandler)
	e.POST("/api/cl/job/:jobid/archive", archiveJobHandler)
	e.GET("/api/cl/job/:jobid", getJobHandler)
	e.GET("/api/cl/jobs", listJobHandler)

	// サーバーを起動
	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func getSession(c echo.Context) (string, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		c.Logger().Error("Error read session", err)
		return "", err
	}
	if sess.Values["email"] == nil {
		c.Logger().Error("Error read session")
		return "", fmt.Errorf("error read session")
	}

	email := sess.Values["email"].(string)

	return email, nil
}

func setSession(c echo.Context, email string) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	sess.Values["email"] = email
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

func deleteSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	sess.Values["email"] = ""
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

// ベンチマーカー向けAPI
// POST /initialize
// ベンチマーカーが起動したときに最初に呼ぶ
// データベースの初期化などが実行されるため、スキーマを変更した場合などは適宜改変すること
func initializeHandler(c echo.Context) error {
	// MySQLデータベースを初期化
	out, err := exec.Command("../sql/init.sh").CombinedOutput()
	if err != nil {
		c.Logger().Errorf("Error exec init.sh: %v, output: %s", err, string(out))
		return c.JSON(http.StatusInternalServerError, "Error initializing database")
	}

	type InitializeResponse struct {
		Lang string `json:"lang"`
	}
	res := InitializeResponse{
		Lang: "go",
	}
	return c.JSON(http.StatusOK, res)
}

// ベンチマーカー向けAPI
// POST /finalize
// ベンチマーカーが終了するときに最後に呼ぶ
// デフォルトでは何も処理はしていないため、必要があれば実装を入れる
func finalizeHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok")
}

// CSアカウント作成API
// POST /cs/signup
func csSignupHandler(c echo.Context) error {
	// リクエストパラメータを取得
	type SignupRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	req := new(SignupRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger().Error("Error generating bcrypt hash:", err)
		return c.JSON(http.StatusInternalServerError, "Error generating bcrypt hash")
	}

	// アカウントを作成
	res, err := db.ExecContext(c.Request().Context(), "INSERT INTO user (email, password, name, user_type) VALUES (?, ?, ?, ?)", req.Email, hashedPassword, req.Name, "CS")
	if err != nil {
		// 登録済みの場合は409を返す
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == MYSQL_ER_DUP_ENTRY {
			return c.JSON(http.StatusConflict, "Email address is already used")
		}

		c.Logger().Error("Error creating CS account:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating account")
	}

	// セッションを作成
	err = setSession(c, req.Email)
	if err != nil {
		c.Logger().Error("Error setting session:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating account")
	}

	userID, err := res.LastInsertId()
	if err != nil {
		c.Logger().Error("Error getting user ID:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating account")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "CS account created successfully", "id": userID})
}

// CSログインAPI
// POST /cs/login
func csLoginHandler(c echo.Context) error {
	// リクエストパラメータを取得
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// パスワードをDBから取得
	var storedPassword string
	err := db.QueryRowContext(c.Request().Context(), "SELECT password FROM user WHERE email = ?", req.Email).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, "Invalid email or password")
		}
		c.Logger().Error("Error querying database:", err)
		return c.JSON(http.StatusInternalServerError, "Error logging in")
	}

	// パスワードを比較
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid email or password")
	}

	// セッションを作成
	err = setSession(c, req.Email)
	if err != nil {
		c.Logger().Error("Error setting session:", err)
		return c.JSON(http.StatusInternalServerError, "Error logging in")
	}

	return c.JSON(http.StatusOK, "Logged in successfully")
}

// CSログアウトAPI
// POST /cs/logout
func csLogoutHandler(c echo.Context) error {
	// ログイン認証
	_, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// ログアウト処理
	err = deleteSession(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error logging out")
	}

	return c.JSON(http.StatusOK, "Logged out successfully")
}

// CS求人検索API
// GET /cs/job_search
func searchJobHandler(c echo.Context) error {
	// リクエストパラメータを取得
	type JobSearchRequest struct {
		Keyword    string `query:"keyword"`
		MinSalary  int    `query:"min_salary"` // grater than or equal
		MaxSalary  int    `query:"max_salary"` // less than
		Tag        string `query:"tag"`
		IndustryID string `query:"industry_id"`
		Page       int    `query:"page"` // 0-indexed
	}
	req := JobSearchRequest{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// SQLクエリの基本部分を作成
	query := "SELECT id, title, description, salary, tags, created_at, updated_at FROM job WHERE is_active = true AND is_archived = false"
	params := []interface{}{}

	// フリーワード検索
	if req.Keyword != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword)
	}

	// 給与範囲検索
	if req.MinSalary > 0 {
		query += " AND salary >= ?"
		params = append(params, req.MinSalary)
	}
	if req.MaxSalary > 0 {
		query += " AND salary <= ?"
		params = append(params, req.MaxSalary)
	}

	// タグ検索
	// タグはカンマ区切りで格納されているため
	// - jobにタグが複数ある場合
	//   - 検索対象のタグが最初にある場合
	//   - 途中にある場合
	//   - 最後にある場合
	// - jobにタグが1つしかない場合
	//   - 検索対象のタグと完全一致
	// という4パターンを考慮する必要がある
	if req.Tag != "" {
		query += " AND (tags LIKE ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?)"
		params = append(params, req.Tag+",%", "%,"+req.Tag+",%", "%,"+req.Tag, req.Tag)
	}

	// ソート順指定
	query = query + " ORDER BY updated_at DESC, id desc"

	// クエリを実行
	rows, err := db.QueryContext(c.Request().Context(), query, params...)
	if err != nil {
		c.Logger().Error("Error searching jobs:", err)
		return c.JSON(http.StatusInternalServerError, "Error searching jobs")
	}
	defer rows.Close()

	type Job struct {
		ID             int       `json:"id"`
		JobTitle       string    `json:"title"`
		JobDescription string    `json:"description"`
		Salary         float64   `json:"salary"`
		Tags           string    `json:"tags"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	var jobs []Job
	for rows.Next() {
		var job Job
		err := rows.Scan(&job.ID, &job.JobTitle, &job.JobDescription, &job.Salary, &job.Tags, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			c.Logger().Error("Error scanning row:", err)
			return c.JSON(http.StatusInternalServerError, "Error searching jobs")
		}
		jobs = append(jobs, job)
	}

	type Company struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Industry string `json:"industry"`
	}

	type JobWithCompany struct {
		Job
		Company Company `json:"company"`
	}

	// 求人ごとの企業情報を取得
	jobsWithCompany := []JobWithCompany{}
	for _, job := range jobs {
		var company Company
		var industryID string
		err := db.QueryRowContext(c.Request().Context(), "SELECT company.id, company.name, industry_category.name as industry, company.industry_id FROM company JOIN industry_category ON company.industry_id = industry_category.id WHERE company.id = (SELECT company_id FROM user WHERE id = (SELECT create_user_id FROM job WHERE id = ?))", job.ID).Scan(&company.ID, &company.Name, &company.Industry, &industryID)
		if err != nil {
			c.Logger().Error("Error fetch company from db:", err)
			return c.JSON(http.StatusInternalServerError, "Error searching jobs")
		}

		// 検索時に指定した業種と違う場合はスキップ
		if req.IndustryID != "" && req.IndustryID != industryID {
			continue
		}

		jobsWithCompany = append(jobsWithCompany, JobWithCompany{
			job,
			company,
		})
	}

	type JobSearchResponse struct {
		Jobs        []JobWithCompany `json:"jobs"`
		Page        int              `json:"page"`
		HasNextPage bool             `json:"has_next_page"`
	}
	resp := JobSearchResponse{}

	for i, job := range jobsWithCompany {
		if i < (req.Page)*JOB_SEARCH_PAGE_SIZE {
			continue
		}

		if len(resp.Jobs) >= JOB_SEARCH_PAGE_SIZE {
			resp.HasNextPage = true
			break
		}
		resp.Jobs = append(resp.Jobs, job)
	}

	return c.JSON(http.StatusOK, resp)
}

// CS求人への応募API
// POST /cs/application
func applyJobHandler(c echo.Context) error {
	// ログイン認証
	email, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// リクエストパラメータを取得
	type ApplicationRequest struct {
		JobID int `json:"job_id"`
	}
	req := new(ApplicationRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// トランザクションを開始
	tx, err := db.BeginTx(c.Request().Context(), nil)
	if err != nil {
		c.Logger().Error("Error starting transaction:", err)
		return c.JSON(http.StatusInternalServerError, "Error applying for job")
	}
	defer tx.Rollback()

	// ユーザー情報をDBから取得
	type User struct {
		ID       int    `json:"id"`
		UserType string `json:"user_type"`
	}
	var user User
	err = tx.QueryRowContext(c.Request().Context(), "SELECT id, user_type FROM user WHERE email = ?", email).Scan(&user.ID, &user.UserType)
	if err != nil {
		c.Logger().Error("Error querying database:", err)
		return c.JSON(http.StatusInternalServerError, "Error applying for job")
	}

	// CSユーザーでなければ403を返す
	if user.UserType != "CS" {
		return c.JSON(http.StatusForbidden, "Forbidden")
	}

	// 求人が応募可能か確認すると同時にロックを取得
	var canApply bool
	err = tx.QueryRowContext(c.Request().Context(), "SELECT is_active = true AND is_archived = false FROM job WHERE id = ? FOR UPDATE", req.JobID).Scan(&canApply)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, "Job not found")
		}
		c.Logger().Error("Error fetch job from database:", err)
		return c.JSON(http.StatusInternalServerError, "Error applying for job")
	}
	if !canApply {
		return c.JSON(http.StatusUnprocessableEntity, "Job is not accepting applications")
	}

	// 応募済みかどうか確認
	var exists bool
	err = tx.QueryRowContext(c.Request().Context(), "SELECT EXISTS (SELECT 1 FROM application WHERE job_id = ? AND user_id = ?)", req.JobID, user.ID).Scan(&exists)
	if err != nil {
		c.Logger().Error("Error fetch application from database:", err)
		return c.JSON(http.StatusInternalServerError, "Error applying for job")
	}

	if exists {
		return c.JSON(http.StatusConflict, "Already applied for the job")
	}

	// データベースに応募情報を挿入
	res, err := tx.ExecContext(c.Request().Context(), "INSERT INTO application (job_id, user_id) VALUES (?, ?)", req.JobID, user.ID)
	if err != nil {
		c.Logger().Error("Error applying for job:", err)
		return c.JSON(http.StatusInternalServerError, "Error applying for job")
	}

	// トランザクションをコミット
	err = tx.Commit()
	if err != nil {
		c.Logger().Error("Error commiting transaction:", err)
		return c.JSON(http.StatusInternalServerError, "Error applying for job")
	}

	applicationID, err := res.LastInsertId()
	if err != nil {
		c.Logger().Error("Error getting application ID:", err)
		return c.JSON(http.StatusInternalServerError, "Error applying for job")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "Successfully applied for the job", "id": applicationID})
}

// CS応募一覧取得API
// GET /cs/applications
func listApplicationHandler(c echo.Context) error {
	// ログイン認証
	email, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// リクエストパラメータを取得
	type ApplicationListRequest struct {
		Page int `query:"page"` // 0-indexed
	}
	req := new(ApplicationListRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// 応募一覧を取得
	rows, err := db.QueryContext(c.Request().Context(), "SELECT a.id, a.job_id, a.user_id, a.created_at FROM application a JOIN user u ON a.user_id = u.id WHERE u.email = ? ORDER BY a.created_at DESC", email)
	if err != nil {
		c.Logger().Error("Error querying database:", err)
		return c.JSON(http.StatusInternalServerError, "Error getting applications")
	}
	defer rows.Close()

	type Job struct {
		ID             int       `json:"id"`
		JobTitle       string    `json:"title"`
		JobDescription string    `json:"description"`
		Salary         float64   `json:"salary"`
		Tags           string    `json:"tags"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	type Application struct {
		ID        int       `json:"id"`
		JobID     int       `json:"job_id"`
		UserID    int       `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		Job       Job       `json:"job"`
	}

	var applications []Application
	for rows.Next() {
		var application Application
		err := rows.Scan(&application.ID, &application.JobID, &application.UserID, &application.CreatedAt)
		if err != nil {
			c.Logger().Error("Error scanning row:", err)
			continue
		}
		applications = append(applications, application)
	}

	type ApplicationsResponse struct {
		Applications []Application `json:"applications"`
		Page         int           `json:"page"`
		HasNextPage  bool          `json:"has_next_page"`
	}

	// 求人情報を取得
	for i, application := range applications {

		var job Job
		err := db.QueryRowContext(c.Request().Context(), "SELECT id, title, description, salary, tags, created_at, updated_at FROM job WHERE id = ?", application.JobID).Scan(&job.ID, &job.JobTitle, &job.JobDescription, &job.Salary, &job.Tags, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			c.Logger().Error("Error querying database:", err)
			return c.JSON(http.StatusInternalServerError, "Error getting applications")
		}
		applications[i].Job = job
	}

	resp := ApplicationsResponse{}
	for i, application := range applications {
		if i < (req.Page)*APPLICATION_LIST_PAGE_SIZE {
			continue
		}

		if len(resp.Applications) >= APPLICATION_LIST_PAGE_SIZE {
			resp.HasNextPage = true
			break
		}
		resp.Applications = append(resp.Applications, application)
	}

	return c.JSON(http.StatusOK, resp)
}

// CL企業登録API
// POST /cl/company
func createCompanyHandler(c echo.Context) error {
	// リクエストパラメータを取得
	type CompanyRequest struct {
		Name       string `json:"name"`
		IndustryID string `json:"industry_id"`
	}
	req := new(CompanyRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// 企業をデータベースに登録
	result, err := db.ExecContext(c.Request().Context(), "INSERT INTO company (name, industry_id) VALUES (?, ?)", req.Name, req.IndustryID)
	if err != nil {
		c.Logger().Error("Error creating company:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating company")
	}

	companyID, err := result.LastInsertId()
	if err != nil {
		c.Logger().Error("Error getting company ID:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating company")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"message": "Company created successfully", "id": companyID})
}

// CLアカウント作成API
// POST /cl/signup
func clSignupHandler(c echo.Context) error {
	// リクエストパラメータを取得
	type SignupRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Name      string `json:"name"`
		CompanyID int    `json:"company_id"`
	}
	req := new(SignupRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger().Error("Error hashing password:", err)
		return c.JSON(http.StatusInternalServerError, "Error signing up")
	}

	// ユーザーをデータベースに登録
	res, err := db.ExecContext(c.Request().Context(), "INSERT INTO user (email, password, name, user_type, company_id) VALUES (?, ?, ?, 'CL', ?)", req.Email, hashedPassword, req.Name, req.CompanyID)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			switch mysqlErr.Number {
			case MYSQL_ER_DUP_ENTRY:
				// 登録済みの場合は409を返す
				return c.JSON(http.StatusConflict, "Email address is already used")
			case ER_NO_REFERENCED_ROW_2:
				// 存在しない企業IDの場合は400を返す
				return c.JSON(http.StatusBadRequest, "Company not found")
			}
		}

		c.Logger().Error("Error creating user:", err)
		return c.JSON(http.StatusInternalServerError, "Error signing up")
	}

	// セッションを作成
	err = setSession(c, req.Email)
	if err != nil {
		c.Logger().Error("Error setting session:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating account")
	}

	userID, err := res.LastInsertId()
	if err != nil {
		c.Logger().Error("Error getting user ID:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating account")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "Signed up successfully", "id": userID})
}

// CLログインAPI
// POST /cl/login
func clLoginHandler(c echo.Context) error {
	// リクエストパラメータを取得
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// パスワードをDBから取得
	var storedPassword string
	err := db.QueryRowContext(c.Request().Context(), "SELECT password FROM user WHERE email = ? AND user_type = 'CL'", req.Email).Scan(&storedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, "Invalid email or password")
		}
		c.Logger().Error("Error querying database:", err)
		return c.JSON(http.StatusInternalServerError, "Error logging in")
	}

	// パスワードを比較
	if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid email or password")
	}

	// セッションを作成
	err = setSession(c, req.Email)
	if err != nil {
		c.Logger().Error("Error setting session:", err)
		return c.JSON(http.StatusInternalServerError, "Error logging in")
	}

	return c.JSON(http.StatusOK, "Logged in successfully")
}

// CLログアウトAPI
// POST /cl/logout
func clLogoutHandler(c echo.Context) error {
	// ログイン認証
	_, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// ログアウト処理
	err = deleteSession(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error logging out")
	}

	return c.JSON(http.StatusOK, "Logged out successfully")
}

// CL求人作成API
// POST /cl/job
func createJobHandler(c echo.Context) error {
	// ログイン認証
	email, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// ユーザーをDBから取得
	type User struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		Name      string `json:"name"`
		UserType  string `json:"user_type"`
		CompanyID int    `json:"company_id"`
	}
	var user User
	err = db.QueryRowContext(c.Request().Context(), "SELECT id, email, password, name, user_type, company_id FROM user WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.Password, &user.Name, &user.UserType, &user.CompanyID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Logger().Error("Session user not found", err)
		} else {
			c.Logger().Error("Error fetch user from db", err)
		}
		return c.JSON(http.StatusInternalServerError, "Error creating job")
	}

	// 企業アカウントでなければ403を返す
	if user.UserType != "CL" {
		return c.JSON(http.StatusForbidden, "No permission")
	}

	// リクエストパラメータを取得
	type JobRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Salary      int    `json:"salary"`
		Tags        string `json:"tags"`
	}
	req := new(JobRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// 求人をデータベースに登録
	result, err := db.ExecContext(c.Request().Context(), "INSERT INTO job (title, description, salary, tags, is_active, create_user_id) VALUES (?, ?, ?, ?, true, ?)", req.Title, req.Description, req.Salary, req.Tags, user.ID)
	if err != nil {
		c.Logger().Error("Error creating job:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating job")
	}

	jobID, err := result.LastInsertId()
	if err != nil {
		c.Logger().Error("Error getting job ID:", err)
		return c.JSON(http.StatusInternalServerError, "Error creating job")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"message": "Job created successfully", "id": jobID})
}

// ログインユーザーが求人を閲覧・編集できるかどうかチェックするための関数
func canAccessJob(c echo.Context, jobID string, email string, includeArchived bool) (bool, error) {
	// ログインユーザーを取得
	type CLUser struct {
		UserType  string `json:"user_type"`
		CompanyID int    `json:"company_id"`
	}
	var user CLUser
	err := db.QueryRowContext(c.Request().Context(), "SELECT user_type, company_id FROM user WHERE email = ?", email).Scan(&user.UserType, &user.CompanyID)
	if err != nil {
		c.Logger().Error("Error fetch user from db:", err)
		return false, c.JSON(http.StatusInternalServerError, "Error updating job")
	}

	// 企業アカウントでなければ403を返す
	if user.UserType != "CL" {
		return false, c.JSON(http.StatusForbidden, "No permission")
	}

	// 求人を取得して存在するかチェック
	type Job struct {
		CreateUserID int  `json:"create_user_id"`
		IsArchived   bool `json:"is_archived"`
	}
	var job Job
	err = db.QueryRowContext(c.Request().Context(), "SELECT create_user_id, is_archived FROM job WHERE id = ?", jobID).Scan(&job.CreateUserID, &job.IsArchived)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, c.JSON(http.StatusNotFound, "Job not found")
		}
		c.Logger().Error("Error fetch job from database:", err)
		return false, c.JSON(http.StatusInternalServerError, "Error updating job")
	}
	if !includeArchived && job.IsArchived {
		return false, c.JSON(http.StatusUnprocessableEntity, "Job archived")
	}

	// 求人作成ユーザーを取得
	var jobCreateUser CLUser
	err = db.QueryRowContext(c.Request().Context(), "SELECT company_id FROM user WHERE id = ?", job.CreateUserID).Scan(&jobCreateUser.CompanyID)
	if err != nil {
		c.Logger().Error("Error fetch job create user from database:", err)
		return false, c.JSON(http.StatusInternalServerError, "Error getting job")
	}

	// 求人作成ユーザーとログインユーザーの所属会社が異なる場合は403を返す
	if jobCreateUser.CompanyID != user.CompanyID {
		return false, c.JSON(http.StatusForbidden, "No permission")
	}
	return true, nil
}

// CL求人更新API
// PATCH /cl/job/:jobid
func updateJobHandler(c echo.Context) error {
	// ログイン認証
	email, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// リクエストパラメータを取得
	type UpdateJobRequest struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Salary      *int    `json:"salary"`
		Tags        *string `json:"tags"`
		IsActive    *bool   `json:"is_active"`
	}
	req := new(UpdateJobRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}
	jobID := c.Param("jobid")

	// 編集できるかどうかチェック
	ok, err := canAccessJob(c, jobID, email, false)
	if !ok {
		return err
	}

	// 求人情報を更新
	query := "UPDATE job SET"
	params := []interface{}{}
	if req.Title != nil {
		query += " title = ?,"
		params = append(params, *req.Title)
	}
	if req.Description != nil {
		query += " description = ?,"
		params = append(params, *req.Description)
	}
	if req.Salary != nil {
		query += " salary = ?,"
		params = append(params, *req.Salary)
	}
	if req.Tags != nil {
		query += " tags = ?,"
		params = append(params, *req.Tags)
	}
	if req.IsActive != nil {
		query += " is_active = ?,"
		params = append(params, *req.IsActive)
	}
	query = query[:len(query)-1] + " WHERE id = ?"
	params = append(params, jobID)
	_, err = db.ExecContext(c.Request().Context(), query, params...)
	if err != nil {
		c.Logger().Error("Error updating job:", err)
		return c.JSON(http.StatusInternalServerError, "Error updating job")
	}

	return c.JSON(http.StatusOK, "Job updated successfully")
}

// CL求人アーカイブAPI
// POST /cl/job/:jobid/archive
// 求人をアーカイブすると
// GET /cs/job_search, GET /cl/jobs で取得できなくなる
// GET /cs/applications, GET /cl/job/:jobid では引き続き取得可能
func archiveJobHandler(c echo.Context) error {
	// ログイン認証
	email, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// リクエストパラメータを取得
	jobID := c.Param("jobid")

	// アーカイブできるかどうかチェック
	ok, err := canAccessJob(c, jobID, email, false)
	if !ok {
		return err
	}

	// 求人をアーカイブ
	_, err = db.ExecContext(c.Request().Context(), "UPDATE job SET is_archived = true WHERE id = ?", jobID)
	if err != nil {
		c.Logger().Error("Error archiving job:", err)
		return c.JSON(http.StatusInternalServerError, "Error archiving job")
	}

	return c.JSON(http.StatusOK, "Job archived successfully")
}

// CL求人取得API
// GET /cl/job/:jobid
func getJobHandler(c echo.Context) error {
	// ログイン認証
	email, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// リクエストパラメータを取得
	jobID := c.Param("jobid")

	// 閲覧できるかどうかチェック
	ok, err := canAccessJob(c, jobID, email, true)
	if !ok {
		return err
	}

	type CSUser struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	type Application struct {
		ID        int       `json:"id"`
		JobID     int       `json:"job_id"`
		UserID    int       `json:"-"`
		CreatedAt time.Time `json:"created_at"`
		Applicant CSUser    `json:"applicant"`
	}

	type Job struct {
		ID             int           `json:"id"`
		JobTitle       string        `json:"title"`
		JobDescription string        `json:"description"`
		Salary         int           `json:"salary"`
		Tags           string        `json:"tags"`
		IsActive       bool          `json:"is_active"`
		CreateUserID   int           `json:"create_user_id"`
		CreatedAt      time.Time     `json:"created_at"`
		UpdatedAt      time.Time     `json:"updated_at"`
		Applications   []Application `json:"applications"`
	}

	// 求人を取得
	var job Job
	err = db.QueryRowContext(c.Request().Context(), "SELECT id, title, description, salary, tags, is_active, create_user_id, created_at, updated_at FROM job WHERE id = ?", jobID).Scan(&job.ID, &job.JobTitle, &job.JobDescription, &job.Salary, &job.Tags, &job.IsActive, &job.CreateUserID, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		c.Logger().Error("Error querying database:", err)
		return c.JSON(http.StatusInternalServerError, "Error getting job")
	}

	// 求人への応募を取得
	var applications []Application
	rows, err := db.QueryContext(c.Request().Context(), "SELECT id, job_id, user_id, created_at FROM application WHERE job_id = ? ORDER BY created_at", jobID)
	if err != nil {
		c.Logger().Error("Error querying database:", err)
		return c.JSON(http.StatusInternalServerError, "Error getting job")
	}
	defer rows.Close()

	for rows.Next() {
		var application Application
		err := rows.Scan(&application.ID, &application.JobID, &application.UserID, &application.CreatedAt)
		if err != nil {
			c.Logger().Error("Error scanning row:", err)
			continue
		}
		applications = append(applications, application)
	}

	// 応募者の情報を取得
	for i, application := range applications {
		var user CSUser
		err := db.QueryRowContext(c.Request().Context(), "SELECT id, email, name FROM user WHERE id = ?", application.UserID).Scan(&user.ID, &user.Email, &user.Name)
		if err != nil {
			c.Logger().Error("Error querying database:", err)
			return c.JSON(http.StatusInternalServerError, "Error getting job")
		}
		applications[i].Applicant = user
	}

	job.Applications = applications

	return c.JSON(http.StatusOK, job)
}

// CL求人一覧取得API
// GET /cl/jobs
func listJobHandler(c echo.Context) error {
	// ログイン認証
	email, err := getSession(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Not logged in")
	}

	// ログインユーザーを取得
	type CLUser struct {
		ID        int    `json:"id"`
		UserType  string `json:"user_type"`
		CompanyID int    `json:"company_id"`
	}
	var user CLUser
	err = db.QueryRowContext(c.Request().Context(), "SELECT id, user_type, company_id FROM user WHERE email = ?", email).Scan(&user.ID, &user.UserType, &user.CompanyID)
	if err != nil {
		c.Logger().Error("Error fetch user from db:", err)
		return c.JSON(http.StatusInternalServerError, "Error getting jobs")
	}

	// 企業アカウントでなければ403を返す
	if user.UserType != "CL" {
		return c.JSON(http.StatusForbidden, "No permission")
	}

	// リクエストパラメータを取得
	type JobListRequest struct {
		Page int `query:"page"` // 0-indexed
	}
	req := new(JobListRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request payload")
	}

	// 求人一覧を取得
	type Job struct {
		ID             int       `json:"id"`
		JobTitle       string    `json:"title"`
		JobDescription string    `json:"description"`
		Salary         int       `json:"salary"`
		Tags           string    `json:"tags"`
		IsActive       bool      `json:"is_active"`
		CreateUserID   int       `json:"create_user_id"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	rows, err := db.QueryContext(c.Request().Context(), "SELECT id, title, description, salary, tags, is_active, create_user_id, created_at, updated_at FROM job WHERE is_archived = false AND create_user_id IN (SELECT id FROM user WHERE company_id = ?) ORDER BY updated_at DESC, id", user.CompanyID)
	if err != nil {
		c.Logger().Error("Error querying database:", err)
		return c.JSON(http.StatusInternalServerError, "Error getting jobs")
	}
	defer rows.Close()

	type JobListResponse struct {
		Jobs        []Job `json:"jobs"`
		Page        int   `json:"page"`
		HasNextPage bool  `json:"has_next_page"`
	}
	resp := JobListResponse{}

	i := 0
	for rows.Next() {
		if i < (req.Page)*JOB_LIST_PAGE_SIZE {
			i++
			continue
		}

		if len(resp.Jobs) >= JOB_LIST_PAGE_SIZE {
			resp.HasNextPage = true
			break
		}

		var job Job
		err := rows.Scan(&job.ID, &job.JobTitle, &job.JobDescription, &job.Salary, &job.Tags, &job.IsActive, &job.CreateUserID, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			c.Logger().Error("Error scanning row:", err)
			return c.JSON(http.StatusInternalServerError, "Error getting jobs")
		}
		resp.Jobs = append(resp.Jobs, job)
		i++
	}
	return c.JSON(http.StatusOK, resp)
}
