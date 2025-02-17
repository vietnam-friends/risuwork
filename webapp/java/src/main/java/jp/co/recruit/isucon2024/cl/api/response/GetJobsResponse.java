package jp.co.recruit.isucon2024.cl.api.response;

import lombok.Getter;
import lombok.Setter;

import java.sql.Timestamp;
import java.util.ArrayList;
import java.util.List;

@Getter
@Setter
public class GetJobsResponse {

    List<Job> jobs = new ArrayList<>();
    int page;
    boolean has_next_page;

    @Getter
    @Setter
    public static class Job {
        private int id;
        private String title;
        private String description;
        private int salary;
        private String tags;
        private Boolean is_active;
        private int create_user_id;
        private Timestamp created_at;
        private Timestamp updated_at;
    }

}
