import { exec } from "child_process";
import express, { Request, Response } from "express";
import mysqlPromise, {
  PoolOptions,
  QueryError,
  ResultSetHeader,
  RowDataPacket,
} from "mysql2/promise";
import bcrypt from "bcryptjs";
import session from "cookie-session";
import awsXRay from "aws-xray-sdk";

const JOB_SEARCH_PAGE_SIZE = 50;
const APPLICATION_LIST_PAGE_SIZE = 20;
const JOB_LIST_PAGE_SIZE = 50;

const dbConfig = {
  host: process.env["DB_HOST"] ?? "localhost",
  port: Number(process.env["DB_PORT"] ?? 3306),
  database: process.env["DB_NAME"] ?? "risuwork",
  user: process.env["DB_USER"] ?? "isucon",
  password: process.env["DB_PASS"] ?? "isucon",
} satisfies PoolOptions;

const pool = mysqlPromise.createPool(dbConfig);

awsXRay.captureMySQL(require("mysql2"));

const app = express();

const XRayExpress = awsXRay.express;
app.use(XRayExpress.openSegment("app"));

app.use(express.json());

const sessionCookieName = "risuwork";
app.use(
  session({
    name: sessionCookieName,
    secret: "secret",
    path: "/",
    maxAge: 3600000,
    httpOnly: true,
    sameSite: "strict",
  })
);

app.use((req, res, next) => {
  console.log("Request URL:", req.url);
  console.log("Request Query:", req.query);
  console.log("Request Body:", req.body);

  const originalWrite = res.write;
  const originalEnd = res.end;
  const chunks: any[] = [];

  res.write = function (chunk, ...args: any[]) {
    chunks.push(chunk);
    // @ts-expect-error
    return originalWrite.apply(res, [chunk, ...args]);
  };

  res.end = function (chunk, ...args: any[]) {
    if (chunk) {
      chunks.push(chunk);
    }
    const body = Buffer.concat(chunks).toString("utf8");
    console.log("Response Body:", body);
    // @ts-expect-error
    return originalEnd.apply(res, [chunk, ...args]);
  };

  next();
});

app.use((req, res, next) => {
  res.setHeader("Cache-Control", "no-store");
  next();
});

const getSession = (req: Request) => {
  if (!req.session || !req.session.email) {
    console.error("Error read session");
    return null;
  }

  return req.session.email;
};

const setSession = (req: Request) => {
  if (!req.session) {
    console.error("Session not found");
    return false;
  }

  const { email } = req.body;
  req.session.email = email;

  return true;
};

const deleteSession = (req: Request) => {
  if (!req.session) {
    console.error("Session not found");
    return false;
  }

  req.session.email = "";
  return true;
};

// ベンチマーカー向けAPI
// ベンチマーカーが起動したときに最初に呼ぶ
// データベースの初期化などが実行されるため、スキーマを変更した場合などは適宜改変すること
app.post("/api/initialize", (req: Request, res: Response) => {
  exec("../sql/init.sh", (err, stdout, stderr) => {
    if (err) {
      console.error(`Error exec init.sh: ${err}, output: ${stderr}`);
      return res.status(500).json({ error: "Error initializing database" });
    }

    return res.status(200).json({
      lang: "nodejs",
    });
  });
});

// ベンチマーカー向けAPI
// ベンチマーカーが終了するときに最後に呼ぶ
// デフォルトでは何も処理はしていないため、必要があれば実装を入れる
app.post("/api/finalize", (req: Request, res: Response) => {
  return res.status(200).json("ok");
});

