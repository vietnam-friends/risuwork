package jp.co.recruit.isucon2024.cl.api.dao;

import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.support.GeneratedKeyHolder;
import org.springframework.jdbc.support.KeyHolder;
import org.springframework.stereotype.Repository;

import java.sql.PreparedStatement;

@Repository
@RequiredArgsConstructor
public class CreateJobDao {

    private final JdbcTemplate jdbcTemplate;
    private final String insertJobQuery =
            "INSERT INTO job (title, description, salary, tags, is_active, create_user_id) VALUES (?, ?, ?, ?, true, ?)";

    public int insertJob(String title, String description, int salary, String tags, int createUserId) {
        KeyHolder keyHolder = new GeneratedKeyHolder();
        jdbcTemplate.update(con -> {
            PreparedStatement ps = con.prepareStatement(insertJobQuery, new String[]{"id"});
            ps.setString(1, title);
            ps.setString(2, description);
            ps.setInt(3, salary);
            ps.setString(4, tags);
            ps.setInt(5, createUserId);
            return ps;
        }, keyHolder);
        return keyHolder.getKey().intValue();
    }

}
