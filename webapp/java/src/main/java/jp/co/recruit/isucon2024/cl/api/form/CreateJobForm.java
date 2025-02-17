package jp.co.recruit.isucon2024.cl.api.form;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class CreateJobForm {

    private String title;
    private String description;
    private int salary;
    private String tags;

}
