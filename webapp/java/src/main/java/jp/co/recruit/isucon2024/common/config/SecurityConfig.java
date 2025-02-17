package jp.co.recruit.isucon2024.common.config;

import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;

@Configuration
@EnableWebSecurity
@RequiredArgsConstructor
public class SecurityConfig {

    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder();
    }

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {

        http.authorizeHttpRequests(authorize -> authorize
                        .requestMatchers(
                                "/api/initialize",
                                "/api/cs/signup",
                                "/api/cs/login",
                                "/api/cs/logout",
                                "/api/cs/job_search",
                                "/api/cs/application",
                                "/api/cs/applications",
                                "/api/cl/company",
                                "/api/cl/signup",
                                "/api/cl/login",
                                "/api/cl/logout",
                                "/api/cl/job/**",
                                "/api/cl/jobs")
                        .permitAll()
                        .anyRequest().denyAll())
                .csrf(AbstractHttpConfigurer::disable);

        return http.build();

    }

}
