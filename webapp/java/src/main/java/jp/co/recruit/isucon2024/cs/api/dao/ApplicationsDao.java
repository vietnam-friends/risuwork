package jp.co.recruit.isucon2024.cs.api.dao;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.cs.api.entity.ApplicationWithJobEntity;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.BeanPropertyRowMapper;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

import java.sql.Timestamp;
import java.util.List;

@Repository
@RequiredArgsConstructor
public class ApplicationsDao {

    private final JdbcTemplate jdbcTemplate;

    private final String selectAllApplicationsByEmailQuery =
            "SELECT a.id, a.job_id, a.user_id, a.created_at FROM application a JOIN user u ON a.user_id = u.id WHERE u.email = ? ORDER BY a.created_at DESC";

    private final String selectJobByIdQuery =
            "SELECT id, title, description, salary, tags, created_at, updated_at FROM job WHERE id = ?";

    public List<ApplicationWithJobEntity> selectAllApplicationsByEmail(String email) {

        return jdbcTemplate.queryForList(selectAllApplicationsByEmailQuery, email).stream().map(row -> {
            ApplicationWithJobEntity entity = new ApplicationWithJobEntity();
            entity.setId((Integer) row.get("id"));
            entity.setJob_id((Integer) row.get("job_id"));
            entity.setUser_id((Integer) row.get("user_id"));
            entity.setCreated_at((Timestamp) row.get("created_at"));
            return entity;
        }).toList();

    }

    public JobEntity selectJobById(int id) {
        return jdbcTemplate.queryForObject(
                selectJobByIdQuery,
                new BeanPropertyRowMapper<>(JobEntity.class),
                id);
    }

    // N+1問題解決: JOINを使用して一度に応募とジョブ情報を取得
    public List<ApplicationWithJobEntity> selectApplicationsWithJobByEmail(String email) {
        String query = "SELECT " +
                "a.id as application_id, a.job_id, a.user_id, a.created_at as application_created_at, " +
                "j.id as job_id, j.title, j.description, j.salary, j.tags, j.created_at as job_created_at, j.updated_at " +
                "FROM application a " +
                "JOIN user u ON a.user_id = u.id " +
                "JOIN job j ON a.job_id = j.id " +
                "WHERE u.email = ? " +
                "ORDER BY a.created_at DESC";
        
        return jdbcTemplate.query(query, (rs, rowNum) -> {
            ApplicationWithJobEntity entity = new ApplicationWithJobEntity();
            // Application data
            entity.setId(rs.getInt("application_id"));
            entity.setJob_id(rs.getInt("job_id"));
            entity.setUser_id(rs.getInt("user_id"));
            entity.setCreated_at(rs.getTimestamp("application_created_at"));
            
            // Job data
            JobEntity jobEntity = new JobEntity();
            jobEntity.setId(rs.getInt("job_id"));
            jobEntity.setTitle(rs.getString("title"));
            jobEntity.setDescription(rs.getString("description"));
            jobEntity.setSalary(rs.getInt("salary"));
            jobEntity.setTags(rs.getString("tags"));
            jobEntity.setCreated_at(rs.getTimestamp("job_created_at"));
            jobEntity.setUpdated_at(rs.getTimestamp("updated_at"));
            
            entity.setJob(jobEntity);
            return entity;
        }, email);
    }

}
