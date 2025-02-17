package jp.co.recruit.isucon2024.cl.api.form;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class UpdateJobForm {

    private String title;
    private String description;
    private Integer salary;
    private String tags;
    private Boolean is_active;

}
