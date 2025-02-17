package jp.co.recruit.isucon2024.common.dao;

import jp.co.recruit.isucon2024.common.entity.UserEntity;
import lombok.RequiredArgsConstructor;
import org.springframework.dao.EmptyResultDataAccessException;
import org.springframework.jdbc.core.BeanPropertyRowMapper;
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
public class UserDao {

    private final JdbcTemplate jdbcTemplate;

    private String selectByEmailAndPasswordQuery =
            "SELECT * FROM user WHERE email = ?";
    private String selectClUserByEmailQuery =
            "SELECT * FROM user WHERE email = ? AND user_type = 'CL'";
    private String insertCsUserQuery =
            "INSERT INTO user (email, password, name, user_type) VALUES (?, ?, ?, 'CS')";
    private String insertClUserQuery =
            "INSERT INTO user (email, password, name, user_type, company_id) VALUES (?, ?, ?, 'CL', ?)";
    private String selectUserCompanyIdByUserId =
            "SELECT company_id FROM user WHERE id = ?";
    private String selectUserByUserId =
            "SELECT * FROM user WHERE id = ?";

    public UserEntity selectByEmail(String email) {
        try {
            return jdbcTemplate.queryForObject(
                    selectByEmailAndPasswordQuery,
                    new BeanPropertyRowMapper<>(UserEntity.class),
                    email);
        } catch (EmptyResultDataAccessException e) {
            return null;
        }
    }

    public UserEntity selectClUserByEmail(String email) {
        try {
            return jdbcTemplate.queryForObject(
                    selectClUserByEmailQuery,
                    new BeanPropertyRowMapper<>(UserEntity.class),
                    email);
        } catch (EmptyResultDataAccessException e) {
            return null;
        }
    }

    public int insertCsUser(String email, String password, String name) {
        KeyHolder keyHolder = new GeneratedKeyHolder();
        jdbcTemplate.update(new PreparedStatementCreator() {
            @Override
            public PreparedStatement createPreparedStatement(Connection connection) throws SQLException {
                PreparedStatement ps = connection.prepareStatement(insertCsUserQuery, new String[]{"id"});
                ps.setString(1, email);
                ps.setString(2, password);
                ps.setString(3, name);
                return ps;
            }
        }, keyHolder);
        return keyHolder.getKey().intValue();
    }

    public int insertClUser(String email, String password, String name, int companyId) {
        KeyHolder keyHolder = new GeneratedKeyHolder();
        jdbcTemplate.update(new PreparedStatementCreator() {
            @Override
            public PreparedStatement createPreparedStatement(Connection connection) throws SQLException {
                PreparedStatement ps = connection.prepareStatement(insertClUserQuery, new String[]{"id"});
                ps.setString(1, email);
                ps.setString(2, password);
                ps.setString(3, name);
                ps.setInt(4, companyId);
                return ps;
            }
        }, keyHolder);

        return keyHolder.getKey().intValue();
    }

    public Integer selectUserCompanyIdByUserId(int userId) {
        return jdbcTemplate.queryForObject(
                selectUserCompanyIdByUserId,
                Integer.class,
                userId);
    }

    public UserEntity selectUserByUserId(int userId) {
        return jdbcTemplate.queryForObject(
                selectUserByUserId,
                new BeanPropertyRowMapper<>(UserEntity.class),
                userId);
    }

}