// CSアカウント作成API
app.post("/api/cs/signup", async (req: Request, res: Response) => {
  // リクエストパラメータを取得
  type CsSignupRequest = {
    email: string;
    password: string;
    name: string;
  };
  const { email, password, name } = req.body as CsSignupRequest;
  if (!name || !password || !email) {
    return res.status(400).json({ message: "Invalid request payload" });
  }

  // パスワードをハッシュ化
  let hashedPassword: string;
  try {
    hashedPassword = bcrypt.hashSync(password, 10);
  } catch (err) {
    console.error("Error generating bcrypt hash:", err);
    return res.status(500).json({ message: "Error generating bcrypt hash" });
  }

  let userId: number;
  // アカウントを作成
  try {
    const [result] = await pool.query<ResultSetHeader>(
      "INSERT INTO user (email, password, name, user_type) VALUES (?, ?, ?, ?)",
      [email, hashedPassword, name, "CS"]
    );

    userId = result.insertId;
  } catch (err) {
    if (err instanceof Error && (err as QueryError).code === "ER_DUP_ENTRY") {
      return res.status(409).json({ message: "Email address is already used" });
    }

    console.error("Error creating CS account:", err);
    return res.status(500).json({ message: "Error creating account" });
  }

  // セッションを作成
  const ok = setSession(req);
  if (!ok) {
    return res.status(500).json({ message: "Error creating account" });
  }

  return res
    .status(200)
    .json({ message: "CS account created successfully", id: userId });
});

// CSログインAPI
app.post("/api/cs/login", async (req: Request, res: Response) => {
  // リクエストパラメータを取得
  type CsLoginRequest = {
    email: string;
    password: string;
  };
  const { email, password } = req.body as CsLoginRequest;
  if (!password || !email) {
    return res.status(400).json({ message: "Invalid request payload" });
  }

  let storedPassword: string;
  try {
    type User = {
      password: string;
    };
    type UserRow = RowDataPacket & User;
    // パスワードをDBから取得
    const [rows] = await pool.query<UserRow[]>(
      "SELECT password FROM user WHERE email = ?",
      [email]
    );

    if (!rows.length) {
      return res.status(401).json({ message: "Invalid email or password" });
    }

    storedPassword = rows[0].password;
  } catch (err) {
    console.error("Error querying database:", err);
    return res.status(500).json({ message: "Error logging in" });
  }

  // パスワードを比較
  const isMatch = bcrypt.compareSync(password, storedPassword);
  if (!isMatch) {
    return res.status(401).json({ message: "Invalid email or password" });
  }

  // セッションを作成
  const ok = setSession(req);
  if (!ok) {
    return res.status(500).json({ message: "Error creating account" });
  }

  return res.status(200).json({ message: "Logged in successfully" });
});

// CSログアウトAPI
app.post("/api/cs/logout", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  // ログアウト処理
  const ok = deleteSession(req);
  if (!ok) {
    return res.status(500).json({ message: "Error logging out" });
  }

  return res.status(200).json("Logged out successfully");
});

