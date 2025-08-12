package jp.co.recruit.isucon2024.cs.api.dao;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.cs.api.entity.CompanyWithIndustryNameEntity;
import jp.co.recruit.isucon2024.cs.api.entity.JobWithCompanyEntity;
import lombok.RequiredArgsConstructor;
import org.springframework.jdbc.core.BeanPropertyRowMapper;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

@Repository
@RequiredArgsConstructor
public class JobSearchDao {

    private final JdbcTemplate jdbcTemplate;

    private final String selectJobsBySearchParamsBaseQuery =
            "SELECT id, title, description, salary, tags, created_at, updated_at " +
                    "FROM job WHERE is_active = true AND is_archived = false";

    private final String selectCompanyWithIndustryNameByJobIdBaseQuery =
            "SELECT company.id, company.name, industry_category.name as industry, company.industry_id " +
                    "FROM company JOIN industry_category " +
                    "ON company.industry_id = industry_category.id " +
                    "WHERE company.id = " +
                    "(SELECT company_id FROM user WHERE id = (SELECT create_user_id FROM job WHERE id = ?))";

    public List<JobEntity> selectJobsBySearchParams(
            String keyword,
            String tag,
            Integer minSalary,
            Integer maxSalary) {

        String query = selectJobsBySearchParamsBaseQuery;
        List<Object> params = new ArrayList<>();

        // フリーワード検索
        if (keyword != null && !keyword.isEmpty()) {
            query += " AND (title LIKE ? OR description LIKE ?)";
            String keywordParam = "%" + keyword + "%";
            params.add(keywordParam);
            params.add(keywordParam);
        }

        // 給与範囲検索
        if (minSalary != null && minSalary > 0) {
            query += " AND salary >= ?";
            params.add(minSalary);
        }
        if (maxSalary != null && maxSalary > 0) {
            query += " AND salary <= ?";
            params.add(maxSalary);
        }

        // タグ検索
        // タグはカンマ区切りで格納されているため
        // - jobにタグが複数ある場合
        //   - 検索対象のタグが最初にある場合
        //   - 途中にある場合
        //   - 最後にある場合
        // - jobにタグが1つしかない場合
        //   - 検索対象のタグと完全一致
        // という4パターンを考慮する必要がある
        if (tag != null && !tag.isEmpty()) {
            query += " AND (tags LIKE ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?)";
            String tagParam1 = tag + ",%";
            String tagParam2 = "%," + tag + ",%";
            String tagParam3 = "%," + tag;
            String tagParam4 = tag;
            Collections.addAll(params, tagParam1, tagParam2, tagParam3, tagParam4);
        }

        // ソート順指定
        query = query + " ORDER BY updated_at DESC, id desc";

        // クエリを実行
        return jdbcTemplate
                .query(query, params.toArray(), new BeanPropertyRowMapper<>(JobEntity.class));
    }

    public CompanyWithIndustryNameEntity selectCompanyWithIndustryNameByJobId(int jobId) {
        return jdbcTemplate.queryForObject(
                selectCompanyWithIndustryNameByJobIdBaseQuery,
                new BeanPropertyRowMapper<>(CompanyWithIndustryNameEntity.class),
                jobId);
    }

    // N+1問題解決: JOINを使用して一度に全データを取得
    public List<JobWithCompanyEntity> selectJobsWithCompanyBySearchParams(
            String keyword,
            String tag,
            Integer minSalary,
            Integer maxSalary,
            String industryId) {
        
        String query = "SELECT " +
                "j.id as job_id, j.title, j.description, j.salary, j.tags, j.created_at, j.updated_at, " +
                "c.id as company_id, c.name as company_name, ic.name as industry, c.industry_id " +
                "FROM job j " +
                "JOIN user u ON j.create_user_id = u.id " +
                "JOIN company c ON u.company_id = c.id " +
                "JOIN industry_category ic ON c.industry_id = ic.id " +
                "WHERE j.is_active = true AND j.is_archived = false";
        
        List<Object> params = new ArrayList<>();
        
        // フリーワード検索
        if (keyword != null && !keyword.isEmpty()) {
            query += " AND (j.title LIKE ? OR j.description LIKE ?)";
            String keywordParam = "%" + keyword + "%";
            params.add(keywordParam);
            params.add(keywordParam);
        }
        
        // 給与範囲検索
        if (minSalary != null && minSalary > 0) {
            query += " AND j.salary >= ?";
            params.add(minSalary);
        }
        if (maxSalary != null && maxSalary > 0) {
            query += " AND j.salary <= ?";
            params.add(maxSalary);
        }
        
        // タグ検索
        if (tag != null && !tag.isEmpty()) {
            query += " AND (j.tags LIKE ? OR j.tags LIKE ? OR j.tags LIKE ? OR j.tags LIKE ?)";
            String tagParam1 = tag + ",%";
            String tagParam2 = "%," + tag + ",%";
            String tagParam3 = "%," + tag;
            String tagParam4 = tag;
            Collections.addAll(params, tagParam1, tagParam2, tagParam3, tagParam4);
        }
        
        // 業界検索
        if (industryId != null && !industryId.isEmpty()) {
            query += " AND c.industry_id = ?";
            params.add(industryId);
        }
        
        // ソート順指定
        query += " ORDER BY j.updated_at DESC, j.id DESC";
        
        // クエリを実行
        return jdbcTemplate.query(query, params.toArray(), (rs, rowNum) -> {
            JobWithCompanyEntity entity = new JobWithCompanyEntity();
            entity.setJob_id(rs.getInt("job_id"));
            entity.setTitle(rs.getString("title"));
            entity.setDescription(rs.getString("description"));
            entity.setSalary(rs.getDouble("salary"));
            entity.setTags(rs.getString("tags"));
            entity.setCreated_at(rs.getTimestamp("created_at"));
            entity.setUpdated_at(rs.getTimestamp("updated_at"));
            entity.setCompany_id(rs.getInt("company_id"));
            entity.setCompany_name(rs.getString("company_name"));
            entity.setIndustry(rs.getString("industry"));
            entity.setIndustry_id(rs.getString("industry_id"));
            return entity;
        });
    }

}
