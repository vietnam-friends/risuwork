package jp.co.recruit.isucon2024.cs.api.dao;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import jp.co.recruit.isucon2024.cs.api.entity.CompanyWithIndustryNameEntity;
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

}
