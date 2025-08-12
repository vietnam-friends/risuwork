package jp.co.recruit.isucon2024.cs.api.service;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.cs.api.dao.JobSearchDao;
import jp.co.recruit.isucon2024.cs.api.dto.JobWithCompanyDto;
import jp.co.recruit.isucon2024.cs.api.entity.CompanyWithIndustryNameEntity;
import jp.co.recruit.isucon2024.cs.api.entity.JobWithCompanyEntity;
import jp.co.recruit.isucon2024.cs.api.form.JobSearchForm;
import jp.co.recruit.isucon2024.cs.api.response.JobSearchResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

@Service
@RequiredArgsConstructor
public class JobSearchService {

    private final JobSearchDao jobSearchDao;
    private final int JOB_SEARCH_PAGE_SIZE = 50;

    public ApiResponse<JobSearchResponse> jobSearch(JobSearchForm form) {

        // N+1問題を解決: JOINを使って一度のクエリで全データを取得
        List<JobWithCompanyEntity> jobWithCompanyEntities = jobSearchDao.selectJobsWithCompanyBySearchParams(
                form.getKeyword(), form.getTag(), form.getMin_salary(), form.getMax_salary(), form.getIndustry_id());

        List<JobWithCompanyDto> jobWithCompanyDtoList = 
                jobWithCompanyEntitiesToDtoList(jobWithCompanyEntities);

        return new ApiResponse<>(toJobSearchResponse(jobWithCompanyDtoList, form.getPage()), HttpStatus.OK);
    }

    private List<JobWithCompanyDto> jobEntitiesTojobWithCompanyDtoList(List<JobEntity> jobEntities, String industryId) {
        List<JobWithCompanyDto> jobWithCompanyDtoList = new ArrayList<>();
        for (JobEntity jobEntity : jobEntities) {
            CompanyWithIndustryNameEntity entity
                    = jobSearchDao.selectCompanyWithIndustryNameByJobId(jobEntity.getId());
            if (!industryId.isEmpty() && !entity.getIndustry_id().equals(industryId)) {
                continue;
            }
            jobWithCompanyDtoList.add(toJobWithCompanyDto(jobEntity, entity));
        }
        return jobWithCompanyDtoList;
    }

    // N+1問題解決済みの新しいメソッド
    private List<JobWithCompanyDto> jobWithCompanyEntitiesToDtoList(List<JobWithCompanyEntity> entities) {
        List<JobWithCompanyDto> jobWithCompanyDtoList = new ArrayList<>();
        for (JobWithCompanyEntity entity : entities) {
            jobWithCompanyDtoList.add(toJobWithCompanyDtoFromEntity(entity));
        }
        return jobWithCompanyDtoList;
    }

    private JobSearchResponse toJobSearchResponse(List<JobWithCompanyDto> jobWithCompanyDtoList, int page) {
        JobSearchResponse response = new JobSearchResponse();
        for (int i = 0; i < jobWithCompanyDtoList.size(); i++) {

            if (i < page * JOB_SEARCH_PAGE_SIZE) {
                continue;
            }

            if (!response.getJobs().isEmpty() &&
                    response.getJobs().size() >= JOB_SEARCH_PAGE_SIZE) {
                response.setHas_next_page(true);
                break;
            }

            response.getJobs().add(jobWithCompanyDtoList.get(i));
        }

        return response;
    }

    private JobWithCompanyDto toJobWithCompanyDto(
            JobEntity jobEntity, CompanyWithIndustryNameEntity entity) {
        JobWithCompanyDto dto = new JobWithCompanyDto();
        dto.setId(jobEntity.getId());
        dto.setTitle(jobEntity.getTitle());
        dto.setDescription(jobEntity.getDescription());
        dto.setSalary(jobEntity.getSalary());
        dto.setTags(jobEntity.getTags());
        dto.setCreated_at(jobEntity.getCreated_at());
        dto.setUpdated_at(jobEntity.getUpdated_at());
        dto.setCompany(entity);
        return dto;
    }

    // N+1問題解決済みの新しいメソッド
    private JobWithCompanyDto toJobWithCompanyDtoFromEntity(JobWithCompanyEntity entity) {
        JobWithCompanyDto dto = new JobWithCompanyDto();
        dto.setId(entity.getJob_id());
        dto.setTitle(entity.getTitle());
        dto.setDescription(entity.getDescription());
        dto.setSalary((int)entity.getSalary());
        dto.setTags(entity.getTags());
        dto.setCreated_at(entity.getCreated_at());
        dto.setUpdated_at(entity.getUpdated_at());
        
        // CompanyWithIndustryNameEntityを作成
        CompanyWithIndustryNameEntity companyEntity = new CompanyWithIndustryNameEntity();
        companyEntity.setId(entity.getCompany_id());
        companyEntity.setName(entity.getCompany_name());
        companyEntity.setIndustry(entity.getIndustry());
        companyEntity.setIndustry_id(entity.getIndustry_id());
        
        dto.setCompany(companyEntity);
        return dto;
    }

}