// CS求人検索API
app.get("/api/cs/job_search", async (req: Request, res: Response) => {
  // リクエストパラメータを取得
  type CsJobSearchRequest = {
    keyword: string | null;
    min_salary: string | null | number; // grater than or equal
    max_salary: string | null | number; // less than
    tag: string | null;
    industry_id: string | null | number;
    page: string | null | number; // 0-indexed
  };
  let { keyword, min_salary, max_salary, tag, industry_id, page } =
    req.query as CsJobSearchRequest;
  min_salary = min_salary ? Number(min_salary) : 0;
  max_salary = max_salary ? Number(max_salary) : 0;
  page = page ? Number(page) : 0;

  type Job = {
    id: number;
    title: string;
    description: string;
    salary: number;
    tags: string;
    created_at: Date;
    updated_at: Date;
  };
  let jobs: Job[] = [];
  try {
    // SQLクエリの基本部分を作成
    let query =
      "SELECT id, title, description, salary, tags, created_at, updated_at FROM job WHERE is_active = true AND is_archived = false";
    const params = [];

    // フリーワード検索
    if (keyword) {
      query += " AND (title LIKE ? OR description LIKE ?)";
      const keywordParam = `%${keyword}%`;
      params.push(keywordParam, keywordParam);
    }

    // 給与範囲検索
    if (min_salary) {
      query += " AND salary >= ?";
      params.push(min_salary);
    }
    if (max_salary) {
      query += " AND salary <= ?";
      params.push(max_salary);
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
    if (tag) {
      query +=
        " AND (tags LIKE ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?)";
      params.push(`${tag},%`, `%,${tag},%`, `%,${tag}`, tag);
    }

    // ソート順指定
    query += " ORDER BY updated_at DESC, id DESC";

    type JobRow = RowDataPacket & Job;
    // クエリを実行
    const [jobRows] = await pool.query<JobRow[]>(query, params);

    for (const jobRow of jobRows) {
      const job = {
        id: jobRow.id,
        title: jobRow.title,
        description: jobRow.description,
        salary: jobRow.salary,
        tags: jobRow.tags,
        created_at: jobRow.created_at,
        updated_at: jobRow.updated_at,
      } satisfies Job;
      jobs.push(job);
    }
  } catch (err) {
    console.error("Error searching jobs:", err);
    return res.status(500).json({ message: "Error searching jobs" });
  }

  try {
    type Company = {
      id: number;
      name: string;
      industry: string;
    };
    type JobWithCompany = Job & { company: Company };

    // 求人ごとの企業情報を取得
    const jobsWithCompany: JobWithCompany[] = [];

    type CompanyRow = RowDataPacket &
      Company & {
        industry_id: string;
      };
    for (const job of jobs) {
      const [companyRows] = await pool.query<CompanyRow[]>(
        "SELECT company.id, company.name, industry_category.name as industry, company.industry_id FROM company JOIN industry_category ON company.industry_id = industry_category.id WHERE company.id = (SELECT company_id FROM user WHERE id = (SELECT create_user_id FROM job WHERE id = ?))",
        [job.id]
      );

      if (!companyRows.length) continue;

      const company = companyRows[0];
      // 検索時に指定した業種と違う場合はスキップ
      if (industry_id !== "" && industry_id !== company.industry_id) {
        continue;
      }

      jobsWithCompany.push({ ...job, company });
    }

    type JobSearchResponse = {
      jobs: JobWithCompany[];
      page: number;
      has_next_page: boolean;
    };
    const resp: JobSearchResponse = {
      jobs: [],
      page,
      has_next_page: false,
    };
    for (const [i, job] of jobsWithCompany.entries()) {
      if (i < page * JOB_SEARCH_PAGE_SIZE) {
        continue;
      }

      if (resp.jobs.length >= JOB_SEARCH_PAGE_SIZE) {
        resp.has_next_page = true;
        break;
      }

      resp.jobs.push(job);
    }

    return res.status(200).json(resp);
  } catch (err) {
    console.error("Error fetch company from db:", err);
    return res.status(500).json({ message: "Error searching jobs" });
  }
});

// CS求人への応募API
app.post("/api/cs/application", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  // リクエストパラメータを取得
  type CsApplicationRequest = {
    job_id: number;
  };
  const { job_id } = req.body as CsApplicationRequest;
  if (!job_id) {
    return res.status(404).json("Invalid request payload");
  }

  try {
    type User = {
      id: number;
      user_type: string;
    };
    let user: User;
    type UserRow = RowDataPacket & User;
    // ユーザー情報をDBから取得
    const [userRows] = await pool.query<UserRow[]>(
      "SELECT id, user_type FROM user WHERE email = ?",
      [email]
    );

    user = userRows[0];
    // CSユーザーでなければ403を返す
    if (user?.user_type !== "CS") {
      return res.status(403).json("Forbidden");
    }

    // 求人が応募可能か確認すると同時にロックを取得
    const [canApplyRows] = await pool.query<
      (RowDataPacket & { can_apply: boolean })[]
    >(
      "SELECT (is_active = true AND is_archived = false) AS can_apply FROM job WHERE id = ? FOR UPDATE",
      [job_id]
    );
    if (!canApplyRows.length) {
      return res.status(404).json("Job not found");
    }

    const canApply = canApplyRows[0].can_apply;
    if (!Boolean(canApply)) {
      return res.status(422).json("Job is not accepting applications");
    }

    // 応募済みかどうか確認
    const [existsRows] = await pool.query<
      (RowDataPacket & { is_exists: boolean })[]
    >(
      "SELECT EXISTS (SELECT 1 FROM application WHERE job_id = ? AND user_id = ?) AS is_exists",
      [job_id, user.id]
    );
    const exists = existsRows[0]?.is_exists;
    if (Boolean(exists)) {
      return res.status(409).json("Already applied for the job");
    }

    // データベースに応募情報を挿入
    const [result] = await pool.query<ResultSetHeader>(
      "INSERT INTO application (job_id, user_id) VALUES (?, ?)",
      [job_id, user.id]
    );

    const applicationId = result.insertId;

    return res
      .status(200)
      .json({ message: "Successfully applied for the job", id: applicationId });
  } catch (err) {
    console.error("Error applying for job:", err);
    return res.status(500).json("Error applying for job");
  }
});

