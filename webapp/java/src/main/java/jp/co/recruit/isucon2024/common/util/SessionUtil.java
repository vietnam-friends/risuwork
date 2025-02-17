package jp.co.recruit.isucon2024.common.util;

import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpSession;
import lombok.extern.slf4j.Slf4j;

@Slf4j
public class SessionUtil {

    public static void setEmailToSession(HttpServletRequest request, String email) {
        HttpSession session = request.getSession();
        session.setAttribute("email", email);
    }

    public static String getEmailFromSession(HttpServletRequest request) {
        HttpSession session = request.getSession();
        if (session == null || session.getAttribute("email") == null) {
            log.error("Error read session");
            return "";
        }
        return (String) session.getAttribute("email");
    }

    public static void removeEmailFromSession(HttpServletRequest request) {
        HttpSession session = request.getSession();
        session.removeAttribute("email");
    }

}