package jp.co.recruit.isucon2024.cs.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import jp.co.recruit.isucon2024.cs.api.dao.ApplicationsDao;
import jp.co.recruit.isucon2024.cs.api.entity.ApplicationWithJobEntity;
import jp.co.recruit.isucon2024.cs.api.response.ApplicationsResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class ApplicationsService {

    private final ApplicationsDao applicationsDao;

    private final int APPLICATION_LIST_PAGE_SIZE = 20;

    public ApiResponse<ApplicationsResponse> applications(HttpServletRequest request, int page) {

        String email = SessionUtil.getEmailFromSession(request);

        ApplicationsResponse response = new ApplicationsResponse();
        if (email == null || email.isEmpty()) {
            return new ApiResponse<>(response, HttpStatus.UNAUTHORIZED);
        }

        // N+1問題を解決: JOINを使って一度のクエリで応募と求人情報を取得
        List<ApplicationWithJobEntity> applicationWithJobEntityList
                = applicationsDao.selectApplicationsWithJobByEmail(email);

        for (int i = 0; i < applicationWithJobEntityList.size(); i++) {

            if (i < page * APPLICATION_LIST_PAGE_SIZE) {
                continue;
            }

            if (response.getApplications().size() >= APPLICATION_LIST_PAGE_SIZE) {
                response.setHas_next_page(true);
                break;
            }

            response.getApplications().add(applicationWithJobEntityList.get(i));

        }

        return new ApiResponse<>(response, HttpStatus.OK);
    }

}
