package jp.co.recruit.isucon2024.cl.api.service;

import jakarta.servlet.http.HttpServletRequest;
import jp.co.recruit.isucon2024.cl.api.dao.CreateJobDao;
import jp.co.recruit.isucon2024.cl.api.form.CreateJobForm;
import jp.co.recruit.isucon2024.cl.api.response.CreateJobResponse;
import jp.co.recruit.isucon2024.common.dao.UserDao;
import jp.co.recruit.isucon2024.common.entity.UserEntity;
import jp.co.recruit.isucon2024.common.enums.UserTypeEnum;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.common.util.SessionUtil;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

import java.util.Objects;

@Service
@Transactional(propagation = Propagation.REQUIRED, rollbackFor = Exception.class)
@RequiredArgsConstructor
@Slf4j
public class CreateJobService {

    private final UserDao userDao;
    private final CreateJobDao createJobDao;

    public ApiResponse<CreateJobResponse> createJob(HttpServletRequest request, CreateJobForm form) {

        // ユーザーをDBから取得
        UserEntity user = userDao.selectByEmail(SessionUtil.getEmailFromSession(request));
        if (user == null) {
            log.error("Session user not found");
            return new ApiResponse<>("Error creating job", HttpStatus.UNAUTHORIZED);
        }
        // 企業アカウントでなければ403を返す
        if (!Objects.equals(user.getUserType(), UserTypeEnum.CL.name())) {
            return new ApiResponse<>("No permission", HttpStatus.FORBIDDEN);
        }

        // 求人をデータベースに登録
        int jobId = createJobDao.insertJob(
                form.getTitle(), form.getDescription(), form.getSalary(), form.getTags(), user.getId());

        CreateJobResponse response = new CreateJobResponse();
        response.setId(jobId);
        response.setMessage("Job created successfully");

        return new ApiResponse<>(response, HttpStatus.OK);
    }

}
