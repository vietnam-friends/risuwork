package jp.co.recruit.isucon2024.common.api.service;

import jp.co.recruit.isucon2024.common.api.response.InitializeResponse;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

import java.io.IOException;

@Service
public class InitializeService {

    @Value("${app.init-script-path}")
    private String initScriptPath;

    public ApiResponse<InitializeResponse> initialize() throws IOException, InterruptedException {

        ProcessBuilder processBuilder = new ProcessBuilder(initScriptPath);
        Process process = processBuilder.start();
        int exitCode = process.waitFor();

        if (exitCode != 0) {
            return new ApiResponse<>("Error initializing database", HttpStatus.INTERNAL_SERVER_ERROR);
        }

        return new ApiResponse<>(new InitializeResponse(), HttpStatus.OK);
    }

}