// CS応募一覧取得API
app.get("/api/cs/applications", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  // リクエストパラメータを取得
  type CsApplicationListRequest = {
    page: string | null | number; // 0-indexed
  };
  let { page } = req.query as CsApplicationListRequest;
  page = page ? Number(page) : 0;

  try {
    type Application = {
      id: number;
      job_id: number;
      user_id: number;
      created_at: Date;
    };
    type ApplicationRow = RowDataPacket & Application;
    // 応募一覧を取得
    const [applicationRows] = await pool.query<ApplicationRow[]>(
      "SELECT a.id, a.job_id, a.user_id, a.created_at FROM application a JOIN user u ON a.user_id = u.id WHERE u.email = ? ORDER BY a.created_at DESC",
      [email]
    );

    type Job = {
      id: number;
      title: string;
      description: string;
      salary: number;
      tags: string;
      created_at: Date;
      updated_at: Date;
    };
    type ApplicationWithJob = Application & { job?: Job };
    const applicationWithJobs: ApplicationWithJob[] = [];
    for (const applicationRow of applicationRows) {
      const application = {
        id: applicationRow.id,
        job_id: applicationRow.job_id,
        user_id: applicationRow.user_id,
        created_at: applicationRow.created_at,
      } satisfies ApplicationWithJob;
      applicationWithJobs.push(application);
    }

    // 求人情報を取得
    type JobRow = RowDataPacket & Job;
    for (const [i, applicationWithJob] of applicationWithJobs.entries()) {
      const [jobRows] = await pool.query<JobRow[]>(
        "SELECT id, title, description, salary, tags, created_at, updated_at FROM job WHERE id = ?",
        [applicationWithJob.job_id]
      );

      if (jobRows.length > 0) {
        applicationWithJobs[i].job = jobRows[0];
      }
    }

    type ApplicationsResponse = {
      applications: ApplicationWithJob[];
      page: number;
      has_next_page: boolean;
    };
    const resp: ApplicationsResponse = {
      applications: [],
      page: page,
      has_next_page: false,
    };
    for (const [i, applicationWithJob] of applicationWithJobs.entries()) {
      if (i < page * APPLICATION_LIST_PAGE_SIZE) {
        continue;
      }

      if (resp.applications.length >= APPLICATION_LIST_PAGE_SIZE) {
        resp.has_next_page = true;
        break;
      }

      resp.applications.push(applicationWithJob);
    }

    return res.status(200).json(resp);
  } catch (err) {
    console.error("Error querying database:", err);
    return res.status(500).json("Error getting applications");
  }
});

