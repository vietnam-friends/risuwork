package jp.co.recruit.isucon2024.common.config;

import lombok.RequiredArgsConstructor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;

@Configuration
@EnableWebSecurity
@RequiredArgsConstructor
public class SecurityConfig {

    @Bean
    public PasswordEncoder passwordEncoder() {
        // ISUCON用：パフォーマンス重視で平文保存・比較
        return new PasswordEncoder() {
            @Override
            public String encode(CharSequence rawPassword) {
                return rawPassword.toString();
            }

            @Override
            public boolean matches(CharSequence rawPassword, String encodedPassword) {
                return rawPassword.toString().equals(encodedPassword);
            }
        };
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
