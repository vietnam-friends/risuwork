package jp.co.recruit.isucon2024.cs.api.dao;

import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.List;

@Repository
@RequiredArgsConstructor
public class UpdateJobDao {

    private final JdbcTemplate jdbcTemplate;

    private final String updateJobBaseQuery =
            "UPDATE job SET ";

    public int updateJob(
            String title,
            String description,
            Integer salary,
            String tags,
            Boolean isActive,
            int jobId) {

        String query = updateJobBaseQuery;
        List<Object> params = new ArrayList<>();

        if (title != null && !title.isEmpty()) {
            query += " title = ?,";
            params.add(title);
        }

        if (description != null && !description.isEmpty()) {
            query += " description = ?,";
            params.add(description);
        }

        if (salary != null) {
            query += " salary = ?,";
            params.add(salary);
        }

        if (tags != null && !tags.isEmpty()) {
            query += " tags = ?,";
            params.add(tags);
        }

        if (isActive != null) {
            query += " is_active = ?,";
            params.add(isActive);
        }

        query = query.substring(0, query.length() - 1);
        query += " WHERE id = ?";
        params.add(jobId);

        return jdbcTemplate.update(query, params.toArray());

    }
}
