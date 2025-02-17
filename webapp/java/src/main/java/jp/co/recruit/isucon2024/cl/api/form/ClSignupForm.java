package jp.co.recruit.isucon2024.cl.api.form;

import lombok.Getter;
import lombok.Setter;

@Getter
@Setter
public class ClSignupForm {

    private String email;
    private String password;
    private String name;
    private int company_id;

}
