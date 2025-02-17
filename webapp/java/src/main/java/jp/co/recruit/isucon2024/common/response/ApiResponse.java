package jp.co.recruit.isucon2024.common.response;

import org.springframework.http.HttpStatusCode;
import org.springframework.http.ResponseEntity;

public class ApiResponse<T> extends ResponseEntity<Object> {

    // 4XX 5XX ERROR レスポンス
    public ApiResponse(String error, HttpStatusCode status) {
        super(error, status);
    }

    // 200 OK レスポンス
    public ApiResponse(T data, HttpStatusCode status) {
        super(data, status);
    }

}
