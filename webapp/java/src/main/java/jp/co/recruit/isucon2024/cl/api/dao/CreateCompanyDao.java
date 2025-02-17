package jp.co.recruit.isucon2024.cl.api.dao;

import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.jdbc.support.GeneratedKeyHolder;
import org.springframework.jdbc.support.KeyHolder;
import org.springframework.stereotype.Repository;

import java.sql.PreparedStatement;

@Repository
@RequiredArgsConstructor
public class CreateCompanyDao {

    private final JdbcTemplate jdbcTemplate;

    private final String insertCompanyQuery = "INSERT INTO company (name, industry_id) VALUES (?, ?)";

    public int insertCompany(String name, String industry_id) {
        KeyHolder keyHolder = new GeneratedKeyHolder();
        jdbcTemplate.update(connection -> {
            PreparedStatement ps = connection.prepareStatement(insertCompanyQuery, new String[]{"id"});
            ps.setString(1, name);
            ps.setString(2, industry_id);
            return ps;
        }, keyHolder);
        return keyHolder.getKey().intValue();
    }

}
