package jp.co.recruit.isucon2024.cs.api.response;

import jp.co.recruit.isucon2024.cs.api.entity.ApplicationWithJobEntity;
import lombok.Getter;
import lombok.Setter;

import java.util.ArrayList;
import java.util.List;

@Getter
@Setter
public class ApplicationsResponse {

    private List<ApplicationWithJobEntity> applications = new ArrayList<>();
    private int page;
    private boolean has_next_page;

}
