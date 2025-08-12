package jp.co.recruit.isucon2024.common.entity;

import lombok.Getter;
import lombok.Setter;

import java.sql.Timestamp;

@Getter
@Setter
public class JobEntity {

    private int id;
    private String title;
    private String description;
    private int salary;
    private String tags;
    private Boolean is_active;
    private Boolean is_archived;
    private int create_user_id;
    private Timestamp created_at;
    private Timestamp updated_at;
    private Integer companyId; // For JOIN queries

}
