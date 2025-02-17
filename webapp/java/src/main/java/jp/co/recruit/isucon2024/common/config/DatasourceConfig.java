package jp.co.recruit.isucon2024.common.config;

import com.amazonaws.xray.sql.TracingDataSource;
import lombok.Getter;
import lombok.Setter;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.boot.jdbc.DataSourceBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import javax.sql.DataSource;

@Configuration
@ConfigurationProperties("spring.datasource")
@Getter
@Setter
public class DatasourceConfig {

    private String url;

    private String username;

    private String password;

    @Bean
    public DataSource dataSource() {
        final DataSource dataSource = DataSourceBuilder.create()
                .url(url)
                .username(username)
                .password(password)
                .build();
        return new TracingDataSource(dataSource);
    }
}
