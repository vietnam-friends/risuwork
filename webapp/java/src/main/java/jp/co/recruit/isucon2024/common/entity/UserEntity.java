package jp.co.recruit.isucon2024.common.entity;

import lombok.Getter;
import lombok.Setter;

import java.sql.Timestamp;

@Getter
@Setter
public class UserEntity {

    private Integer id;
    private String email;
    private String password;
    private String name;
    private String userType;
    private Integer companyId;
    private Timestamp created_at;
    private Timestamp updated_at;

}