// CL企業登録API
app.post("/api/cl/company", async (req: Request, res: Response) => {
  // リクエストパラメータを取得
  type ClCompanyRequest = {
    name: string;
    industry_id: string;
  };
  const { name, industry_id } = req.body as ClCompanyRequest;
  if (!name || !industry_id) {
    return res.status(400).json({ message: "Invalid request payload" });
  }

  try {
    // 企業をデータベースに登録
    const [result] = await pool.query<ResultSetHeader>(
      "INSERT INTO company (name, industry_id) VALUES (?, ?)",
      [name, industry_id]
    );

    const companyId = result.insertId;

    return res
      .status(200)
      .json({ message: "Company created successfully", id: companyId });
  } catch (err) {
    console.error("Error creating company:", err);
    return res.status(500).json({ message: "Error creating company" });
  }
});

// CLアカウント作成API
app.post("/api/cl/signup", async (req: Request, res: Response) => {
  // リクエストパラメータを取得
  type ClSignupRequest = {
    email: string;
    password: string;
    name: string;
    company_id: number;
  };
  const { email, password, name, company_id } = req.body as ClSignupRequest;
  if (!email || !password || !name || !company_id) {
    return res.status(400).json({ message: "Invalid request payload" });
  }

  // パスワードをハッシュ化
  let hashedPassword: string;
  try {
    hashedPassword = bcrypt.hashSync(password, 10);
  } catch (err) {
    console.error("Error hashing password:", err);
    return res.status(500).json({ message: "Error signing up" });
  }

  let userId: number;
  try {
    // ユーザーをデータベースに登録
    const [result] = await pool.query<ResultSetHeader>(
      "INSERT INTO user (email, password, name, user_type, company_id) VALUES (?, ?, ?, ?, ?)",
      [email, hashedPassword, name, "CL", company_id]
    );

    userId = result.insertId;
  } catch (err) {
    // 登録済みの場合は409を返す
    if (err instanceof Error && (err as QueryError).code === "ER_DUP_ENTRY") {
      return res.status(409).json({ message: "Email address is already used" });
    }

    // 存在しない企業IDの場合は400を返す
    if (
      err instanceof Error &&
      (err as QueryError).code === "ER_NO_REFERENCED_ROW_2"
    ) {
      return res.status(400).json({ message: "Company not found" });
    }

    console.error("Error creating user:", err);
    return res.status(500).json({ message: "Error signing up" });
  }

  // セッションを作成
  const ok = setSession(req);
  if (!ok) {
    return res.status(500).json({ message: "Error creating account" });
  }

  return res
    .status(200)
    .json({ message: "Signed up successfully", id: userId });
});

// CLログインAPI
app.post("/api/cl/login", async (req: Request, res: Response) => {
  // リクエストパラメータを取得
  type ClLoginRequest = {
    email: string;
    password: string;
  };
  const { email, password } = req.body as ClLoginRequest;
  if (!email || !password) {
    return res.status(400).json({ message: "Invalid request payload" });
  }

  let storedPassword: string;
  try {
    type User = {
      password: string;
    };
    type UserRow = RowDataPacket & User;
    // パスワードをDBから取得
    const [rows] = await pool.query<UserRow[]>(
      "SELECT password FROM user WHERE email = ? AND user_type = ?",
      [email, "CL"]
    );
    if (!rows.length) {
      return res.status(401).json({ message: "Invalid email or password" });
    }
    storedPassword = rows[0].password;
  } catch (err) {
    console.error("Error querying database:", err);
    return res.status(500).json({ message: "Error logging in" });
  }

  // パスワードを比較
  const isMatch = bcrypt.compareSync(password, storedPassword);
  if (!isMatch) {
    return res.status(401).json({ message: "Invalid email or password" });
  }

  // セッションを作成
  const ok = setSession(req);
  if (!ok) {
    return res.status(500).json({ message: "Error creating account" });
  }

  return res.status(200).json({ message: "Logged in successfully" });
});

// CLログアウトAPI
app.post("/api/cl/logout", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  // ログアウト処理
  const ok = deleteSession(req);
  if (!ok) {
    return res.status(500).json({ message: "Error creating account" });
  }

  return res.status(200).json("Logged out successfully");
});

