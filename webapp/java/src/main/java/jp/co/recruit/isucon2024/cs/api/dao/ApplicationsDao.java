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

}
