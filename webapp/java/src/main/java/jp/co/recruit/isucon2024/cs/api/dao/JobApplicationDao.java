package jp.co.recruit.isucon2024.cs.api.dao;

import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.core.PreparedStatementCreator;
import org.springframework.jdbc.support.GeneratedKeyHolder;
import org.springframework.jdbc.support.KeyHolder;
import org.springframework.stereotype.Repository;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.SQLException;

@Repository
@RequiredArgsConstructor
public class JobApplicationDao {

    private final JdbcTemplate jdbcTemplate;

    private final String selectJobByJobIdForUpdateQuery
            = "SELECT is_active = true AND is_archived = false FROM job WHERE id = ? FOR UPDATE";
    private final String selectExistApplicationByJobIdAndUserIdQuery
            = "SELECT EXISTS (SELECT 1 FROM application WHERE job_id = ? AND user_id = ?)";
    private final String insertJobApplicationQuery
            = "INSERT INTO application (job_id, user_id) VALUES (?, ?)";

    public Boolean selectJobByJobIdForUpdate(int jobId) {
        return jdbcTemplate.queryForObject(selectJobByJobIdForUpdateQuery, Boolean.class, jobId);
    }

    public Boolean selectExistApplicationByJobIdAndUserId(int jobId, int userId) {
        return jdbcTemplate.queryForObject(selectExistApplicationByJobIdAndUserIdQuery, Boolean.class, jobId, userId);
    }

    public int insertJobApplication(int jobId, int userId) {
        KeyHolder keyHolder = new GeneratedKeyHolder();
        jdbcTemplate.update(new PreparedStatementCreator() {
            @Override
            public PreparedStatement createPreparedStatement(Connection connection) throws SQLException {
                PreparedStatement ps = connection.prepareStatement(insertJobApplicationQuery, new String[]{"id"});
                ps.setInt(1, jobId);
                ps.setInt(2, userId);
                return ps;
            }
        }, keyHolder);
        return keyHolder.getKey().intValue();
    }

}
