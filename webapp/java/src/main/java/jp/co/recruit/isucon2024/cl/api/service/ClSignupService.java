package jp.co.recruit.isucon2024.cl.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.cl.api.form.ClSignupForm;
import jp.co.recruit.isucon2024.cl.api.response.ClSingUpResponse;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.dao.DataIntegrityViolationException;
import org.springframework.dao.DuplicateKeyException;
import org.springframework.http.HttpStatus;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

@Service
@RequiredArgsConstructor
public class ClSignupService {

    private final UserDao userDao;
    private final PasswordEncoder passwordEncoder;

    @Transactional(propagation = Propagation.REQUIRED, rollbackFor = Exception.class)
    public ApiResponse<ClSingUpResponse> signup(HttpServletRequest request, ClSignupForm form) {

        // パスワードをハッシュ化
        String hashedPassword = passwordEncoder.encode(form.getPassword());

        // アカウントを作成
        try {
            int userId = userDao.insertClUser(form.getEmail(), hashedPassword, form.getName(), form.getCompany_id());

            ClSingUpResponse response = new ClSingUpResponse();
            response.setId(userId);
            response.setMessage("Signed up successfully");

            // セッションを作成
            SessionUtil.setEmailToSession(request, form.getEmail());
            return new ApiResponse<>(response, HttpStatus.OK);
        } catch (DuplicateKeyException duplicateKeyException) {
            // 登録済みの場合は409を返す
            return new ApiResponse<>("Email address is already used", HttpStatus.CONFLICT);
        } catch (DataIntegrityViolationException dataIntegrityViolationException) {
            // 会社IDが存在しない場合は400を返す
            return new ApiResponse<>("Company not found", HttpStatus.BAD_REQUEST);
        }

    }

}
