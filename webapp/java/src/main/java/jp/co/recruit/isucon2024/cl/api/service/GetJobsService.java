package jp.co.recruit.isucon2024.cl.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.cl.api.dao.GetJobsDao;
import jp.co.recruit.isucon2024.cl.api.response.GetJobsResponse;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.common.entity.UserEntity;
import jp.co.recruit.isucon2024.common.enums.UserTypeEnum;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Objects;

@Service
@RequiredArgsConstructor
public class GetJobsService {

    private final UserDao userDao;
    private final GetJobsDao getJobsDao;

    private final int JOB_LIST_PAGE_SIZE = 50;

    public ApiResponse<GetJobsResponse> getJobs(HttpServletRequest request, int page) {

        String email = SessionUtil.getEmailFromSession(request);
        if (email == null || email.isEmpty()) {
            return new ApiResponse<>("Not logged in", HttpStatus.UNAUTHORIZED);
        }

        // ログイン認証
        UserEntity user = userDao.selectByEmail(email);
        if (user == null) {
            return new ApiResponse<>("Not logged in", HttpStatus.UNAUTHORIZED);
        }
        // 企業アカウントでなければ403を返す
        if (!Objects.equals(user.getUserType(), UserTypeEnum.CL.name())) {
            return new ApiResponse<>("No permission", HttpStatus.FORBIDDEN);
        }

        // 求人一覧を取得
        List<JobEntity> jobs = getJobsDao.selectNotArchivedJobsByCompanyId(user.getCompanyId());

        return new ApiResponse<>(toResponse(jobs, page), HttpStatus.OK);
    }

    private GetJobsResponse toResponse(List<JobEntity> jobs, int page) {

        GetJobsResponse response = new GetJobsResponse();
        for (int i = 0; i < jobs.size(); i++) {

            if (i < page * JOB_LIST_PAGE_SIZE) {
                continue;
            }

            if (response.getJobs().size() >= JOB_LIST_PAGE_SIZE) {
                response.setHas_next_page(true);
                break;
            }

            GetJobsResponse.Job job = new GetJobsResponse.Job();
            job.setId(jobs.get(i).getId());
            job.setTitle(jobs.get(i).getTitle());
            job.setDescription(jobs.get(i).getDescription());
            job.setSalary(jobs.get(i).getSalary());
            job.setTags(jobs.get(i).getTags());
            job.setIs_active(jobs.get(i).getIs_active());
            job.setCreate_user_id(jobs.get(i).getCreate_user_id());
            job.setCreated_at(jobs.get(i).getCreated_at());
            job.setUpdated_at(jobs.get(i).getUpdated_at());
            response.getJobs().add(job);

        }
        return response;
    }

}
