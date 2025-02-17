package jp.co.recruit.isucon2024.cl.api.response;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Getter;
import lombok.Setter;

import java.sql.Timestamp;
import java.util.List;

@Getter
@Setter
public class GetJobResponse {

    private int id;
    private String title;
    private String description;
    private int salary;
    private String tags;
    private Boolean is_active;
    private int create_user_id;
    private Timestamp created_at;
    private Timestamp updated_at;
    private List<Application> applications;

    @Getter
    @Setter
    public static class CSUser {

        private int id;
        private String email;
        private String name;

    }

    @Getter
    @Setter
    public static class Application {

        private int id;
        private int job_id;

        @JsonProperty("-")
        private int user_id;

        private Timestamp created_at;
        private CSUser applicant;

    }
}
