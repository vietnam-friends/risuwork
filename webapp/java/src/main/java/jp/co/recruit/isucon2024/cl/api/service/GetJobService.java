package jp.co.recruit.isucon2024.cl.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.cl.api.dao.GetJobDao;
import jp.co.recruit.isucon2024.cl.api.response.GetJobResponse;
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

import java.util.List;
import java.util.Objects;

@Service
@RequiredArgsConstructor
public class GetJobService {

    private final UserDao userDao;
    private final JobDao jobDao;
    private final GetJobDao getJobDao;

    public ApiResponse<GetJobResponse> getJob(HttpServletRequest request, int jobId) {

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

        // N+1問題解決: 応募情報とユーザー情報を一度に取得
        List<GetJobResponse.Application> applications = getJobDao.selectApplicationsWithUserByJobId(jobId);

        return new ApiResponse<>(toResponse(job, applications), HttpStatus.OK);

    }

    private GetJobResponse.CSUser toCSUser(UserEntity userEntity) {
        GetJobResponse.CSUser user = new GetJobResponse.CSUser();
        user.setId(userEntity.getId());
        user.setEmail(userEntity.getEmail());
        user.setName(userEntity.getName());
        return user;
    }

    private GetJobResponse toResponse(JobEntity job, List<GetJobResponse.Application> applications) {
        GetJobResponse response = new GetJobResponse();
        response.setId(job.getId());
        response.setTitle(job.getTitle());
        response.setDescription(job.getDescription());
        response.setSalary(job.getSalary());
        response.setTags(job.getTags());
        response.setIs_active(job.getIs_active());
        response.setCreate_user_id(job.getCreate_user_id());
        response.setCreated_at(job.getCreated_at());
        response.setUpdated_at(job.getUpdated_at());
        response.setApplications(applications);
        return response;
    }

}