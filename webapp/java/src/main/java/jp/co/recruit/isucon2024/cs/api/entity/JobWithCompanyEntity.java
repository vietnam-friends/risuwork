package jp.co.recruit.isucon2024.cs.api.entity;

import lombok.Data;

import java.sql.Timestamp;

@Data
public class JobWithCompanyEntity {
    private int job_id;
    private String title;
    private String description;
    private double salary;
    private String tags;
    private Timestamp created_at;
    private Timestamp updated_at;
    private int company_id;
    private String company_name;
    private String industry;
    private String industry_id;
}