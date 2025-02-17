package jp.co.recruit.isucon2024.cl.api.service;

import jp.co.recruit.isucon2024.cl.api.dao.CreateCompanyDao;
import jp.co.recruit.isucon2024.cl.api.form.CreateCompanyForm;
import jp.co.recruit.isucon2024.cl.api.response.CreateCompanyResponse;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

@Service
@RequiredArgsConstructor
@Transactional(propagation = Propagation.REQUIRED, rollbackFor = Exception.class)
@Slf4j
public class CreateCompanyService {

    private final CreateCompanyDao createCompanyDao;

    public ApiResponse<CreateCompanyResponse> createCompany(CreateCompanyForm form) {

        // 企業をデータベースに登録
        int id = createCompanyDao.insertCompany(form.getName(), form.getIndustry_id());

        CreateCompanyResponse response = new CreateCompanyResponse();
        response.setMessage("Company created successfully");
        response.setId(id);

        return new ApiResponse<>(response, HttpStatus.OK);
    }

}
