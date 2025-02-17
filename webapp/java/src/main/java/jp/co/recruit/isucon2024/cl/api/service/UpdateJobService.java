package jp.co.recruit.isucon2024.cl.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.cl.api.form.UpdateJobForm;
import jp.co.recruit.isucon2024.common.dao.JobDao;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.common.entity.UserEntity;
import jp.co.recruit.isucon2024.common.enums.UserTypeEnum;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import jp.co.recruit.isucon2024.cs.api.dao.UpdateJobDao;
import lombok.RequiredArgsConstructor;
import org.springframework.dao.EmptyResultDataAccessException;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.util.Objects;

@Service
@Transactional(propagation = Propagation.REQUIRED, rollbackFor = Exception.class)
@RequiredArgsConstructor
public class UpdateJobService {

    private final UpdateJobDao updateJobDao;
    private final UserDao userDao;
    private final JobDao jobDao;

    public ApiResponse<String> updateJob(HttpServletRequest request, int jobId, UpdateJobForm form) {

        // ログイン認証
        UserEntity user = userDao.selectByEmail(SessionUtil.getEmailFromSession(request));
        if (user == null) {
            return new ApiResponse<>("Not logged in", HttpStatus.UNAUTHORIZED);
        }
        // 企業アカウントでなければ403を返す
        if (!Objects.equals(user.getUserType(), UserTypeEnum.CL.name())) {
            return new ApiResponse<>("No permission", HttpStatus.FORBIDDEN);
        }

        // 求人を取得して存在するかチェック
        JobEntity job;
        try {
            job = jobDao.selectJobById(jobId);
            if (job.getIs_archived()) {
                return new ApiResponse<>("Job archived", HttpStatus.UNPROCESSABLE_ENTITY);
            }
        } catch (EmptyResultDataAccessException e) {
            return new ApiResponse<>("Job not found", HttpStatus.NOT_FOUND);
        }

        Integer companyId = userDao.selectUserCompanyIdByUserId(job.getCreate_user_id());
        if (companyId == null) {
            return new ApiResponse<>("Error getting job", HttpStatus.INTERNAL_SERVER_ERROR);
        }

        // 求人作成ユーザーとログインユーザーの所属会社が異なる場合は403を返す
        if (!Objects.equals(companyId, user.getCompanyId())) {
            return new ApiResponse<>("No permission", HttpStatus.FORBIDDEN);
        }

        // 求人情報を更新
        updateJobDao.updateJob(
                form.getTitle(),
                form.getDescription(),
                form.getSalary(),
                form.getTags(),
                form.getIs_active(),
                jobId);

        return new ApiResponse<>("Job updated successfully", HttpStatus.OK);
    }

}
