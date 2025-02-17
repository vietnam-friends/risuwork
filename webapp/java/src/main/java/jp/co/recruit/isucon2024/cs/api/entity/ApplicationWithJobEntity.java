package jp.co.recruit.isucon2024.cs.api.entity;

import jp.co.recruit.isucon2024.common.entity.JobEntity;
import lombok.Getter;
import lombok.Setter;

import java.sql.Timestamp;

@Getter
@Setter
public class ApplicationWithJobEntity {

    private int id;
    private int job_id;
    private int user_id;
    private Timestamp created_at;
    private JobEntity job;
    
}
