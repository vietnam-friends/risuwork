package jp.co.recruit.isucon2024.cs.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import jp.co.recruit.isucon2024.cs.api.form.CsSignupForm;
import jp.co.recruit.isucon2024.cs.api.response.CsSignUpResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.dao.DuplicateKeyException;
import org.springframework.http.HttpStatus;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

@Service
@RequiredArgsConstructor
@Slf4j
public class CsSignupService {

    private final UserDao userDao;
    private final PasswordEncoder passwordEncoder;

    @Transactional(propagation = Propagation.REQUIRED, rollbackFor = Exception.class)
    public ApiResponse<CsSignUpResponse> signup(HttpServletRequest request, CsSignupForm form) {

        // パスワードをハッシュ化
        String hashedPassword = passwordEncoder.encode(form.getPassword());

        // アカウントを作成
        try {
            int userId = userDao.insertCsUser(form.getEmail(), hashedPassword, form.getName());
            // セッションを作成
            SessionUtil.setEmailToSession(request, form.getEmail());
            CsSignUpResponse response = new CsSignUpResponse();
            response.setId(userId);
            response.setMessage("CS account created successfully");
            return new ApiResponse<>(response, HttpStatus.OK);
        } catch (DuplicateKeyException e) {
            // 登録済みの場合は409を返す
            return new ApiResponse<>("Email address is already used", HttpStatus.CONFLICT);
        }


    }
}
