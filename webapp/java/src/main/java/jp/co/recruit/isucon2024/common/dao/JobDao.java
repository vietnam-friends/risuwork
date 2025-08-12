package jp.co.recruit.isucon2024.common.dao;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.BeanPropertyRowMapper;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

@Repository
@RequiredArgsConstructor
public class JobDao {

    private final JdbcTemplate jdbcTemplate;

    private final String selectJobByIdQuery =
            "SELECT * FROM job WHERE id = ?";
    
    private final String selectJobWithCompanyByIdQuery =
            "SELECT j.*, u.company_id FROM job j JOIN user u ON j.create_user_id = u.id WHERE j.id = ?";

    public JobEntity selectJobById(int jobId) {
        return jdbcTemplate.queryForObject(selectJobByIdQuery,
                new BeanPropertyRowMapper<>(JobEntity.class), jobId);
    }
    
    public JobEntity selectJobWithCompanyById(int jobId) {
        return jdbcTemplate.queryForObject(selectJobWithCompanyByIdQuery, (rs, rowNum) -> {
            JobEntity job = new JobEntity();
            job.setId(rs.getInt("id"));
            job.setTitle(rs.getString("title"));
            job.setDescription(rs.getString("description"));
            job.setSalary(rs.getInt("salary"));
            job.setTags(rs.getString("tags"));
            job.setIs_active(rs.getBoolean("is_active"));
            job.setIs_archived(rs.getBoolean("is_archived"));
            job.setCreate_user_id(rs.getInt("create_user_id"));
            job.setCreated_at(rs.getTimestamp("created_at"));
            job.setUpdated_at(rs.getTimestamp("updated_at"));
            job.setCompanyId(rs.getInt("company_id")); // Store company_id in job entity
            return job;
        }, jobId);
    }

}