// CL求人作成API
app.post("/api/cl/job", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  try {
    // ユーザーをDBから取得
    type User = {
      id: number;
      email: string;
      password: string;
      name: string;
      user_type: string;
      company_id: number;
    };
    type UserRow = RowDataPacket & User;
    const [userRows] = await pool.query<UserRow[]>(
      "SELECT id, email, password, name, user_type, company_id FROM user WHERE email = ?",
      [email]
    );

    if (!userRows.length) {
      console.error("Session user not found");
      return res.status(500).json({ message: "Error creating job" });
    }

    // 企業アカウントでなければ403を返す
    const user = userRows[0];
    if (user.user_type !== "CL") {
      return res.status(403).json({ message: "No permission" });
    }

    // リクエストパラメータを取得
    type ClJobRequest = {
      title: string;
      description: string;
      salary: number;
      tags: string;
    };
    const { title, description, salary, tags } = req.body as ClJobRequest;
    if (!title || !description || !salary || !tags) {
      return res.status(400).json({ message: "Invalid request payload" });
    }

    // 求人をデータベースに登録
    const [result] = await pool.query<ResultSetHeader>(
      "INSERT INTO job (title, description, salary, tags, is_active, create_user_id) VALUES (?, ?, ?, ?, true, ?)",
      [title, description, salary, tags, user.id]
    );

    const jobId = result.insertId;
    return res
      .status(200)
      .json({ message: "Job created successfully", id: jobId });
  } catch (err) {
    console.error("Error creating job:", err);
    return res.status(500).json({ message: "Error creating job" });
  }
});

// ログインユーザーが求人を閲覧・編集できるかどうかチェックするための関数
async function canAccessJob(
  jobId: string,
  email: string,
  includeArchived: boolean
) {
  type CLUser = {
    user_type: string;
    company_id: number;
  };
  let user: CLUser;
  try {
    type UserRow = RowDataPacket & CLUser;
    // ログインユーザーを取得
    const [userRows] = await pool.query<UserRow[]>(
      "SELECT user_type, company_id FROM user WHERE email = ?",
      [email]
    );

    user = userRows[0];
    if (!user) {
      return { ok: false, status: 500, message: "Error updating job" };
    }

    // 企業アカウントでなければ403を返す
    if (user.user_type !== "CL") {
      return { ok: false, status: 403, message: "No permission" };
    }
  } catch (err) {
    console.error("Error fetch user from db:", err);
    return { ok: false, status: 500, message: "No permission" };
  }

  type Job = {
    create_user_id: number;
    is_archived: boolean;
  };
  let job: Job;
  try {
    type JobRow = RowDataPacket & Job;
    // 求人を取得して存在するかチェック
    const [jobRows] = await pool.query<JobRow[]>(
      "SELECT create_user_id, is_archived FROM job WHERE id = ?",
      [jobId]
    );

    job = jobRows[0];
    if (!job) {
      return { ok: false, status: 404, message: "Job not found" };
    }

    if (!includeArchived && Boolean(job.is_archived)) {
      return { ok: false, status: 422, message: "Job archived" };
    }
  } catch (err) {
    console.error("Error fetch job from database:", err);
    return { ok: false, status: 500, message: "Error updating job" };
  }

  try {
    type UserRow = RowDataPacket & {
      company_id: number;
    };
    // 求人作成ユーザーを取得
    const [userRow] = await pool.query<UserRow[]>(
      "SELECT company_id FROM user WHERE id = ?",
      [job.create_user_id]
    );

    const jobCreateUser = userRow[0];
    if (!jobCreateUser) {
      return { ok: false, status: 500, message: "Error getting job" };
    }

    // 求人作成ユーザーとログインユーザーの所属会社が異なる場合は403を返す
    if (jobCreateUser.company_id !== user.company_id) {
      return { ok: false, status: 403, message: "No permission" };
    }

    return { ok: true, status: 200, message: "success" };
  } catch (err) {
    console.error("Error fetch job create user from database:", err);
    return { ok: false, status: 500, message: "Error getting job" };
  }
}

