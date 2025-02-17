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

    public JobEntity selectJobById(int jobId) {
        return jdbcTemplate.queryForObject(selectJobByIdQuery,
                new BeanPropertyRowMapper<>(JobEntity.class), jobId);
    }

}
