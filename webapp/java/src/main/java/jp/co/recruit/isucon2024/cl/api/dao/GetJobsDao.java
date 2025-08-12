package jp.co.recruit.isucon2024.cl.api.dao;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

import java.sql.Timestamp;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

@Repository
@RequiredArgsConstructor
public class GetJobsDao {

    private final JdbcTemplate jdbcTemplate;

    private final String selectNotArchivedJobsByCompanyIdQuery =
            "SELECT j.id, j.title, j.description, j.salary, j.tags, j.is_active, j.create_user_id, j.created_at, j.updated_at " +
                    "FROM job j JOIN user u ON j.create_user_id = u.id " +
                    "WHERE j.is_archived = false AND u.company_id = ? " +
                    "ORDER BY j.updated_at DESC, j.id";

    public List<JobEntity> selectNotArchivedJobsByCompanyId(int companyId) {
        List<Map<String, Object>> rows = jdbcTemplate.queryForList(selectNotArchivedJobsByCompanyIdQuery, companyId);
        List<JobEntity> jobs = new ArrayList<>();
        for (Map<String, Object> row : rows) {
            JobEntity job = new JobEntity();
            job.setId((Integer) row.get("id"));
            job.setTitle((String) row.get("title"));
            job.setDescription((String) row.get("description"));
            job.setSalary((Integer) row.get("salary"));
            job.setTags((String) row.get("tags"));
            job.setIs_active((Boolean) row.get("is_active"));
            job.setCreate_user_id((Integer) row.get("create_user_id"));
            job.setCreated_at((Timestamp) row.get("created_at"));
            job.setUpdated_at((Timestamp) row.get("updated_at"));
            jobs.add(job);
        }
        return jobs;
    }
}
