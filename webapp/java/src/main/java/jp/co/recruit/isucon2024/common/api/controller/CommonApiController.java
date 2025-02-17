package jp.co.recruit.isucon2024.common.api.controller;

import com.amazonaws.xray.spring.aop.XRayEnabled;
import jp.co.recruit.isucon2024.common.api.response.InitializeResponse;
import jp.co.recruit.isucon2024.common.api.service.InitializeService;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RestController;

import java.io.IOException;

@RestController
@RequiredArgsConstructor
@XRayEnabled
public class CommonApiController {

    private final InitializeService initializeService;

    // ベンチマーカー向けAPI
    @PostMapping("/api/initialize")
    public ApiResponse<InitializeResponse> initialize() throws IOException, InterruptedException {
        return initializeService.initialize();
    }

    // ベンチマーカー向けAPI
    @PostMapping("/api/finalize")
    public ApiResponse<String> finalizer() {
        return new ApiResponse<>("ok", HttpStatus.OK);
    }
}
