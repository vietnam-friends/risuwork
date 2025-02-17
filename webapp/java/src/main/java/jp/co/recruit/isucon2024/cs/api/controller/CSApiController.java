package jp.co.recruit.isucon2024.cs.api.controller;

import com.amazonaws.xray.spring.aop.XRayEnabled;
import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.common.form.LoginForm;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.cs.api.form.CsSignupForm;
import jp.co.recruit.isucon2024.cs.api.form.JobApplicationForm;
import jp.co.recruit.isucon2024.cs.api.form.JobSearchForm;
import jp.co.recruit.isucon2024.cs.api.response.ApplicationsResponse;
import jp.co.recruit.isucon2024.cs.api.response.CsSignUpResponse;
import jp.co.recruit.isucon2024.cs.api.response.JobApplicationResponse;
import jp.co.recruit.isucon2024.cs.api.response.JobSearchResponse;
import jp.co.recruit.isucon2024.cs.api.service.ApplicationsService;
import jp.co.recruit.isucon2024.cs.api.service.CsLoginService;
import jp.co.recruit.isucon2024.cs.api.service.CsLogoutService;
import jp.co.recruit.isucon2024.cs.api.service.CsSignupService;
import jp.co.recruit.isucon2024.cs.api.service.JobApplicationService;
import jp.co.recruit.isucon2024.cs.api.service.JobSearchService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ModelAttribute;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequiredArgsConstructor
@XRayEnabled
public class CSApiController {

    private final CsLoginService loginService;
    private final CsSignupService signupService;
    private final CsLogoutService logoutService;
    private final JobSearchService jobSearchService;
    private final JobApplicationService jobApplicationService;
    private final ApplicationsService applicationsService;

    // CSログインAPI
    @PostMapping("/api/cs/login")
    public ApiResponse<String> login(HttpServletRequest request, @RequestBody LoginForm form) {
        return loginService.login(request, form);
    }

    // CSアカウント作成API
    @PostMapping("/api/cs/signup")
    public ApiResponse<CsSignUpResponse> signup(HttpServletRequest request, @RequestBody CsSignupForm form) {
        return signupService.signup(request, form);
    }

    // CSログアウトAPI
    @PostMapping("/api/cs/logout")
    public ApiResponse<String> logout(HttpServletRequest request) {
        return logoutService.logout(request);
    }

    // CS求人検索API
    @GetMapping("/api/cs/job_search")
    public ApiResponse<JobSearchResponse> jobSearch(@ModelAttribute JobSearchForm form) {
        return jobSearchService.jobSearch(form);
    }

    // CS求人応募API
    @PostMapping("/api/cs/application")
    public ApiResponse<JobApplicationResponse> jobApplication(HttpServletRequest request, @RequestBody JobApplicationForm form) {
        return jobApplicationService.application(request, form);
    }

    // CS応募一覧取得API
    @GetMapping("/api/cs/applications")
    public ApiResponse<ApplicationsResponse> applications(
            HttpServletRequest request,
            @RequestParam(name = "page", required = false, defaultValue = "0") int page) {
        return applicationsService.applications(request, page);
    }

}
