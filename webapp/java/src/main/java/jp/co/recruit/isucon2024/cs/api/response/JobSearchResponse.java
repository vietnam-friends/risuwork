package jp.co.recruit.isucon2024.cs.api.response;

import jp.co.recruit.isucon2024.cs.api.dto.JobWithCompanyDto;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;
import java.util.List;

@Getter
@Setter
public class JobSearchResponse {

    private List<JobWithCompanyDto> jobs = new ArrayList<>();
    private int page;
    private boolean has_next_page;

}
