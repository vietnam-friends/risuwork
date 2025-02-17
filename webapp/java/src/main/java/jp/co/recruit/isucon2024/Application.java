package jp.co.recruit.isucon2024;

import com.amazonaws.xray.spring.aop.XRayEnabled;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.autoconfigure.security.servlet.UserDetailsServiceAutoConfiguration;

@SpringBootApplication(exclude = {UserDetailsServiceAutoConfiguration.class})
@XRayEnabled
public class Application {

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }

}
