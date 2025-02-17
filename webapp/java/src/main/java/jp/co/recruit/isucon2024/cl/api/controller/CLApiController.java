package jp.co.recruit.isucon2024.cl.api.controller;

import com.amazonaws.xray.spring.aop.XRayEnabled;
import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.cl.api.form.ClSignupForm;
import jp.co.recruit.isucon2024.cl.api.form.CreateCompanyForm;
import jp.co.recruit.isucon2024.cl.api.form.CreateJobForm;
import jp.co.recruit.isucon2024.cl.api.form.UpdateJobForm;
import jp.co.recruit.isucon2024.cl.api.response.ClSingUpResponse;
import jp.co.recruit.isucon2024.cl.api.response.CreateCompanyResponse;
import jp.co.recruit.isucon2024.cl.api.response.CreateJobResponse;
import jp.co.recruit.isucon2024.cl.api.response.GetJobResponse;
import jp.co.recruit.isucon2024.cl.api.response.GetJobsResponse;
import jp.co.recruit.isucon2024.cl.api.service.ArchiveJobService;
import jp.co.recruit.isucon2024.cl.api.service.ClLoginService;
import jp.co.recruit.isucon2024.cl.api.service.ClLogoutService;
import jp.co.recruit.isucon2024.cl.api.service.ClSignupService;
import jp.co.recruit.isucon2024.cl.api.service.CreateCompanyService;
import jp.co.recruit.isucon2024.cl.api.service.CreateJobService;
import jp.co.recruit.isucon2024.cl.api.service.GetJobService;
import jp.co.recruit.isucon2024.cl.api.service.GetJobsService;
import jp.co.recruit.isucon2024.cl.api.service.UpdateJobService;
import jp.co.recruit.isucon2024.common.form.LoginForm;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequiredArgsConstructor
@XRayEnabled
public class CLApiController {

    private final ClLoginService clLoginService;
    private final ClSignupService clSignupService;
    private final ClLogoutService clLogoutService;
    private final CreateCompanyService createCompanyService;
    private final CreateJobService createJobService;
    private final UpdateJobService updateJobService;
    private final ArchiveJobService archiveJobService;
    private final GetJobService getJobService;
    private final GetJobsService getJobsService;

    // CLログインAPI
    @PostMapping("/api/cl/login")
    public ApiResponse<String> login(HttpServletRequest request, @RequestBody LoginForm form) {
        return clLoginService.login(request, form);
    }

    // CLログアウトAPI
    @PostMapping("/api/cl/logout")
    public ApiResponse<String> logout(HttpServletRequest request) {
        return clLogoutService.logout(request);
    }

    // CLアカウント作成API
    @PostMapping("/api/cl/signup")
    public ApiResponse<ClSingUpResponse> signup(HttpServletRequest request, @RequestBody ClSignupForm form) {
        return clSignupService.signup(request, form);
    }

    // CL企業登録API
    @PostMapping("/api/cl/company")
    public ApiResponse<CreateCompanyResponse> createCompany(@RequestBody CreateCompanyForm form) {
        return createCompanyService.createCompany(form);
    }

    // CL求人作成API
    @PostMapping("/api/cl/job")
    public ApiResponse<CreateJobResponse> createJob(HttpServletRequest request, @RequestBody CreateJobForm form) {
        return createJobService.createJob(request, form);
    }

    // CL求人更新API
    @PatchMapping("/api/cl/job/{jobId}")
    public ApiResponse<String> updateJob(
            HttpServletRequest request, @PathVariable int jobId, @RequestBody UpdateJobForm form) {
        return updateJobService.updateJob(request, jobId, form);
    }

    // CL求人アーカイブAPI
    @PostMapping("/api/cl/job/{jobId}/archive")
    public ApiResponse<String> archiveJob(HttpServletRequest request, @PathVariable int jobId) {
        return archiveJobService.archiveJob(request, jobId);
    }

    // CL求人取得API
    @GetMapping("/api/cl/job/{jobId}")
    public ApiResponse<GetJobResponse> getJob(HttpServletRequest request, @PathVariable int jobId) {
        return getJobService.getJob(request, jobId);
    }

    // CL求人一覧取得API
    @GetMapping("/api/cl/jobs")
    public ApiResponse<GetJobsResponse> getJobs(
            HttpServletRequest request,
            @RequestParam(name = "page", required = false, defaultValue = "0") int page) {
        return getJobsService.getJobs(request, page);
    }

}
