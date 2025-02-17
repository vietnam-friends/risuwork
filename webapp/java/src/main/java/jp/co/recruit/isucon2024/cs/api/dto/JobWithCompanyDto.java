package jp.co.recruit.isucon2024.cs.api.dto;

import jp.co.recruit.isucon2024.cs.api.entity.CompanyWithIndustryNameEntity;
import lombok.Getter;
import lombok.Setter;

import java.sql.Timestamp;

@Getter
@Setter
public class JobWithCompanyDto {

    private int id;
    private String title;
    private String description;
    private int salary;
    private String tags;
    private Timestamp created_at;
    private Timestamp updated_at;
    private CompanyWithIndustryNameEntity company;

}
