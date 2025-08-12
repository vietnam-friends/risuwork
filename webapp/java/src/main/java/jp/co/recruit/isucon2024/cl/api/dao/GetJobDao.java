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

    // N+1問題解決: 応募情報とユーザー情報をJOINで一度に取得
    public List<GetJobResponse.Application> selectApplicationsWithUserByJobId(int jobId) {
        String query = "SELECT " +
                "a.id, a.job_id, a.user_id, a.created_at, " +
                "u.email, u.name " +
                "FROM application a " +
                "JOIN user u ON a.user_id = u.id " +
                "WHERE a.job_id = ? " +
                "ORDER BY a.created_at";
        
        return jdbcTemplate.query(query, (rs, rowNum) -> {
            GetJobResponse.Application application = new GetJobResponse.Application();
            application.setId(rs.getInt("id"));
            application.setJob_id(rs.getInt("job_id"));
            application.setUser_id(rs.getInt("user_id"));
            application.setCreated_at(rs.getTimestamp("created_at"));
            
            // ユーザー情報も直接設定
            GetJobResponse.CSUser user = new GetJobResponse.CSUser();
            user.setId(rs.getInt("user_id"));
            user.setEmail(rs.getString("email"));
            user.setName(rs.getString("name"));
            application.setApplicant(user);
            
            return application;
        }, jobId);
    }

}