// CL求人更新API
app.patch("/api/cl/job/:jobid", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  // リクエストパラメータを取得
  type ClUpdateJobRequest = {
    title: string | null;
    description: string | null;
    salary: number | null;
    tags: string | null;
    is_active: boolean | null;
  };
  const { title, description, salary, tags, is_active } =
    req.body as ClUpdateJobRequest;
  const jobId = req.params.jobid;

  // 編集できるかどうかチェック
  const { ok, status, message } = await canAccessJob(jobId, email, false);
  if (!ok) {
    return res.status(status).json(message);
  }

  try {
    // 求人情報を更新
    let query = "UPDATE job SET";
    const params = [];
    if (title !== null) {
      query += " title = ?,";
      params.push(title);
    }
    if (description !== null) {
      query += " description = ?,";
      params.push(description);
    }
    if (salary !== null) {
      query += " salary = ?,";
      params.push(salary);
    }
    if (tags !== null) {
      query += " tags = ?,";
      params.push(tags);
    }
    if (is_active !== null) {
      query += " is_active = ?,";
      params.push(Boolean(is_active));
    }

    query = query.slice(0, -1) + " WHERE id = ?";
    params.push(jobId);

    await pool.query(query, params);

    return res.status(200).json({ message: "Job updated successfully" });
  } catch (err) {
    console.error("Error updating job:", err);
    return res.status(500).json({ message: "Error updating job" });
  }
});

// CL求人アーカイブAPI
// 求人をアーカイブすると
// GET /cs/job_search, GET /cl/jobs で取得できなくなる
// GET /cs/applications, GET /cl/job/:jobid では引き続き取得可能
app.post("/api/cl/job/:jobid/archive", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  // リクエストパラメータを取得
  const jobId = req.params.jobid;

  // アーカイブできるかどうかチェック
  const { ok, status, message } = await canAccessJob(jobId, email, false);
  if (!ok) {
    return res.status(status).json(message);
  }

  try {
    // 求人をアーカイブ
    await pool.query("UPDATE job SET is_archived = true WHERE id = ?", [jobId]);

    return res.status(200).json({ message: "Job archived successfully" });
  } catch (err) {
    console.error("Error archiving job:", err);
    return res.status(500).json({ message: "Error archiving job" });
  }
});

// CL求人取得API
app.get("/api/cl/job/:jobid", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  // リクエストパラメータを取得
  const jobId = req.params.jobid;

  // 閲覧できるかどうかチェック
  const { ok, status, message } = await canAccessJob(jobId, email, true);
  if (!ok) {
    return res.status(status).json(message);
  }

  try {
    type Job = {
      id: number;
      title: string;
      description: string;
      salary: number;
      tags: string;
      is_active: boolean;
      create_user_id: number;
      created_at: Date;
      updated_at: Date;
    };
    type JobRow = RowDataPacket & Job;
    // 求人を取得
    const [jobRows] = await pool.query<JobRow[]>(
      "SELECT id, title, description, salary, tags, is_active, create_user_id, created_at, updated_at FROM job WHERE id = ?",
      [jobId]
    );
    if (!jobRows.length) {
      return res.status(404).json({ message: "Job not found" });
    }
    const job = jobRows[0];
    job.is_active = Boolean(job.is_active);

    type Application = {
      id: number;
      job_id: number;
      user_id: number;
      created_at: Date;
    };
    type ApplicationRow = RowDataPacket & Application;
    // 求人への応募を取得
    const [applicationRows] = await pool.query<ApplicationRow[]>(
      "SELECT id, job_id, user_id, created_at FROM application WHERE job_id = ? ORDER BY created_at",
      [jobId]
    );

    type User = {
      id: number;
      email: string;
      name: string;
    };
    type ApplicationWithUser = Application & { applicant?: User };
    const applications: ApplicationWithUser[] = [];
    for (const applicationRow of applicationRows) {
      const application = {
        id: applicationRow.id,
        job_id: applicationRow.job_id,
        user_id: applicationRow.user_id,
        created_at: applicationRow.created_at,
      } satisfies ApplicationWithUser;
      applications.push(application);
    }

    type UserRow = RowDataPacket & User;
    for (const [i, application] of applications.entries()) {
      // 応募者の情報を取得
      const [userRows] = await pool.query<UserRow[]>(
        "SELECT id, email, name FROM user WHERE id = ?",
        [application.user_id]
      );
      if (userRows.length > 0) {
        const user = userRows[0];
        applications[i].applicant = {
          id: user.id,
          email: user.email,
          name: user.name,
        };
      }
    }

    job.applications = applications;

    return res.status(200).json(job);
  } catch (error) {
    console.error("Error querying database:", error);
    return res.status(500).json({ message: "Error getting job" });
  }
});

