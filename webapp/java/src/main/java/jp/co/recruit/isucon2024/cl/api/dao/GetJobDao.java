package jp.co.recruit.isucon2024.cl.api.dao;

import jp.co.recruit.isucon2024.cl.api.response.GetJobResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.BeanPropertyRowMapper;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
@RequiredArgsConstructor
public class GetJobDao {

    private final JdbcTemplate jdbcTemplate;

    private final String selectApplicationByJobIdQuery =
            "SELECT id, job_id, user_id, created_at FROM application WHERE job_id = ? ORDER BY created_at";

    public List<GetJobResponse.Application> selectApplicationByJobId(int jobId) {
        return jdbcTemplate.query(selectApplicationByJobIdQuery,
                new BeanPropertyRowMapper<>(GetJobResponse.Application.class), jobId);
    }

}
