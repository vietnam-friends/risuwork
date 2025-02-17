package jp.co.recruit.isucon2024.cs.api.form;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class JobSearchForm {

    private String keyword;
    private Integer min_salary;
    private Integer max_salary;
    private String tag;
    private String industry_id;
    private int page;
    
}