// CL求人一覧取得API
app.get("/api/cl/jobs", async (req: Request, res: Response) => {
  // ログイン認証
  const email = getSession(req);
  if (!email) {
    return res.status(401).json("Not logged in");
  }

  type CLUser = {
    id: number;
    user_type: string;
    company_id: number;
  };
  let user: CLUser;
  try {
    type UserRow = RowDataPacket & CLUser;
    // ログインユーザーを取得
    const [userRows] = await pool.query<UserRow[]>(
      "SELECT id, user_type, company_id FROM user WHERE email = ?",
      [email]
    );

    if (!userRows.length) {
      return res.status(404).json({ message: "User not found" });
    }

    user = userRows[0];
    // 企業アカウントでなければ403を返す
    if (user.user_type !== "CL") {
      return res.status(403).json({ message: "No permission" });
    }
  } catch (error) {
    console.error("Error fetch user from db:", error);
    return res.status(500).json({ message: "Error getting jobs" });
  }

  try {
    // リクエストパラメータを取得
    type JobListRequest = {
      page: string | null | number; // 0-indexed
    };
    // リクエストパラメータを取得
    let { page } = req.query as JobListRequest;
    page = page ? Number(page) : 0;

    type Job = {
      id: number;
      title: string;
      description: string;
      salary: number;
      tags: string;
      is_active: boolean;
      create_user_id: number;
      created_at: Date;
      updated_at: Date;
    };
    type JobRow = RowDataPacket & Job;
    // 求人一覧を取得
    const [jobRows] = await pool.query<JobRow[]>(
      "SELECT id, title, description, salary, tags, is_active, create_user_id, created_at, updated_at FROM job WHERE is_archived = false AND create_user_id IN (SELECT id FROM user WHERE company_id = ?) ORDER BY updated_at DESC, id",
      [user.company_id]
    );

    type JobListResponse = {
      jobs: Job[];
      page: number;
      has_next_page: boolean;
    };
    const resp: JobListResponse = { jobs: [], page, has_next_page: false };

    let i = 0;
    for (const jobRow of jobRows) {
      if (i < page * JOB_LIST_PAGE_SIZE) {
        i++;
        continue;
      }

      if (resp.jobs.length >= JOB_LIST_PAGE_SIZE) {
        resp.has_next_page = true;
        break;
      }

      const job = {
        id: jobRow.id,
        title: jobRow.title,
        description: jobRow.description,
        salary: jobRow.salary,
        tags: jobRow.tags,
        is_active: Boolean(jobRow.is_active),
        create_user_id: jobRow.create_user_id,
        created_at: jobRow.created_at,
        updated_at: jobRow.updated_at,
      } satisfies Job;
      resp.jobs.push(job);
      i++;
    }

    return res.status(200).json(resp);
  } catch (error) {
    console.error("Error querying database:", error);
    return res.status(500).json({ message: "Error getting jobs" });
  }
});

app.use(XRayExpress.closeSegment());

const port = 8080;
app.listen(port, () => {
  console.log(`[server]: Server is running at http://localhost:${port}`);
});
