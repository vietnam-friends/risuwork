package jp.co.recruit.isucon2024.cl.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.entity.UserEntity;
import jp.co.recruit.isucon2024.common.form.LoginForm;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class ClLoginService {

    private final UserDao userDao;
    private final PasswordEncoder passwordEncoder;

    public ApiResponse<String> login(HttpServletRequest request, LoginForm form) {

        UserEntity user = userDao.selectClUserByEmail(form.getEmail());

        // パスワードをDBから取得
        if (user == null) {
            return new ApiResponse<>("Invalid email or password", HttpStatus.UNAUTHORIZED);
        }

        // パスワードを比較
        if (!passwordEncoder.matches(form.getPassword(), user.getPassword())) {
            return new ApiResponse<>("Invalid email or password", HttpStatus.UNAUTHORIZED);
        }

        // セッションを作成
        SessionUtil.setEmailToSession(request, form.getEmail());
        return new ApiResponse<>("Logged in successfully", HttpStatus.OK);
    }

}
