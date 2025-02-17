package jp.co.recruit.isucon2024.cl.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.cl.api.dao.ArchiveJobDao;
import jp.co.recruit.isucon2024.common.dao.JobDao;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.common.entity.UserEntity;
import jp.co.recruit.isucon2024.common.enums.UserTypeEnum;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
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
public class ArchiveJobService {

    private final JobDao jobDao;
    private final UserDao userDao;
    private final ArchiveJobDao archiveJobDao;

    // 求人をアーカイブすると
    // GET /cs/job_search, GET /cl/jobs で取得できなくなる
    // GET /cs/applications, GET /cl/job/:jobid では引き続き取得可能
    public ApiResponse<String> archiveJob(HttpServletRequest request, int jobId) {

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
                return new ApiResponse<>("Job is archived", HttpStatus.BAD_REQUEST);
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

        // 求人をアーカイブ
        archiveJobDao.archiveJob(jobId);

        return new ApiResponse<>("Job archived successfully", HttpStatus.OK);

    }

}
