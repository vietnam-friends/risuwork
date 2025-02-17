package jp.co.recruit.isucon2024.cs.api.service;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.common.response.ApiResponse;
import jp.co.recruit.isucon2024.cs.api.dao.JobSearchDao;
import jp.co.recruit.isucon2024.cs.api.dto.JobWithCompanyDto;
import jp.co.recruit.isucon2024.cs.api.entity.CompanyWithIndustryNameEntity;
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

        List<JobEntity> jobEntities = jobSearchDao.selectJobsBySearchParams(
                form.getKeyword(), form.getTag(), form.getMin_salary(), form.getMax_salary());

        List<JobWithCompanyDto> jobWithCompanyDtoList =
                jobEntitiesTojobWithCompanyDtoList(jobEntities, form.getIndustry_id());

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

}
