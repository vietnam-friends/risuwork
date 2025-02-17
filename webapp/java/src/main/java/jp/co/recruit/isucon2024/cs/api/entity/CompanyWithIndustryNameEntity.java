package jp.co.recruit.isucon2024.cs.api.entity;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class CompanyWithIndustryNameEntity {

    private int id;
    private String name;
    private String industry;
    private String industry_id;

}
