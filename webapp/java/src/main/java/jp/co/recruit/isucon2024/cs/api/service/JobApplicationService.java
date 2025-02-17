package jp.co.recruit.isucon2024.cs.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.entity.UserEntity;
import jp.co.recruit.isucon2024.common.enums.UserTypeEnum;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import jp.co.recruit.isucon2024.cs.api.dao.JobApplicationDao;
import jp.co.recruit.isucon2024.cs.api.form.JobApplicationForm;
import jp.co.recruit.isucon2024.cs.api.response.JobApplicationResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.dao.EmptyResultDataAccessException;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.util.Objects;

@Service
@Transactional(propagation = Propagation.REQUIRED, rollbackFor = Exception.class)
@RequiredArgsConstructor
@Slf4j
public class JobApplicationService {

    private final JobApplicationDao jobApplicationDao;
    private final UserDao userDao;

    public ApiResponse<JobApplicationResponse> application(HttpServletRequest request, JobApplicationForm form) {

        String email = SessionUtil.getEmailFromSession(request);
        if (email == null || email.isEmpty()) {
            return new ApiResponse<>("Not logged in", HttpStatus.UNAUTHORIZED);
        }

        UserEntity user = userDao.selectByEmail(email);
        // CSユーザーでなければ403を返す
        if (user == null || !Objects.equals(user.getUserType(), UserTypeEnum.CS.name())) {
            return new ApiResponse<>("Forbidden", HttpStatus.FORBIDDEN);
        }

        // 求人が応募可能か確認すると同時にロックを取得
        try {
            Boolean canApply = jobApplicationDao.selectJobByJobIdForUpdate(form.getJob_id());
            if (!canApply) {
                return new ApiResponse<>("Job is not accepting applications", HttpStatus.UNPROCESSABLE_ENTITY);
            }
        } catch (EmptyResultDataAccessException e) {
            return new ApiResponse<>("Job not found", HttpStatus.NOT_FOUND);
        }

        // 応募済みかどうか確認
        Boolean exists = jobApplicationDao.selectExistApplicationByJobIdAndUserId(form.getJob_id(), user.getId());
        if (exists != null && exists) {
            return new ApiResponse<>("Already applied for the job", HttpStatus.CONFLICT);
        }

        // データベースに応募情報を挿入
        int applicationId = jobApplicationDao.insertJobApplication(form.getJob_id(), user.getId());

        JobApplicationResponse response = new JobApplicationResponse();
        response.setId(applicationId);
        response.setMessage("Successfully applied for the job");

        return new ApiResponse<>(response, HttpStatus.OK);
    }

}
