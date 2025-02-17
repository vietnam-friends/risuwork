package jp.co.recruit.isucon2024.cs.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

@Service
public class CsLogoutService {

    public ApiResponse<String> logout(HttpServletRequest request) {
        String email = SessionUtil.getEmailFromSession(request);
        if (email == null || email.isEmpty()) {
            return new ApiResponse<>("Not logged in", HttpStatus.UNAUTHORIZED);
        }
        SessionUtil.removeEmailFromSession(request);
        return new ApiResponse<>("Logged out successfully", HttpStatus.OK);
    }

}
