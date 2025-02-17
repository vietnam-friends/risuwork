package jp.co.recruit.isucon2024.cl.api.dao;

import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

@Repository
@RequiredArgsConstructor
public class ArchiveJobDao {

    private final JdbcTemplate jdbcTemplate;

    private final String archiveJobQuery = "UPDATE job SET is_archived = true WHERE id = ?";

    public void archiveJob(int jobId) {
        jdbcTemplate.update(archiveJobQuery, jobId);
    }

}
